import React, { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { BaseNode } from './BaseNode';
import { LLMNodeConfig } from '../../schema/graph';

export const LLMNode = memo(({ data, selected }: NodeProps<{ config: LLMNodeConfig }>) => {
  const isStructured = data.config.response_mode === 'json';

  return (
    <BaseNode title="LLM Agent" selected={selected} color="blue">
      <div className="font-medium text-white mb-1">{data.config.name}</div>
      <div className="text-[10px] text-gray-400 bg-gray-900/50 px-1.5 py-0.5 rounded inline-block mb-2">
        {data.config.model}
      </div>
      
      <div className="flex flex-col gap-2 mt-2">
        <div className="relative h-4">
            <span className="text-[10px] text-gray-500 absolute left-0">IN MSG</span>
            <Handle
                type="target"
                position={Position.Left}
                id="in_message"
                className="w-2 h-2 !bg-green-500 border-2 border-gray-800"
                style={{ top: '50%' }}
            />
        </div>
        <div className="relative h-4">
            <span className="text-[10px] text-gray-500 absolute left-0">TOOLBOX</span>
            <Handle
                type="target"
                position={Position.Left}
                id="in_toolbox"
                className="w-2 h-2 !bg-amber-500 border-2 border-gray-800"
                style={{ top: '50%' }}
            />
        </div>
      </div>

      <div className="mt-4 flex justify-end">
        <div className="relative h-4 w-full flex justify-end">
            <span className="text-[10px] text-gray-500 mr-2 uppercase">
                {isStructured ? 'Structured' : 'Message'}
            </span>
            <Handle
                type="source"
                position={Position.Right}
                id="out_message"
                className={`w-2 h-2 border-2 border-gray-800 ${isStructured ? '!bg-purple-500' : '!bg-green-500'}`}
                style={{ top: '50%' }}
            />
        </div>
      </div>
    </BaseNode>
  );
});

LLMNode.displayName = 'LLMNode';
