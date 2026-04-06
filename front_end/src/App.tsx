import React, { useState, useCallback } from 'react';
import ReactFlow, { 
  Background, 
  ConnectionLineType,
  ConnectionMode,
  Controls, 
  MarkerType,
  Panel,
  useReactFlow,
  BackgroundVariant
} from 'reactflow';
import 'reactflow/dist/style.css';

import { useGraphStore } from './store/useGraphStore';
import { InputNode } from './components/nodes/InputNode';
import { LLMNode } from './components/nodes/LLMNode';
import { ToolboxNode } from './components/nodes/ToolboxNode';
import { OutputNode } from './components/nodes/OutputNode';
import { IfElseNode } from './components/nodes/IfElseNode';
import { WhileNode } from './components/nodes/WhileNode';
import { WorkflowEdge } from './components/edges/WorkflowEdge';
import { SidePanel } from './components/SidePanel';
import { LogPanel } from './components/LogPanel';
import { Palette } from './components/Palette';
import { saveGraph, executeGraph, loadGraph, loadGraphs } from './api/client';
import { extractDisplayContent, extractFinalResponse, getUsageSnapshot, isFinalAgentResponse } from './utils/execution';

const nodeTypes = {
  input_node: InputNode,
  llm_node: LLMNode,
  toolbox: ToolboxNode,
  output_node: OutputNode,
  if_else_node: IfElseNode,
  while_node: WhileNode,
};

const edgeTypes = {
  workflow: WorkflowEdge,
};

interface RuntimeStats {
  steps: number;
  elapsedMs: number;
  totalTokens: number;
  currentNode: string | null;
  exitReason: string | null;
  nodeVisits: Record<string, number>;
}

