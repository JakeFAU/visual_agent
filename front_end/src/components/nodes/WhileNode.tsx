import React, { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { BaseNode } from './BaseNode';
import { WhileNodeConfig } from '../../schema/graph';

export const WhileNode = memo(({ data, selected }: NodeProps<{ config: WhileNodeConfig }>) => {
  return (
    <div className="relative">
      <BaseNode title="While / Loop" selected={selected} color="purple">
        <div className="text-[10px] text-gray-400 mb-1">Condition</div>
        <div className="bg-gray-900 rounded p-1.5 font-mono text-[10px] text-blue-300 break-all mb-2">
          {data.config.condition}
        </div>
        <div className="text-[10px] text-gray-500 italic">
          Max Iterations: {data.config.max_iterations}
        </div>
        
        <div className="flex flex-col gap-2 mt-2 items-end h-10">
          <div className="flex items-center">
              <span className="text-[10px] text-blue-400 mr-2 uppercase italic">Loop</span>
          </div>
          <div className="flex items-center">
              <span className="text-[10px] text-gray-400 mr-2 uppercase italic">Done</span>
          </div>
        </div>
      </BaseNode>
      
      <Handle
        type="target"
        position={Position.Left}
        id="logic"
        className="w-2 h-2 !bg-gray-500 border-2 border-gray-800"
        style={{ left: -4 }}
      />

      <Handle
        type="source"
        position={Position.Right}
        id="logic"
        className="w-2 h-2 !bg-blue-400 border-2 border-gray-800"
        style={{ right: -4, top: '88px' }}
      />
      <Handle
        type="source"
        position={Position.Right}
        id="logic"
        className="w-2 h-2 !bg-gray-400 border-2 border-gray-800"
        style={{ right: -4, top: '108px' }}
      />
    </div>
  );
});

WhileNode.displayName = 'WhileNode';
