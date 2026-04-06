import React, { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { BaseNode } from './BaseNode';
import { IfElseNodeConfig } from '../../schema/graph';

export const IfElseNode = memo(({ data, selected }: NodeProps<{ config: IfElseNodeConfig }>) => {
  return (
    <div className="relative">
      <BaseNode title="If / Else" selected={selected} color="purple">
        <div className="text-[10px] text-gray-400 mb-1">Condition ({data.config.condition_language})</div>
        <div className="bg-gray-900 rounded p-1.5 font-mono text-[10px] text-purple-300 break-all line-clamp-2">
          {data.config.condition}
        </div>
        
        <div className="flex flex-col gap-2 mt-2 items-end h-10">
          <div className="flex items-center">
              <span className="text-[10px] text-green-500 mr-2 uppercase italic">True</span>
          </div>
          <div className="flex items-center">
              <span className="text-[10px] text-red-500 mr-2 uppercase italic">False</span>
          </div>
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
        id="logic"
        className="w-2 h-2 !bg-transparent border-0 opacity-0 pointer-events-none"
        style={{ left: -4 }}
      />

      <Handle
        type="source"
        position={Position.Right}
        id="message:true"
        className="w-2 h-2 !bg-green-500 border-2 border-gray-800"
        style={{ right: -4, top: '78px' }}
      />
      <Handle
        type="source"
        position={Position.Right}
        id="message:false"
        className="w-2 h-2 !bg-red-500 border-2 border-gray-800"
        style={{ right: -4, top: '98px' }}
      />
      <Handle
        type="source"
        position={Position.Right}
        id="out_true"
        className="w-2 h-2 !bg-transparent border-0 opacity-0 pointer-events-none"
        style={{ right: -4, top: '78px' }}
      />
      <Handle
        type="source"
        position={Position.Right}
        id="out_false"
        className="w-2 h-2 !bg-transparent border-0 opacity-0 pointer-events-none"
        style={{ right: -4, top: '98px' }}
      />
    </div>
  );
});

IfElseNode.displayName = 'IfElseNode';
