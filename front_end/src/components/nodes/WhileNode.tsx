import React, { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { BaseNode } from './BaseNode';
import { WhileNodeConfig } from '../../schema/graph';

export const WhileNode = memo(({ data, selected }: NodeProps<{ config: WhileNodeConfig }>) => {
  return (
    <BaseNode title="While / Loop" selected={selected} color="purple">
      <div className="text-[10px] text-gray-400 mb-1">Condition</div>
      <div className="bg-gray-900 rounded p-1.5 font-mono text-[10px] text-blue-300 break-all mb-2">
        {data.config.condition}
      </div>
      <div className="text-[10px] text-gray-500 italic">
        Max Iterations: {data.config.max_iterations}
      </div>
      
      <Handle
        type="target"
        position={Position.Left}
        id="in_start"
        className="w-2 h-2 !bg-gray-500 border-2 border-gray-800"
      />

      <div className="flex flex-col gap-2 mt-2 items-end">
        <div className="relative h-4 w-full flex justify-end">
            <span className="text-[10px] text-blue-400 mr-2 uppercase italic">Loop</span>
            <Handle
                type="source"
                position={Position.Right}
                id="out_loop"
                className="w-2 h-2 !bg-blue-400 border-2 border-gray-800"
                style={{ top: '50%' }}
            />
        </div>
        <div className="relative h-4 w-full flex justify-end">
            <span className="text-[10px] text-gray-400 mr-2 uppercase italic">Done</span>
            <Handle
                type="source"
                position={Position.Right}
                id="out_done"
                className="w-2 h-2 !bg-gray-400 border-2 border-gray-800"
                style={{ top: '50%' }}
            />
        </div>
      </div>
    </BaseNode>
  );
});

WhileNode.displayName = 'WhileNode';
