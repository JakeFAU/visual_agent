import React, { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { BaseNode } from './BaseNode';
import { IfElseNodeConfig } from '../../schema/graph';

export const IfElseNode = memo(({ data, selected }: NodeProps<{ config: IfElseNodeConfig }>) => {
  return (
    <BaseNode title="If / Else" selected={selected} color="purple">
      <div className="text-[10px] text-gray-400 mb-1">Condition ({data.config.condition_language})</div>
      <div className="bg-gray-900 rounded p-1.5 font-mono text-[10px] text-purple-300 break-all line-clamp-2">
        {data.config.condition}
      </div>
      
      <Handle
        type="target"
        position={Position.Left}
        id="in_data"
        className="w-2 h-2 !bg-blue-500 border-2 border-gray-800"
      />

      <div className="flex flex-col gap-2 mt-2 items-end">
        <div className="relative h-4 w-full flex justify-end">
            <span className="text-[10px] text-green-500 mr-2 uppercase italic">True</span>
            <Handle
                type="source"
                position={Position.Right}
                id="out_true"
                className="w-2 h-2 !bg-green-500 border-2 border-gray-800"
                style={{ top: '50%' }}
            />
        </div>
        <div className="relative h-4 w-full flex justify-end">
            <span className="text-[10px] text-red-500 mr-2 uppercase italic">False</span>
            <Handle
                type="source"
                position={Position.Right}
                id="out_false"
                className="w-2 h-2 !bg-red-500 border-2 border-gray-800"
                style={{ top: '50%' }}
            />
        </div>
      </div>
    </BaseNode>
  );
});

IfElseNode.displayName = 'IfElseNode';
