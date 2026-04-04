import React, { useMemo } from 'react';
import ReactFlow, { 
  Background, 
  Controls, 
  Panel
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

const App: React.FC = () => {
  const { 
    nodes, 
    edges, 
    onNodesChange, 
    onEdgesChange, 
    onConnect,
    exportGraph,
    importGraph
  } = useGraphStore();

  const nodeTypes = useMemo(() => ({
    input_node: InputNode,
    llm_node: LLMNode,
    toolbox: ToolboxNode,
    output_node: OutputNode,
    if_else_node: IfElseNode,
    while_node: WhileNode,
  }), []);

  const handleExport = () => {
    const graph = exportGraph();
    const blob = new Blob([JSON.stringify(graph, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `${graph.name}.json`;
    link.click();
  };

  return (
    <div className="flex flex-col h-screen bg-gray-950 text-gray-100">
      <header className="h-14 border-b border-gray-800 flex items-center justify-between px-6 bg-gray-900">
        <div className="flex items-center gap-2">
          <div className="w-6 h-6 bg-blue-600 rounded flex items-center justify-center font-bold text-xs">V</div>
          <span className="font-bold tracking-tight">Visual Agent</span>
        </div>
        <div className="flex gap-3">
          <button 
            className="px-4 py-1.5 rounded bg-gray-800 border border-gray-700 text-sm font-medium hover:bg-gray-700 transition-colors"
            onClick={() => {
                const input = document.createElement('input');
                input.type = 'file';
                input.onchange = (e: any) => {
                    const file = e.target.files[0];
                    const reader = new FileReader();
                    reader.onload = (re) => importGraph(re.target?.result as string);
                    reader.readAsText(file);
                };
                input.click();
            }}
          >
            Import
          </button>
          <button 
            className="px-4 py-1.5 rounded bg-blue-600 text-white text-sm font-medium hover:bg-blue-500 transition-colors shadow-lg shadow-blue-900/20"
            onClick={handleExport}
          >
            Export & Deploy
          </button>
        </div>
      </header>

      <main className="flex-1 flex overflow-hidden">
        <div className="flex-1 relative">
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            nodeTypes={nodeTypes}
            fitView
          >
            <Background color="#333" gap={20} />
            <Controls />
            <Panel position="top-left" className="bg-gray-900 border border-gray-800 p-2 rounded shadow-xl">
                <div className="text-[10px] font-bold text-gray-500 uppercase tracking-widest mb-1">Status</div>
                <div className="flex items-center gap-2">
                    <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse"></div>
                    <span className="text-xs font-medium">IDE Ready</span>
                </div>
            </Panel>
          </ReactFlow>
        </div>
        <SidePanel />
      </main>
    </div>
  );
};

export default App;
