import React, { useMemo, useState, useCallback } from 'react';
import ReactFlow, { 
  Background, 
  Controls, 
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
import { SidePanel } from './components/SidePanel';
import { LogPanel } from './components/LogPanel';
import { saveGraph, executeGraph, loadGraphs, API_BASE } from './api/client';

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

  const nodeTypes = useMemo(() => ({
    input_node: InputNode,
    llm_node: LLMNode,
    toolbox: ToolboxNode,
    output_node: OutputNode,
    if_else_node: IfElseNode,
    while_node: WhileNode,
  }), []);

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
    const graph = exportGraph();
    try {
        await saveGraph(graph);
        addLog('info', `Graph '${graph.name}' saved successfully.`);
    } catch (e) {
        addLog('error', `Failed to save graph: ${e}`);
    }
  };

  const handleLoad = async () => {
    try {
        const names = await loadGraphs();
        if (!names || names.length === 0) {
            alert("No saved graphs found.");
            return;
        }
        const selected = window.prompt(`Enter graph name to load:\nAvailable: ${names.join(', ')}`);
        if (selected) {
            const resp = await fetch(`${API_BASE}/graphs/${selected}`);
            const data = await resp.json();
            importGraph(data);
            addLog('info', `Graph '${selected}' loaded.`);
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
    const userInput = window.prompt("Enter agent input:", "Hello, who are you?");
    if (userInput === null) return;

    setIsLogOpen(true);
    addLog('info', 'Starting execution...');
    
    const graph = exportGraph();
    try {
        await executeGraph(graph, userInput, (event) => {
            console.log("[DEBUG] IDE Received Event:", JSON.stringify(event, null, 2));
            
            const logType = event.type === 'agent_event' ? (event.author || 'agent') : event.type;
            
            let logContent = event.content;
            
            // ADK events embed model.LLMResponse, which contains genai.Content
            // We need to handle both uppercase (Go default) and lowercase (JSON tags)
            const content = event.content?.Content || event.content?.content;
            const parts = content?.Parts || content?.parts;

            if (Array.isArray(parts)) {
                const text = parts.map((p: any) => p.Text || p.text || "").join('');
                if (text) {
                    logContent = text;
                } else {
                    // It might be a tool call or other structured part
                    logContent = JSON.stringify(parts);
                }
            } else if (event.type === 'agent_event' && !logContent) {
                logContent = "Agent message received (no text content)";
            }

            addLog(logType, logContent);
        });
    } catch (e) {
        addLog('error', `Execution failed: ${e}`);
    }
  };

  const addLog = (type: string, content: any) => {
    setLogs(prev => [...prev, { type, content, timestamp: new Date() }]);
  };

  return (
    <div className="flex flex-col h-screen bg-gray-950 text-gray-100 overflow-hidden">
      <header className="h-14 border-b border-gray-800 flex items-center justify-between px-6 bg-gray-900 shrink-0">
        <div className="flex items-center gap-2">
          <div className="w-6 h-6 bg-blue-600 rounded flex items-center justify-center font-bold text-xs shadow-lg shadow-blue-900/20 text-white">V</div>
          <span className="font-bold tracking-tight text-white">Visual Agent <span className="text-gray-600 font-normal ml-1">IDE</span></span>
        </div>
        <div className="flex gap-3">
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
            isOpen={isLogOpen} 
            onToggle={() => setIsLogOpen(!isLogOpen)} 
            onClear={() => setLogs([])}
          />
        </div>
        <SidePanel />
      </main>
    </div>
  );
};

export default App;
