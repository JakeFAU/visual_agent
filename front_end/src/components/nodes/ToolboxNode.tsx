import React, { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { BaseNode } from './BaseNode';
import { ToolboxNodeConfig } from '../../schema/graph';

export const ToolboxNode = memo(({ data, selected }: NodeProps<{ config: ToolboxNodeConfig }>) => {
  const toolCount = data.config.tools.length;
  const mcpCount = data.config.mcp_servers.length;
  const funcCount = data.config.custom_functions.length;

  return (
    <BaseNode title="Toolbox" selected={selected} color="amber">
      <div className="text-xs text-gray-300 space-y-1">
        <div className="flex justify-between">
            <span>Built-in Tools:</span>
            <span className="font-mono text-amber-400">{toolCount}</span>
        </div>
        <div className="flex justify-between">
            <span>MCP Servers:</span>
            <span className="font-mono text-amber-400">{mcpCount}</span>
        </div>
        <div className="flex justify-between">
            <span>Custom Funcs:</span>
            <span className="font-mono text-amber-400">{funcCount}</span>
        </div>
      </div>
      
      <Handle
        type="source"
        position={Position.Right}
        id="in_toolbox"
        className="w-2 h-2 !bg-amber-500 border-2 border-gray-800"
      />
    </BaseNode>
  );
});

ToolboxNode.displayName = 'ToolboxNode';
