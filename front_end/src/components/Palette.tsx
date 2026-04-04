import React from 'react';
import { MessageSquare, Cpu, Briefcase, Play, Split, Repeat } from 'lucide-react';

const NODE_TYPES = [
  { type: 'input_node', label: 'Input', icon: <MessageSquare size={16} />, color: 'bg-green-500' },
  { type: 'llm_node', label: 'LLM Agent', icon: <Cpu size={16} />, color: 'bg-blue-500' },
  { type: 'toolbox', label: 'Toolbox', icon: <Briefcase size={16} />, color: 'bg-amber-500' },
  { type: 'output_node', label: 'Output', icon: <Play size={16} />, color: 'bg-gray-500' },
  { type: 'if_else_node', label: 'If / Else', icon: <Split size={16} />, color: 'bg-purple-500' },
  { type: 'while_node', label: 'While', icon: <Repeat size={16} />, color: 'bg-purple-500' },
];

export const Palette: React.FC = () => {
  const onDragStart = (event: React.DragEvent, nodeType: string) => {
    event.dataTransfer.setData('application/reactflow', nodeType);
    event.dataTransfer.effectAllowed = 'move';
  };

  return (
    <aside className="w-16 bg-gray-900 border-r border-gray-800 flex flex-col items-center py-4 gap-4 shrink-0 overflow-y-auto custom-scrollbar">
      <div className="text-[8px] font-bold text-gray-600 uppercase tracking-tighter mb-2 text-center">Nodes</div>
      {NODE_TYPES.map((node) => (
        <div
          key={node.type}
          className={`w-10 h-10 rounded-lg ${node.color} flex items-center justify-center cursor-grab active:cursor-grabbing shadow-lg hover:brightness-110 transition-all text-white`}
          onDragStart={(event) => onDragStart(event, node.type)}
          draggable
          title={node.label}
        >
          {node.icon}
        </div>
      ))}
    </aside>
  );
};
