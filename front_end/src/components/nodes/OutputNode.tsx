import React, { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { BaseNode } from './BaseNode';
import { OutputNodeConfig } from '../../schema/graph';

export const OutputNode = memo(({ data, selected }: NodeProps<{ config: OutputNodeConfig }>) => {
  return (
    <BaseNode title="Final Output" selected={selected} color="gray">
      <div className="font-medium text-white mb-1">{data.config.name}</div>
      <div className="flex justify-between text-[10px] text-gray-400">
        <span>Key: <span className="text-blue-400">{data.config.output_key}</span></span>
        <span>Format: <span className="text-purple-400">{data.config.format}</span></span>
      </div>
      
      <Handle
        type="target"
        position={Position.Left}
        id="in_result"
        className="w-2 h-2 !bg-green-500 border-2 border-gray-800"
      />
    </BaseNode>
  );
});

OutputNode.displayName = 'OutputNode';
