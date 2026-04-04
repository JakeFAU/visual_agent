import React from 'react';
import { Plus, Trash2, Hash, Type, ToggleLeft, Braces } from 'lucide-react';

interface SchemaField {
  name: string;
  type: 'string' | 'number' | 'boolean' | 'object';
  description: string;
}

interface SchemaFieldBuilderProps {
  schema: any;
  onChange: (schema: any) => void;
}

export const SchemaFieldBuilder: React.FC<SchemaFieldBuilderProps> = ({ schema, onChange }) => {
  // Convert Draft 7 properties to internal list
  const properties = schema?.properties || {};
  const fields: SchemaField[] = Object.keys(properties).map(name => ({
    name,
    type: properties[name].type,
    description: properties[name].description || '',
  }));

  const updateSchema = (newFields: SchemaField[]) => {
    const newProperties: any = {};
    newFields.forEach(f => {
      newProperties[f.name] = {
        type: f.type,
        description: f.description,
      };
    });
    onChange({
      type: 'object',
      properties: newProperties,
      required: newFields.map(f => f.name),
    });
  };

  const addField = () => {
    updateSchema([...fields, { name: 'new_field', type: 'string', description: '' }]);
  };

  const removeField = (name: string) => {
    updateSchema(fields.filter(f => f.name !== name));
  };

  const updateField = (name: string, updates: Partial<SchemaField>) => {
    updateSchema(fields.map(f => f.name === name ? { ...f, ...updates } : f));
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'number': return <Hash size={10} />;
      case 'boolean': return <ToggleLeft size={10} />;
      case 'object': return <Braces size={10} />;
      default: return <Type size={10} />;
    }
  };

  return (
    <div className="space-y-3">
      {fields.map((field, index) => (
        <div key={index} className="bg-gray-800/50 border border-gray-700 rounded p-2 space-y-2">
          <div className="flex justify-between items-center gap-2">
            <div className="flex items-center gap-1.5 flex-1 bg-gray-900 rounded px-2 py-1 border border-gray-800">
                {getTypeIcon(field.type)}
                <input
                    type="text"
                    value={field.name}
                    onChange={(e) => updateField(field.name, { name: e.target.value })}
                    className="bg-transparent text-[11px] font-mono text-purple-400 focus:outline-none w-full"
                />
            </div>
            <select
                value={field.type}
                onChange={(e) => updateField(field.name, { type: e.target.value as any })}
                className="bg-gray-900 border border-gray-800 text-[10px] text-gray-400 rounded px-1 py-1 focus:outline-none"
            >
                <option value="string">String</option>
                <option value="number">Number</option>
                <option value="boolean">Boolean</option>
                <option value="object">Object</option>
            </select>
            <button onClick={() => removeField(field.name)} className="text-gray-600 hover:text-red-500 transition-colors">
                <Trash2 size={12} />
            </button>
          </div>
          <input
            type="text"
            placeholder="Field description..."
            value={field.description}
            onChange={(e) => updateField(field.name, { description: e.target.value })}
            className="w-full bg-transparent text-[10px] text-gray-500 italic px-1 focus:outline-none border-b border-transparent focus:border-gray-700"
          />
        </div>
      ))}
      
      <button
        onClick={addField}
        className="w-full py-1.5 border border-dashed border-gray-700 rounded text-[10px] font-bold uppercase tracking-wider text-gray-500 hover:border-gray-600 hover:text-gray-400 flex items-center justify-center gap-2"
      >
        <Plus size={12} /> Add Field
      </button>
    </div>
  );
};
