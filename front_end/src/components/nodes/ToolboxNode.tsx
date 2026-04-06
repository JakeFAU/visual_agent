import React, { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { BaseNode } from './BaseNode';
import { ToolboxNodeConfig } from '../../schema/graph';

export const ToolboxNode = memo(({ data, selected }: NodeProps<{ config: ToolboxNodeConfig }>) => {
  const toolCount = data.config.tools.length;
  const mcpCount = data.config.mcp_servers.length;
  const funcCount = data.config.custom_functions.length;

  return (
    <div className="relative">
      <BaseNode title="Toolbox" selected={selected} color="amber">
        <div className="text-xs text-gray-300 space-y-1">
          <div className="flex justify-between text-[10px]">
              <span>Built-in Tools:</span>
              <span className="font-mono text-amber-400">{toolCount}</span>
          </div>
          <div className="flex justify-between text-[10px]">
              <span>MCP Servers:</span>
              <span className="font-mono text-amber-400">{mcpCount}</span>
          </div>
          <div className="flex justify-between text-[10px]">
              <span>Custom Funcs:</span>
              <span className="font-mono text-amber-400">{funcCount}</span>
          </div>
        </div>
      </BaseNode>
      
      <Handle
        type="source"
        position={Position.Right}
        id="toolbox:out"
        className="w-2 h-2 !bg-amber-500 border-2 border-gray-800"
        style={{ right: -4 }}
      />
    </div>
  );
});

ToolboxNode.displayName = 'ToolboxNode';