const App: React.FC = () => {
  const { 
    nodes, 
    edges, 
    onNodesChange, 
    onEdgesChange, 
    onConnect,
    exportGraph,
    importGraph,
    validateGraph,
    addNode,
    name
  } = useGraphStore();

  const { screenToFlowPosition } = useReactFlow();
  const [logs, setLogs] = useState<any[]>([]);
  const [isLogOpen, setIsLogOpen] = useState(false);
  const [finalResponse, setFinalResponse] = useState<string | null>(null);
  const [executionInput, setExecutionInput] = useState('Hello, who are you?');
  const [maxStepsInput] = useState('128');
  const [maxDurationSecondsInput] = useState('60');
  const [maxTokensInput] = useState('');
  const [runtimeStats, setRuntimeStats] = useState<RuntimeStats | null>(null);

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  }, []);

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault();

      const type = event.dataTransfer.getData('application/reactflow');
      if (typeof type === 'undefined' || !type) return;

      const position = screenToFlowPosition({
        x: event.clientX,
        y: event.clientY,
      });

      addNode(type, position);
    },
    [screenToFlowPosition, addNode]
  );

  const handleSave = async () => {
    setIsLogOpen(true);
    const graph = exportGraph();
    addLog('info', `Saving graph '${graph.name}'...`);
    try {
        await saveGraph(graph);
        addLog('info', `Graph '${graph.name}' saved successfully.`);
    } catch (e) {
        addLog('error', `Failed to save graph: ${e}`);
    }
  };

  const handleLoad = async () => {
    setIsLogOpen(true);
    addLog('info', 'Loading saved graphs...');
    try {
        const names = await loadGraphs();
        if (!names || names.length === 0) {
            addLog('info', 'No saved graphs found.');
            alert("No saved graphs found.");
            return;
        }
        const selected = window.prompt(`Enter graph name to load:\nAvailable: ${names.join(', ')}`);
        if (selected) {
            addLog('info', `Loading graph '${selected}'...`);
            const data = await loadGraph(selected);
            importGraph(data);
            addLog('info', `Graph '${selected}' loaded.`);
        } else {
            addLog('info', 'Load cancelled.');
        }
    } catch (e) {
        addLog('error', `Failed to load graphs: ${e}`);
    }
  };

  const handleValidate = () => {
    const isValid = validateGraph();
    setIsLogOpen(true);
    if (isValid) {
        addLog('info', 'Graph contract validation: SUCCESS');
    } else {
        const errors = useGraphStore.getState().validationErrors;
        errors.forEach(err => addLog('error', `Validation Error: ${err}`));
    }
  };

  const handleDeploy = async () => {
    const userInput = executionInput;
    if (!userInput.trim()) {
      setIsLogOpen(true);
      addLog('error', 'Execution input is empty.');
      return;
    }

    const budget = {
      max_steps: parseOptionalInt(maxStepsInput),
      max_duration_ms: secondsToMilliseconds(maxDurationSecondsInput),
      max_total_tokens: parseOptionalInt(maxTokensInput),
    };
    const startedAt = Date.now();

    setFinalResponse(null);
    setIsLogOpen(true);
    setRuntimeStats({
      steps: 0,
      elapsedMs: 0,
      totalTokens: 0,
      currentNode: null,
      exitReason: null,
      nodeVisits: {},
    });
    addLog('info', 'Starting execution...');
    addLog('info', `Input: ${userInput}`);
    addLog('info', `Budget: steps=${budget.max_steps ?? 'unbounded'}, duration=${budget.max_duration_ms ? `${Math.round(budget.max_duration_ms / 1000)}s` : 'unbounded'}, tokens=${budget.max_total_tokens ?? 'unbounded'}`);
    
    const graph = exportGraph();
    // Output nodes declare which state keys represent user-facing results.
    // Those keys are later used to prefer structured workflow outputs over raw
    // transport events when populating the log and final-response panels.
    const outputKeys = graph.nodes
      .filter((node: any) => node.type === 'output_node')
      .map((node: any) => node.config.output_key)
      .filter(Boolean);

    try {
        await executeGraph(graph, userInput, budget, (event) => {
            console.log("[DEBUG] IDE Received Event:", JSON.stringify(event, null, 2));

            if (event.type === 'diagnostic') {
                handleDiagnosticEvent(event.content, startedAt);
                return;
            }

            const logType = event.type === 'agent_event' ? (event.author || 'agent') : event.type;
            // Log rendering is intentionally more forgiving than the final
            // response panel so intermediate state deltas and tool events remain
            // visible during execution.
            const responseText = extractFinalResponse(event, outputKeys);
            const logContent = responseText ?? extractDisplayContent(event, outputKeys);

            if (logContent != null) {
              addLog(logType, logContent);
            }
            updateUsageStats(event.content, startedAt);

            if (isFinalAgentResponse(event) && responseText) {
                setFinalResponse(responseText);
            }
        });
    } catch (e) {
        addLog('error', `Execution failed: ${e}`);
    }
  };

  const addLog = (type: string, content: any) => {
    setLogs(prev => [...prev, { type, content, timestamp: new Date() }]);
  };

  const updateUsageStats = (content: any, startedAt: number) => {
    const usage = getUsageSnapshot(content);
    if (!usage) {
      return;
    }

    setRuntimeStats((prev) => {
      if (!prev) {
        return prev;
      }

      return {
        ...prev,
        elapsedMs: Date.now() - startedAt,
        totalTokens: prev.totalTokens + usage.totalTokens,
      };
    });
  };

  const handleDiagnosticEvent = (diagnostic: any, startedAt: number) => {
    if (!diagnostic || typeof diagnostic !== 'object') {
      return;
    }

    switch (diagnostic.kind) {
      case 'node_enter':
        setRuntimeStats((prev) => {
          const next = prev ?? {
            steps: 0,
            elapsedMs: 0,
            totalTokens: 0,
            currentNode: null,
            exitReason: null,
            nodeVisits: {},
          };
          return {
            ...next,
            steps: Number(diagnostic.step ?? next.steps),
            elapsedMs: Date.now() - startedAt,
            currentNode: diagnostic.agent_name || diagnostic.node_id || next.currentNode,
            nodeVisits: {
              ...next.nodeVisits,
              [(diagnostic.agent_name || diagnostic.node_id || 'unknown')]: Number(diagnostic.visit_count ?? 0),
            },
          };
        });
        addLog('step', `Step ${diagnostic.step}: entering ${diagnostic.agent_name || diagnostic.node_id} (visit ${diagnostic.visit_count})`);
        break;
      case 'transition':
        addLog('route', `${diagnostic.from_agent_name} -> ${diagnostic.to_agent_name} via ${diagnostic.reason}`);
        break;
      case 'summary':
        setRuntimeStats((prev) => ({
          steps: Number(diagnostic.steps ?? prev?.steps ?? 0),
          elapsedMs: Number(diagnostic.elapsed_ms ?? prev?.elapsedMs ?? 0),
          totalTokens: Number(diagnostic.total_tokens ?? prev?.totalTokens ?? 0),
          currentNode: (diagnostic.current_node as string) || prev?.currentNode || null,
          exitReason: (diagnostic.exit_reason as string) || prev?.exitReason || null,
          nodeVisits: (diagnostic.node_visits as Record<string, number>) || prev?.nodeVisits || {},
        }));
        addLog('info', `Execution summary: ${diagnostic.exit_reason}, ${diagnostic.steps} steps, ${diagnostic.total_tokens} tokens, ${(Number(diagnostic.elapsed_ms ?? 0) / 1000).toFixed(1)}s`);
        break;
      default:
        break;
    }
  };

  return (
    <div className="flex flex-col h-screen bg-gray-950 text-gray-100 overflow-hidden">
      <header className="h-14 border-b border-gray-800 flex items-center justify-between px-6 bg-gray-900 shrink-0">
        <div className="flex items-center gap-2">
          <div className="w-6 h-6 bg-blue-600 rounded flex items-center justify-center font-bold text-xs shadow-lg shadow-blue-900/20 text-white">V</div>
          <span className="font-bold tracking-tight text-white">Visual Agent <span className="text-gray-600 font-normal ml-1">IDE</span></span>
        </div>
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2 px-3 py-1.5 rounded bg-gray-950 border border-gray-800 w-56 xl:w-80">
            <span className="text-[10px] font-bold uppercase tracking-widest text-gray-500 shrink-0">Input</span>
            <input
              type="text"
              value={executionInput}
              onChange={(e) => setExecutionInput(e.target.value)}
              placeholder="Enter agent input"
              className="w-full bg-transparent text-sm text-gray-100 focus:outline-none placeholder:text-gray-600"
            />
          </div>
          <button 
            className="px-3 py-1.5 rounded bg-gray-800 border border-gray-700 text-[10px] font-bold uppercase tracking-widest text-gray-400 hover:bg-gray-700 transition-all active:scale-95"
            onClick={handleLoad}
          >
            Load
          </button>
          <button 
            className="px-3 py-1.5 rounded bg-gray-800 border border-gray-700 text-[10px] font-bold uppercase tracking-widest text-gray-400 hover:bg-gray-700 transition-all active:scale-95"
            onClick={handleValidate}
          >
            Validate
          </button>
          <button 
            className="px-3 py-1.5 rounded bg-gray-800 border border-gray-700 text-[10px] font-bold uppercase tracking-widest text-gray-400 hover:bg-gray-700 transition-all active:scale-95"
            onClick={handleSave}
          >
            Save
          </button>
          <button 
            className="px-4 py-1.5 rounded bg-blue-600 text-white text-[11px] font-bold uppercase tracking-widest hover:bg-blue-500 transition-all shadow-lg shadow-blue-900/30 active:scale-95"
            onClick={handleDeploy}
          >
            Export & Deploy
          </button>
        </div>
      </header>

      <main className="flex-1 flex overflow-hidden relative">
        <Palette />
        <div className="flex-1 relative h-full">
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onDrop={onDrop}
            onDragOver={onDragOver}
            nodeTypes={nodeTypes}
            edgeTypes={edgeTypes}
            connectionMode={ConnectionMode.Loose}
            connectionLineType={ConnectionLineType.SmoothStep}
            connectionRadius={40}
            fitViewOptions={{ padding: 0.4, maxZoom: 0.85 }}
            defaultEdgeOptions={{
              type: 'workflow',
              style: { stroke: '#94a3b8', strokeWidth: 2 },
              markerEnd: { type: MarkerType.ArrowClosed, color: '#94a3b8' },
            }}
            fitView
          >
            <Background color="#222" gap={20} variant={BackgroundVariant.Lines} />
            <Controls className="!bg-gray-900 !border-gray-800 shadow-2xl" />
            <Panel position="top-left" className="bg-gray-900/80 backdrop-blur-md border border-gray-800 p-2 rounded shadow-xl">
                <div className="text-[9px] font-bold text-gray-500 uppercase tracking-widest mb-1">Graph Status</div>
                <div className="flex items-center gap-2">
                    <div className="w-1.5 h-1.5 rounded-full bg-green-500 shadow-sm shadow-green-500"></div>
                    <span className="text-[10px] font-semibold text-gray-300">{name}</span>
                </div>
            </Panel>
          </ReactFlow>
          
          <LogPanel 
            logs={logs} 
            response={finalResponse}
            stats={runtimeStats}
            isOpen={isLogOpen} 
            onToggle={() => setIsLogOpen(!isLogOpen)} 
            onClear={() => {
              setLogs([]);
              setFinalResponse(null);
              setRuntimeStats(null);
            }}
          />
        </div>
        <SidePanel />
      </main>
    </div>
  );
};

export default App;

const parseOptionalInt = (value: string) => {
  const trimmed = value.trim();
  if (!trimmed) {
    return undefined;
  }
  const parsed = Number.parseInt(trimmed, 10);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : undefined;
};

const secondsToMilliseconds = (value: string) => {
  const seconds = parseOptionalInt(value);
  return seconds ? seconds * 1000 : undefined;
};
