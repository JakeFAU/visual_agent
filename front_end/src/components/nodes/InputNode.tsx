import React, { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { BaseNode } from './BaseNode';
import { InputNodeConfig } from '../../schema/graph';

export const InputNode = memo(({ data, selected }: NodeProps<{ config: InputNodeConfig }>) => {
  return (
    <div className="relative">
      <BaseNode title="User Input" selected={selected} color="green">
        <div className="font-medium text-white mb-1">{data.config.name}</div>
        <div className="text-xs text-gray-400 line-clamp-2">{data.config.description}</div>
      </BaseNode>
      
      <Handle
        type="source"
        position={Position.Right}
        id="message"
        className="w-2 h-2 !bg-green-500 border-2 border-gray-800"
        style={{ right: -4 }}
      />
    </div>
  );
});

InputNode.displayName = 'InputNode';
