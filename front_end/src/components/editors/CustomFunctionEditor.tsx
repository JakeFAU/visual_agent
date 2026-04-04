import React, { useState } from 'react';
import { Plus, Trash2, Code } from 'lucide-react';

interface CustomFunctionConfig {
  name: string;
  description: string;
  parameters: Record<string, any>;
}

interface CustomFunctionEditorProps {
  functions: CustomFunctionConfig[];
  onChange: (functions: CustomFunctionConfig[]) => void;
}

export const CustomFunctionEditor: React.FC<CustomFunctionEditorProps> = ({ functions, onChange }) => {
  const [errorIndex, setErrorIndex] = useState<number | null>(null);

  const addFunction = () => {
    onChange([...functions, { 
      name: 'new_function', 
      description: '', 
      parameters: { type: 'object', properties: {} } 
    }]);
  };

  const removeFunction = (index: number) => {
    onChange(functions.filter((_, i) => i !== index));
  };

  const updateFunction = (index: number, updates: Partial<CustomFunctionConfig>) => {
    onChange(functions.map((f, i) => (i === index ? { ...f, ...updates } : f)));
  };

  const handleParamsChange = (index: number, value: string) => {
    try {
      const parsed = JSON.parse(value);
      updateFunction(index, { parameters: parsed });
      setErrorIndex(null);
    } catch (e) {
      setErrorIndex(index);
    }
  };

  return (
    <div className="space-y-4">
      {functions.map((fn, index) => (
        <div key={index} className="bg-gray-800/50 border border-gray-700 rounded-md p-3 space-y-3">
          <div className="flex justify-between items-center">
            <input
              type="text"
              value={fn.name}
              onChange={(e) => updateFunction(index, { name: e.target.value })}
              className="bg-transparent text-sm font-semibold text-purple-400 focus:outline-none focus:ring-0 w-full"
              placeholder="function_name"
            />
            <button
              onClick={() => removeFunction(index)}
              className="text-gray-500 hover:text-red-500 transition-colors"
            >
              <Trash2 size={14} />
            </button>
          </div>

          <div className="space-y-2">
            <div className="space-y-1">
              <label className="text-[9px] font-bold text-gray-500 uppercase tracking-widest">Description</label>
              <input
                type="text"
                value={fn.description}
                onChange={(e) => updateFunction(index, { description: e.target.value })}
                className="w-full bg-gray-900 border border-gray-700 rounded px-2 py-1 text-xs text-white focus:outline-none focus:ring-1 focus:ring-blue-500"
                placeholder="What this function does"
              />
            </div>
            <div className="space-y-1">
              <div className="flex justify-between items-center">
                <label className="text-[9px] font-bold text-gray-500 uppercase tracking-widest flex items-center gap-1">
                  <Code size={10} /> Parameters (JSON Schema)
                </label>
                {errorIndex === index && <span className="text-[8px] text-red-500 font-bold uppercase">Invalid JSON</span>}
              </div>
              <textarea
                defaultValue={JSON.stringify(fn.parameters, null, 2)}
                onChange={(e) => handleParamsChange(index, e.target.value)}
                className={`w-full bg-gray-950 border ${errorIndex === index ? 'border-red-500' : 'border-gray-700'} rounded px-2 py-1 text-[10px] text-gray-300 focus:outline-none focus:ring-1 focus:ring-blue-500 font-mono h-32 resize-none`}
              />
            </div>
          </div>
        </div>
      ))}

      <button
        onClick={addFunction}
        className="w-full py-2 border-2 border-dashed border-gray-700 rounded-md text-gray-500 hover:border-gray-600 hover:text-gray-400 transition-all flex items-center justify-center gap-2 text-xs font-bold uppercase tracking-wider"
      >
        <Plus size={14} /> Add Function
      </button>
    </div>
  );
};
