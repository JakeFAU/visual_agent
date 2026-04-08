import React, { useCallback, useRef, useState } from 'react';
import ReactFlow, {
  Background,
  BackgroundVariant,
  ConnectionLineType,
  ConnectionMode,
  Controls,
  MarkerType,
  Panel,
  useReactFlow,
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
import { GraphLibraryModal } from './components/GraphLibraryModal';
import { executeGraph, loadGraph, loadGraphs, saveGraph } from './api/client';
import { exampleGraphs, ExampleGraph, loadExampleGraph } from './examples/library';
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

const emptyRuntimeStats = (): RuntimeStats => ({
  steps: 0,
  elapsedMs: 0,
  totalTokens: 0,
  currentNode: null,
  exitReason: null,
  nodeVisits: {},
});

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
    clearGraph,
    name,
    setName,
  } = useGraphStore();

  const { screenToFlowPosition } = useReactFlow();
  const importInputRef = useRef<HTMLInputElement | null>(null);

  const [logs, setLogs] = useState<any[]>([]);
  const [isLogOpen, setIsLogOpen] = useState(false);
  const [finalResponse, setFinalResponse] = useState<string | null>(null);
  const [executionInput, setExecutionInput] = useState('Hello, who are you?');
  const [maxStepsInput, setMaxStepsInput] = useState('128');
  const [maxDurationSecondsInput, setMaxDurationSecondsInput] = useState('60');
  const [maxTokensInput, setMaxTokensInput] = useState('');
  const [runtimeStats, setRuntimeStats] = useState<RuntimeStats | null>(null);
  const [isLibraryOpen, setIsLibraryOpen] = useState(false);
  const [savedGraphs, setSavedGraphs] = useState<string[]>([]);
  const [isLoadingSavedGraphs, setIsLoadingSavedGraphs] = useState(false);
  const [libraryError, setLibraryError] = useState<string | null>(null);
  const [activeLibraryItemKey, setActiveLibraryItemKey] = useState<string | null>(null);

  const addLog = useCallback((type: string, content: any) => {
    const timestamp = new Date();
    setLogs((prev) => {
      const last = prev[prev.length - 1];
      if (
        last &&
        shouldSuppressDuplicateAgentLog(last, type, content, timestamp)
      ) {
        return prev;
      }

      return [...prev, { type, content, timestamp }];
    });
  }, []);

  const resetRunState = useCallback(() => {
    setFinalResponse(null);
    setRuntimeStats(null);
  }, []);

  const refreshSavedGraphs = useCallback(async () => {
    setIsLoadingSavedGraphs(true);
    setLibraryError(null);

    try {
      const names = await loadGraphs();
      setSavedGraphs([...names].sort((left, right) => left.localeCompare(right)));
    } catch (error) {
      setLibraryError(getErrorMessage(error));
      setSavedGraphs([]);
    } finally {
      setIsLoadingSavedGraphs(false);
    }
  }, []);

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
    [screenToFlowPosition, addNode],
  );

  const handleNewGraph = useCallback(() => {
    clearGraph();
    resetRunState();
    setIsLogOpen(true);
    addLog('info', 'Started a new blank workflow.');
  }, [addLog, clearGraph, resetRunState]);

  const handleOpenLibrary = useCallback(() => {
    setIsLibraryOpen(true);
    void refreshSavedGraphs();
  }, [refreshSavedGraphs]);

  const handleSave = useCallback(async () => {
    setIsLogOpen(true);
    const graph = exportGraph();
    addLog('info', `Saving graph '${graph.name}'...`);

    try {
      await saveGraph(graph);
      addLog('info', `Graph '${graph.name}' saved successfully.`);
      setSavedGraphs((prev) => {
        const next = new Set(prev);
        next.add(graph.name);
        return [...next].sort((left, right) => left.localeCompare(right));
      });
    } catch (error) {
      addLog('error', `Failed to save graph: ${getErrorMessage(error)}`);
    }
  }, [addLog, exportGraph]);

  const handleLoadSavedGraph = useCallback(async (graphName: string) => {
    setActiveLibraryItemKey(`saved:${graphName}`);

    try {
      const data = await loadGraph(graphName);
      const imported = importGraph(data);
      if (!imported) {
        throw new Error(`Saved graph '${graphName}' does not match the current contract.`);
      }
      resetRunState();
      setIsLibraryOpen(false);
      setIsLogOpen(true);
      addLog('info', `Loaded saved workflow '${graphName}'.`);
    } catch (error) {
      setLibraryError(getErrorMessage(error));
      setIsLogOpen(true);
      addLog('error', `Failed to load graph '${graphName}': ${getErrorMessage(error)}`);
    } finally {
      setActiveLibraryItemKey(null);
    }
  }, [addLog, importGraph, resetRunState]);

  const handleLoadExampleGraph = useCallback(async (example: ExampleGraph) => {
    setActiveLibraryItemKey(`example:${example.id}`);

    try {
      const data = await loadExampleGraph(example);
      const imported = importGraph(data);
      if (!imported) {
        throw new Error(`Example '${example.title}' does not match the current contract.`);
      }
      resetRunState();
      setIsLibraryOpen(false);
      setIsLogOpen(true);
      addLog('info', `Loaded example '${example.title}'.`);
    } catch (error) {
      setLibraryError(getErrorMessage(error));
      setIsLogOpen(true);
      addLog('error', `Failed to load example '${example.title}': ${getErrorMessage(error)}`);
    } finally {
      setActiveLibraryItemKey(null);
    }
  }, [addLog, importGraph, resetRunState]);

  const handleImportFileSelection = useCallback(async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    event.target.value = '';

    if (!file) {
      return;
    }

    try {
      const text = await file.text();
      const data = JSON.parse(text);
      const imported = importGraph(data);
      if (!imported) {
        throw new Error(`'${file.name}' does not match the current graph contract.`);
      }
      resetRunState();
      setIsLogOpen(true);
      addLog('info', `Imported workflow from '${file.name}'.`);
    } catch (error) {
      setIsLogOpen(true);
      addLog('error', `Failed to import '${file.name}': ${getErrorMessage(error)}`);
    }
  }, [addLog, importGraph, resetRunState]);

  const handleValidate = useCallback(() => {
    const isValid = validateGraph();
    setIsLogOpen(true);

    if (isValid) {
      addLog('info', 'Graph contract validation: SUCCESS');
      return;
    }

    const errors = useGraphStore.getState().validationErrors;
    errors.forEach((error) => addLog('error', `Validation Error: ${error}`));
  }, [addLog, validateGraph]);

  const updateUsageStats = useCallback((content: any, startedAt: number) => {
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
  }, []);

  const handleDiagnosticEvent = useCallback((diagnostic: any, startedAt: number) => {
    if (!diagnostic || typeof diagnostic !== 'object') {
      return;
    }

    switch (diagnostic.kind) {
      case 'node_enter':
        setRuntimeStats((prev) => {
          const next = prev ?? emptyRuntimeStats();

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
        addLog(
          'info',
          `Execution summary: ${diagnostic.exit_reason}, ${diagnostic.steps} steps, ${diagnostic.total_tokens} tokens, ${(Number(diagnostic.elapsed_ms ?? 0) / 1000).toFixed(1)}s`,
        );
        break;
      default:
        break;
    }
  }, [addLog]);

  const handleDeploy = useCallback(async () => {
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
    setRuntimeStats(emptyRuntimeStats());
    addLog('info', 'Starting execution...');
    addLog('info', `Input: ${userInput}`);
    addLog(
      'info',
      `Budget: steps=${budget.max_steps ?? 'unbounded'}, duration=${budget.max_duration_ms ? `${Math.round(budget.max_duration_ms / 1000)}s` : 'unbounded'}, tokens=${budget.max_total_tokens ?? 'unbounded'}`,
    );

    const graph = exportGraph();
    const outputKeys = graph.nodes
      .filter((node: any) => node.type === 'output_node')
      .map((node: any) => node.config.output_key)
      .filter(Boolean);

    try {
      await executeGraph(graph, userInput, budget, (event) => {
        if (event.type === 'diagnostic') {
          handleDiagnosticEvent(event.content, startedAt);
          return;
        }

        const logType = event.type === 'agent_event' ? (event.author || 'agent') : event.type;
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
    } catch (error) {
      addLog('error', `Execution failed: ${getErrorMessage(error)}`);
    }
  }, [
    addLog,
    executionInput,
    exportGraph,
    handleDiagnosticEvent,
    maxDurationSecondsInput,
    maxStepsInput,
    maxTokensInput,
    updateUsageStats,
  ]);

  return (
    <div className="flex h-screen flex-col overflow-hidden bg-gray-950 text-gray-100">
      <header className="shrink-0 border-b border-gray-800 bg-gray-900 px-6 py-4">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div className="flex min-w-[280px] flex-1 items-center gap-4">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-blue-600 text-sm font-bold text-white shadow-lg shadow-blue-900/30">
              V
            </div>
            <div className="min-w-[220px] max-w-xl flex-1">
              <div className="text-[10px] font-bold uppercase tracking-[0.22em] text-gray-500">Workflow</div>
              <div className="mt-2 flex items-center gap-3">
                <input
                  type="text"
                  value={name}
                  onChange={(event) => setName(event.target.value)}
                  placeholder="Name this workflow"
                  className="w-full rounded-lg border border-gray-800 bg-gray-950 px-3 py-2 text-sm font-semibold text-white outline-none transition-colors focus:border-blue-500"
                />
                <div className="hidden rounded-full border border-emerald-500/20 bg-emerald-950/20 px-3 py-1 text-[10px] font-bold uppercase tracking-[0.18em] text-emerald-300 xl:block">
                  v0 local-first
                </div>
              </div>
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            <button
              type="button"
              className="rounded-lg border border-gray-800 bg-gray-950 px-3 py-2 text-[10px] font-bold uppercase tracking-[0.18em] text-gray-300 transition-colors hover:border-gray-700 hover:text-white"
              onClick={handleNewGraph}
            >
              New
            </button>
            <button
              type="button"
              className="rounded-lg border border-gray-800 bg-gray-950 px-3 py-2 text-[10px] font-bold uppercase tracking-[0.18em] text-gray-300 transition-colors hover:border-gray-700 hover:text-white"
              onClick={handleOpenLibrary}
            >
              Open
            </button>
            <button
              type="button"
              className="rounded-lg border border-gray-800 bg-gray-950 px-3 py-2 text-[10px] font-bold uppercase tracking-[0.18em] text-gray-300 transition-colors hover:border-gray-700 hover:text-white"
              onClick={() => importInputRef.current?.click()}
            >
              Import JSON
            </button>
            <button
              type="button"
              className="rounded-lg border border-gray-800 bg-gray-950 px-3 py-2 text-[10px] font-bold uppercase tracking-[0.18em] text-gray-300 transition-colors hover:border-gray-700 hover:text-white"
              onClick={handleValidate}
            >
              Validate
            </button>
            <button
              type="button"
              className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-[10px] font-bold uppercase tracking-[0.18em] text-gray-200 transition-colors hover:border-gray-600 hover:bg-gray-700"
              onClick={handleSave}
            >
              Save
            </button>
            <button
              type="button"
              className="rounded-lg bg-blue-600 px-4 py-2 text-[10px] font-bold uppercase tracking-[0.18em] text-white shadow-lg shadow-blue-900/30 transition-colors hover:bg-blue-500"
              onClick={handleDeploy}
            >
              Run Workflow
            </button>
          </div>
        </div>

        <div className="mt-4 flex flex-wrap items-end gap-3">
          <div className="min-w-[240px] flex-[2_1_360px] rounded-xl border border-gray-800 bg-gray-950 px-3 py-2">
            <div className="text-[10px] font-bold uppercase tracking-[0.18em] text-gray-500">Execution Input</div>
            <input
              type="text"
              value={executionInput}
              onChange={(event) => setExecutionInput(event.target.value)}
              placeholder="Enter agent input"
              className="mt-2 w-full bg-transparent text-sm text-gray-100 outline-none placeholder:text-gray-600"
            />
          </div>

          <div className="min-w-[280px] flex-[1.4_1_420px] rounded-xl border border-gray-800 bg-gray-950 px-3 py-2">
            <div className="text-[10px] font-bold uppercase tracking-[0.18em] text-gray-500">Run Budget</div>
            <div className="mt-2 grid gap-2 sm:grid-cols-3">
              <label className="block">
                <div className="text-[10px] uppercase tracking-[0.14em] text-gray-600">Steps</div>
                <input
                  type="number"
                  min="1"
                  value={maxStepsInput}
                  onChange={(event) => setMaxStepsInput(event.target.value)}
                  placeholder="128"
                  className="mt-1 w-full rounded-md border border-gray-800 bg-gray-900 px-2.5 py-2 text-sm text-white outline-none transition-colors focus:border-blue-500"
                />
              </label>
              <label className="block">
                <div className="text-[10px] uppercase tracking-[0.14em] text-gray-600">Seconds</div>
                <input
                  type="number"
                  min="1"
                  value={maxDurationSecondsInput}
                  onChange={(event) => setMaxDurationSecondsInput(event.target.value)}
                  placeholder="60"
                  className="mt-1 w-full rounded-md border border-gray-800 bg-gray-900 px-2.5 py-2 text-sm text-white outline-none transition-colors focus:border-blue-500"
                />
              </label>
              <label className="block">
                <div className="text-[10px] uppercase tracking-[0.14em] text-gray-600">Tokens</div>
                <input
                  type="number"
                  min="1"
                  value={maxTokensInput}
                  onChange={(event) => setMaxTokensInput(event.target.value)}
                  placeholder="Optional"
                  className="mt-1 w-full rounded-md border border-gray-800 bg-gray-900 px-2.5 py-2 text-sm text-white outline-none transition-colors focus:border-blue-500"
                />
              </label>
            </div>
          </div>

          <div className="min-w-[220px] flex-[1_1_280px] rounded-xl border border-rose-500/20 bg-rose-950/10 px-3 py-2 text-xs leading-5 text-rose-100/75">
            These are run-wide safety caps. Each <span className="font-mono text-rose-200">while_node</span> still enforces its own local{' '}
            <span className="font-mono text-rose-200">max_iterations</span>.
          </div>
        </div>

        <input
          ref={importInputRef}
          type="file"
          accept="application/json,.json"
          className="hidden"
          onChange={handleImportFileSelection}
        />
      </header>

      <main className="relative flex flex-1 overflow-hidden">
        <Palette />
        <div className="relative h-full flex-1">
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
            <Panel position="top-left" className="rounded border border-gray-800 bg-gray-900/80 p-3 shadow-xl backdrop-blur-md">
              <div className="text-[9px] font-bold uppercase tracking-widest text-gray-500">Graph Status</div>
              <div className="mt-2 flex items-center gap-2">
                <div className="h-1.5 w-1.5 rounded-full bg-green-500 shadow-sm shadow-green-500" />
                <span className="text-[11px] font-semibold text-gray-200">{name || 'Untitled Workflow'}</span>
              </div>
              <div className="mt-2 text-[10px] text-gray-500">
                {nodes.length} nodes, {edges.length} edges
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

      <GraphLibraryModal
        isOpen={isLibraryOpen}
        savedGraphs={savedGraphs}
        examples={exampleGraphs}
        isLoadingSavedGraphs={isLoadingSavedGraphs}
        error={libraryError}
        activeItemKey={activeLibraryItemKey}
        onClose={() => {
          setIsLibraryOpen(false);
          setLibraryError(null);
          setActiveLibraryItemKey(null);
        }}
        onRefreshSavedGraphs={() => {
          void refreshSavedGraphs();
        }}
        onLoadSavedGraph={(graphName) => {
          void handleLoadSavedGraph(graphName);
        }}
        onLoadExampleGraph={(example) => {
          void handleLoadExampleGraph(example);
        }}
      />
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

const getErrorMessage = (error: unknown) => {
  if (error instanceof Error) {
    return error.message;
  }

  if (typeof error === 'string') {
    return error;
  }

  return 'Unknown error';
};

const shouldSuppressDuplicateAgentLog = (
  previous: { type: string; content: any; timestamp: Date },
  type: string,
  content: any,
  timestamp: Date,
) => {
  if (isSystemLogType(type) || isSystemLogType(previous.type)) {
    return false;
  }

  return previous.type === type &&
    comparableLogContent(previous.content) === comparableLogContent(content) &&
    timestamp.getTime() - previous.timestamp.getTime() <= 1000;
};

const isSystemLogType = (type: string) =>
  type === 'info' || type === 'error' || type === 'step' || type === 'route' || type === 'done';

const comparableLogContent = (content: any) => {
  if (typeof content === 'string') {
    return content;
  }

  try {
    return JSON.stringify(content);
  } catch {
    return String(content);
  }
};
