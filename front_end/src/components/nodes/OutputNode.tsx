import React, { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { BaseNode } from './BaseNode';
import { OutputNodeConfig } from '../../schema/graph';

export const OutputNode = memo(({ data, selected }: NodeProps<{ config: OutputNodeConfig }>) => {
  return (
    <div className="relative">
      <BaseNode title="Final Output" selected={selected} color="gray">
        <div className="font-medium text-white mb-1">{data.config.name}</div>
        <div className="flex justify-between text-[10px] text-gray-400">
          <span>Key: <span className="text-blue-400">{data.config.output_key}</span></span>
          <span>Format: <span className="text-purple-400">{data.config.format}</span></span>
        </div>
      </BaseNode>
      
      <Handle
        type="target"
        position={Position.Left}
        id="message"
        className="w-2 h-2 !bg-green-500 border-2 border-gray-800"
        style={{ left: -4 }}
      />
      <Handle
        type="target"
        position={Position.Left}
        id="in_message"
        className="w-2 h-2 !bg-transparent border-0 opacity-0 pointer-events-none"
        style={{ left: -4 }}
      />
    </div>
  );
});

OutputNode.displayName = 'OutputNode';
