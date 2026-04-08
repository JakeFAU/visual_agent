import React from 'react';
import { useGraphStore } from '../store/useGraphStore';
import { Accordion } from './ui/Accordion';
import { ToolListEditor } from './editors/ToolListEditor';
import { MCPServerEditor } from './editors/MCPServerEditor';
import { SchemaFieldBuilder } from './editors/SchemaFieldBuilder';

export const SidePanel: React.FC = () => {
  const { selectedNodeId, nodes, updateNodeConfig } = useGraphStore();
  const selectedNode = nodes.find((n) => n.id === selectedNodeId);

  if (!selectedNode) {
    return (
      <div className="w-[320px] bg-gray-900 border-l border-gray-800 p-6 flex flex-col items-center justify-center text-gray-500">
        <p className="text-sm font-medium tracking-tight">Select a node to configure</p>
      </div>
    );
  }

  const config = selectedNode.data.config;

  const handleChange = (key: string, value: any) => {
    updateNodeConfig(selectedNode.id, { ...config, [key]: value });
  };

  return (
    <div className="w-[320px] bg-gray-900 border-l border-gray-800 flex flex-col h-full">
      <div className="p-6 border-b border-gray-800">
        <h3 className="text-lg font-semibold text-white">Configuration</h3>
        <div className="text-[10px] text-gray-500 font-mono mt-1 uppercase tracking-widest">{selectedNode.type}</div>
      </div>
      
      <div className="flex-1 overflow-y-auto p-6 space-y-8 custom-scrollbar">
        {selectedNode.type === 'input_node' && (
          <div className="space-y-6">
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
                className="w-full bg-gray-800 border border-gray-700 rounded p-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-blue-500 h-24 resize-none"
              />
            </div>
          </div>
        )}

        {selectedNode.type === 'llm_node' && (
          <div className="space-y-6">
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
                className="w-full bg-gray-800 border border-gray-700 rounded p-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-blue-500 h-32 resize-none"
              />
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-bold text-gray-500 uppercase tracking-wider">Response Mode</label>
              <div className="flex gap-2">
                <button
                  onClick={() => handleChange('response_mode', 'text')}
                  className={`flex-1 px-3 py-1.5 rounded text-[10px] font-bold uppercase tracking-wider transition-colors ${config.response_mode === 'text' ? 'bg-blue-600 text-white shadow-lg shadow-blue-900/20' : 'bg-gray-800 text-gray-400 hover:bg-gray-750 border border-gray-700'}`}
                >
                  Text
                </button>
                <button
                  onClick={() => handleChange('response_mode', 'json')}
                  className={`flex-1 px-3 py-1.5 rounded text-[10px] font-bold uppercase tracking-wider transition-colors ${config.response_mode === 'json' ? 'bg-purple-600 text-white shadow-lg shadow-purple-900/20' : 'bg-gray-800 text-gray-400 hover:bg-gray-750 border border-gray-700'}`}
                >
                  JSON
                </button>
              </div>
            </div>
            {config.response_mode === 'json' && (
                <div className="space-y-2">
                    <label className="text-[10px] font-bold text-gray-500 uppercase tracking-wider">Output Schema</label>
                    <SchemaFieldBuilder 
                        schema={config.output_schema || { type: 'object', properties: {} }} 
                        onChange={(s) => handleChange('output_schema', s)}
                    />
                </div>
            )}
          </div>
        )}

        {selectedNode.type === 'toolbox' && (
          <div className="space-y-2">
            <Accordion title="Built-in Tools" count={config.tools.length} defaultOpen>
              <ToolListEditor
                selectedTools={config.tools}
                onChange={(tools) => handleChange('tools', tools)}
              />
            </Accordion>
            <Accordion title="MCP Servers" count={config.mcp_servers.length}>
              <MCPServerEditor
                servers={config.mcp_servers}
                onChange={(servers) => handleChange('mcp_servers', servers)}
              />
            </Accordion>
            <Accordion title="Custom Functions (TODO)" count={config.custom_functions.length}>
              <div className="rounded-md border border-amber-500/20 bg-amber-950/20 p-3 text-xs leading-5 text-amber-100/85">
                Custom function declarations are not executable yet.
                They remain a planned feature and are intentionally excluded from the runtime for now.
              </div>
              {config.custom_functions.length > 0 && (
                <div className="space-y-2">
                  {config.custom_functions.map((fn: { name: string; description?: string }, index: number) => (
                    <div key={`${fn.name}-${index}`} className="rounded-md border border-gray-700 bg-gray-800/40 p-3">
                      <div className="text-sm font-semibold text-gray-100">{fn.name}</div>
                      <div className="mt-1 text-[11px] text-gray-400">
                        {fn.description?.trim() || 'No description provided.'}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </Accordion>
          </div>
        )}

        {selectedNode.type === 'if_else_node' && (
            <div className="space-y-4">
                <div className="space-y-2">
                    <label className="text-[10px] font-bold text-gray-500 uppercase tracking-wider">Language</label>
                    <div className="w-full bg-gray-800 border border-gray-700 rounded p-2 text-sm text-white">
                        Common Expression Language (CEL)
                    </div>
                </div>
                <div className="space-y-2">
                    <label className="text-[10px] font-bold text-gray-500 uppercase tracking-wider">Condition</label>
                    <textarea
                        value={config.condition}
                        onChange={(e) => handleChange('condition', e.target.value)}
                        className="w-full bg-gray-800 border border-gray-700 rounded p-2 text-sm text-white font-mono focus:outline-none focus:ring-1 focus:ring-blue-500 h-24 resize-none"
                        placeholder="e.g. state.analyze.status == 'pass'"
                    />
                </div>
            </div>
        )}

        {selectedNode.type === 'while_node' && (
            <div className="space-y-4">
                <div className="rounded border border-rose-900/60 bg-rose-950/20 p-3 text-[11px] text-rose-100/80">
                    This container owns its loop policy. Nodes placed inside it visually define the loop body, while the header controls still apply as run-wide budgets.
                </div>
                <div className="space-y-2">
                    <label className="text-[10px] font-bold text-gray-500 uppercase tracking-wider">Language</label>
                    <div className="w-full bg-gray-800 border border-gray-700 rounded p-2 text-sm text-white">
                        Common Expression Language (CEL)
                    </div>
                </div>
                <div className="space-y-2">
                    <label className="text-[10px] font-bold text-gray-500 uppercase tracking-wider">Condition</label>
                    <textarea
                        value={config.condition}
                        onChange={(e) => handleChange('condition', e.target.value)}
                        className="w-full bg-gray-800 border border-gray-700 rounded p-2 text-sm text-white font-mono focus:outline-none focus:ring-1 focus:ring-rose-500 h-24 resize-none"
                        placeholder="e.g. state.review.status != 'pass' || iteration < 2"
                    />
                    <div className="text-[10px] text-gray-500">
                        Available variables: <span className="font-mono text-gray-300">state</span>, <span className="font-mono text-gray-300">iteration</span>, <span className="font-mono text-gray-300">max_iterations</span>
                    </div>
                </div>
                <div className="space-y-2">
                    <label className="text-[10px] font-bold text-gray-500 uppercase tracking-wider">Max Iterations</label>
                    <input
                        type="number"
                        min="1"
                        value={config.max_iterations}
                        onChange={(e) => handleChange('max_iterations', Math.max(1, Number.parseInt(e.target.value || '1', 10) || 1))}
                        className="w-full bg-gray-800 border border-gray-700 rounded p-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-rose-500"
                    />
                </div>
            </div>
        )}
      </div>
    </div>
  );
};
