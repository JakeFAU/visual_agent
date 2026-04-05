import React from 'react';

interface ToolListEditorProps {
  selectedTools: string[];
  onChange: (tools: string[]) => void;
}

const AVAILABLE_TOOLS = [
  { id: 'google_search', name: 'Google Search', description: 'Search the web for real-time information.' },
];

export const ToolListEditor: React.FC<ToolListEditorProps> = ({ selectedTools, onChange }) => {
  const toggleTool = (toolId: string) => {
    if (selectedTools.includes(toolId)) {
      onChange(selectedTools.filter((id) => id !== toolId));
    } else {
      onChange([...selectedTools, toolId]);
    }
  };

  return (
    <div className="space-y-2">
      {AVAILABLE_TOOLS.map((tool) => (
        <label
          key={tool.id}
          className={`flex items-start gap-3 p-3 rounded-md border cursor-pointer transition-colors ${
            selectedTools.includes(tool.id)
              ? 'bg-blue-900/20 border-blue-500/50'
              : 'bg-gray-800/50 border-gray-700 hover:border-gray-600'
          }`}
        >
          <input
            type="checkbox"
            checked={selectedTools.includes(tool.id)}
            onChange={() => toggleTool(tool.id)}
            className="mt-1 rounded border-gray-700 text-blue-600 focus:ring-blue-500 focus:ring-offset-gray-900 bg-gray-900"
          />
          <div>
            <div className="text-sm font-medium text-white">{tool.name}</div>
            <div className="text-[10px] text-gray-500 mt-0.5">{tool.description}</div>
          </div>
        </label>
      ))}
    </div>
  );
};
