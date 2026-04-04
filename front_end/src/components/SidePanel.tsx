import React from 'react';
import { useGraphStore } from '../store/useGraphStore';

export const SidePanel: React.FC = () => {
  const { selectedNodeId, nodes, updateNodeConfig } = useGraphStore();
  const selectedNode = nodes.find((n) => n.id === selectedNodeId);

  if (!selectedNode) {
    return (
      <div className="w-[320px] bg-gray-900 border-l border-gray-800 p-6 flex flex-col items-center justify-center text-gray-500">
        <p className="text-sm">Select a node to configure</p>
      </div>
    );
  }

  const config = selectedNode.data.config;

  const handleChange = (key: string, value: any) => {
    updateNodeConfig(selectedNode.id, { ...config, [key]: value });
  };

  return (
    <div className="w-[320px] bg-gray-900 border-l border-gray-800 p-6 overflow-y-auto">
      <h3 className="text-lg font-semibold text-white mb-6">Configuration</h3>
      
      <div className="space-y-6">
        <div className="space-y-2">
          <label className="text-[10px] font-bold text-gray-500 uppercase tracking-wider">Node Type</label>
          <div className="text-sm text-blue-400 font-mono">{selectedNode.type}</div>
        </div>

        {selectedNode.type === 'input_node' && (
          <>
            <div className="space-y-2">
              <label className="text-[10px] font-bold text-gray-500 uppercase tracking-wider">Name</label>
              <input
                type="text"
                value={config.name}
                onChange={(e) => handleChange('name', e.target.value)}
                className="w-full bg-gray-800 border border-gray-700 rounded p-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-bold text-gray-500 uppercase tracking-wider">Description</label>
              <textarea
                value={config.description}
                onChange={(e) => handleChange('description', e.target.value)}
                className="w-full bg-gray-800 border border-gray-700 rounded p-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-blue-500 h-24"
              />
            </div>
          </>
        )}

        {selectedNode.type === 'llm_node' && (
          <>
            <div className="space-y-2">
              <label className="text-[10px] font-bold text-gray-500 uppercase tracking-wider">Agent Name</label>
              <input
                type="text"
                value={config.name}
                onChange={(e) => handleChange('name', e.target.value)}
                className="w-full bg-gray-800 border border-gray-700 rounded p-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-bold text-gray-500 uppercase tracking-wider">Model</label>
              <select
                value={config.model}
                onChange={(e) => handleChange('model', e.target.value)}
                className="w-full bg-gray-800 border border-gray-700 rounded p-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                <option value="gemini-2.5-flash">Gemini 2.5 Flash</option>
                <option value="gemini-2.5-pro">Gemini 2.5 Pro</option>
              </select>
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-bold text-gray-500 uppercase tracking-wider">Instruction</label>
              <textarea
                value={config.instruction}
                onChange={(e) => handleChange('instruction', e.target.value)}
                className="w-full bg-gray-800 border border-gray-700 rounded p-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-blue-500 h-32"
              />
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-bold text-gray-500 uppercase tracking-wider">Response Mode</label>
              <div className="flex gap-2">
                <button
                  onClick={() => handleChange('response_mode', 'text')}
                  className={`flex-1 px-3 py-1.5 rounded text-xs font-medium transition-colors ${config.response_mode === 'text' ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-750'}`}
                >
                  Text
                </button>
                <button
                  onClick={() => handleChange('response_mode', 'json')}
                  className={`flex-1 px-3 py-1.5 rounded text-xs font-medium transition-colors ${config.response_mode === 'json' ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-750'}`}
                >
                  JSON
                </button>
              </div>
            </div>
          </>
        )}

        {/* Toolbox, If/Else, While etc. forms follow same pattern */}
        {(selectedNode.type === 'if_else_node' || selectedNode.type === 'while_node') && (
            <div className="space-y-2">
                <label className="text-[10px] font-bold text-gray-500 uppercase tracking-wider">Condition</label>
                <textarea
                    value={config.condition}
                    onChange={(e) => handleChange('condition', e.target.value)}
                    className="w-full bg-gray-800 border border-gray-700 rounded p-2 text-sm text-white font-mono focus:outline-none focus:ring-1 focus:ring-blue-500 h-24"
                />
            </div>
        )}
      </div>
    </div>
  );
};
