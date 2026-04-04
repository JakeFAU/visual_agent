import React from 'react';
import { Plus, Trash2 } from 'lucide-react';

interface MCPServerConfig {
  name: string;
  command: string;
  args: string[];
  env?: Record<string, string>;
}

interface MCPServerEditorProps {
  servers: MCPServerConfig[];
  onChange: (servers: MCPServerConfig[]) => void;
}

export const MCPServerEditor: React.FC<MCPServerEditorProps> = ({ servers, onChange }) => {
  const addServer = () => {
    onChange([...servers, { name: 'new-server', command: '', args: [] }]);
  };

  const removeServer = (index: number) => {
    onChange(servers.filter((_, i) => i !== index));
  };

  const updateServer = (index: number, updates: Partial<MCPServerConfig>) => {
    onChange(servers.map((s, i) => (i === index ? { ...s, ...updates } : s)));
  };

  return (
    <div className="space-y-4">
      {servers.map((server, index) => (
        <div key={index} className="bg-gray-800/50 border border-gray-700 rounded-md p-3 space-y-3">
          <div className="flex justify-between items-center">
            <input
              type="text"
              value={server.name}
              onChange={(e) => updateServer(index, { name: e.target.value })}
              className="bg-transparent text-sm font-semibold text-blue-400 focus:outline-none focus:ring-0 w-full"
              placeholder="Server Name"
            />
            <button
              onClick={() => removeServer(index)}
              className="text-gray-500 hover:text-red-500 transition-colors"
            >
              <Trash2 size={14} />
            </button>
          </div>

          <div className="space-y-2">
            <div className="space-y-1">
              <label className="text-[9px] font-bold text-gray-500 uppercase tracking-widest">Command</label>
              <input
                type="text"
                value={server.command}
                onChange={(e) => updateServer(index, { command: e.target.value })}
                className="w-full bg-gray-900 border border-gray-700 rounded px-2 py-1 text-xs text-white focus:outline-none focus:ring-1 focus:ring-blue-500"
                placeholder="e.g. npx"
              />
            </div>
            <div className="space-y-1">
              <label className="text-[9px] font-bold text-gray-500 uppercase tracking-widest">Arguments</label>
              <input
                type="text"
                value={server.args.join(' ')}
                onChange={(e) => updateServer(index, { args: e.target.value.split(' ').filter(Boolean) })}
                className="w-full bg-gray-900 border border-gray-700 rounded px-2 py-1 text-xs text-white focus:outline-none focus:ring-1 focus:ring-blue-500 font-mono"
                placeholder="e.g. @modelcontextprotocol/server-filesystem /path"
              />
            </div>
          </div>
        </div>
      ))}

      <button
        onClick={addServer}
        className="w-full py-2 border-2 border-dashed border-gray-700 rounded-md text-gray-500 hover:border-gray-600 hover:text-gray-400 transition-all flex items-center justify-center gap-2 text-xs font-bold uppercase tracking-wider"
      >
        <Plus size={14} /> Add MCP Server
      </button>
    </div>
  );
};
