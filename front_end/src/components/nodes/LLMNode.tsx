import React, { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { BaseNode } from './BaseNode';
import { LLMNodeConfig } from '../../schema/graph';

export const LLMNode = memo(({ data, selected }: NodeProps<{ config: LLMNodeConfig }>) => {
  const isStructured = data.config.response_mode === 'json';

  return (
    <div className="relative">
      <BaseNode title="LLM Agent" selected={selected} color="blue">
        <div className="font-medium text-white mb-1">{data.config.name}</div>
        <div className="text-[10px] text-gray-400 bg-gray-900/50 px-1.5 py-0.5 rounded inline-block mb-2">
          {data.config.model}
        </div>
        
        <div className="flex flex-col gap-2 mt-2">
          <div className="h-4 flex items-center">
              <span className="text-[10px] text-gray-500">IN MSG</span>
          </div>
          <div className="h-4 flex items-center">
              <span className="text-[10px] text-gray-500">TOOLBOX</span>
          </div>
        </div>

        <div className="mt-4 flex justify-end h-4 items-center">
            <span className="text-[10px] text-gray-500 mr-2 uppercase">
                {isStructured ? 'Structured' : 'Message'}
            </span>
        </div>
      </BaseNode>

      {/* Target Handles */}
      <Handle
        type="target"
        position={Position.Left}
        id="message"
        className="w-2 h-2 !bg-green-500 border-2 border-gray-800"
        style={{ left: -4, top: '62px' }}
      />
      <Handle
        type="target"
        position={Position.Left}
        id="toolbox_handle"
        className="w-2 h-2 !bg-amber-500 border-2 border-gray-800"
        style={{ left: -4, top: '86px' }}
      />

      {/* Source Handle */}
      <Handle
        type="source"
        position={Position.Right}
        id={isStructured ? "structured_output" : "message"}
        className={`w-2 h-2 border-2 border-gray-800 ${isStructured ? '!bg-purple-500' : '!bg-green-500'}`}
        style={{ right: -4, top: '118px' }}
      />
    </div>
  );
});

LLMNode.displayName = 'LLMNode';
