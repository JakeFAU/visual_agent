import React, { memo } from 'react';
import { Handle, NodeProps, NodeResizer, Position } from 'reactflow';
import { Repeat } from 'lucide-react';
import { WhileNodeConfig } from '../../schema/graph';

export const WhileNode = memo(({ data, selected }: NodeProps<{ config: WhileNodeConfig }>) => {
  return (
    <div className="relative h-full w-full rounded-xl border border-rose-500/50 bg-rose-950/20 shadow-[0_0_0_1px_rgba(251,113,133,0.08)]">
      <NodeResizer
        isVisible={selected}
        minWidth={360}
        minHeight={240}
        lineClassName="!border-rose-400/70"
        handleClassName="!h-3 !w-3 !rounded-sm !border !border-rose-200 !bg-rose-500"
      />

      <div className="flex h-full flex-col overflow-hidden rounded-xl">
        <div className="border-b border-rose-400/20 bg-gradient-to-r from-rose-950/80 via-rose-900/60 to-rose-950/50 px-4 py-3">
          <div className="flex items-center justify-between gap-3">
            <div className="flex items-center gap-2">
              <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-rose-500/15 text-rose-200">
                <Repeat size={14} />
              </div>
              <div>
                <div className="text-[10px] font-bold uppercase tracking-[0.22em] text-rose-200/70">While Loop</div>
                <div className="text-xs text-rose-100/75">This container reruns its internal subgraph while the condition stays true.</div>
              </div>
            </div>
            <div className="rounded-full border border-rose-300/20 bg-rose-950/70 px-2 py-1 font-mono text-[10px] text-rose-200">
              max {data.config.max_iterations}
            </div>
          </div>
          <div className="mt-3 rounded-lg border border-rose-400/10 bg-gray-950/70 px-3 py-2 font-mono text-[11px] text-rose-100/90">
            {data.config.condition}
          </div>
        </div>

        <div className="relative flex-1 bg-[linear-gradient(rgba(251,113,133,0.04)_1px,transparent_1px),linear-gradient(90deg,rgba(251,113,133,0.04)_1px,transparent_1px)] bg-[size:24px_24px]">
          <div className="absolute inset-x-4 top-4 flex items-center justify-between text-[10px] font-semibold uppercase tracking-[0.18em] text-rose-100/45">
            <span>Internal Subgraph</span>
            <span>Drop Nodes Inside</span>
          </div>
        </div>
      </div>

      <div className="pointer-events-none absolute -left-10 top-[46%] -translate-y-1/2 text-[10px] font-bold uppercase tracking-[0.22em] text-emerald-200/80">
        In
      </div>
      <div className="pointer-events-none absolute -left-[58px] top-[72%] -translate-y-1/2 text-[10px] font-bold uppercase tracking-[0.22em] text-amber-200/80">
        Repeat
      </div>
      <div className="pointer-events-none absolute -right-11 top-1/2 -translate-y-1/2 text-[10px] font-bold uppercase tracking-[0.22em] text-emerald-200/80">
        Out
      </div>

      <Handle
        type="target"
        position={Position.Left}
        id="message:enter"
        className="!z-20 !border-2 !border-gray-950 !bg-emerald-500 shadow-[0_0_18px_rgba(16,185,129,0.35)]"
        style={{ left: -18, top: '46%', width: 26, height: 42, borderRadius: 9999 }}
      />
      <Handle
        type="source"
        position={Position.Left}
        id="message:loop"
        className="!z-20 !border-0 !bg-transparent opacity-0"
        style={{ left: 2, top: '58%', width: 34, height: 64, borderRadius: 9999 }}
      />
      <Handle
        type="target"
        position={Position.Left}
        id="message:return"
        className="!z-20 !border-2 !border-gray-950 !bg-amber-500 shadow-[0_0_18px_rgba(251,191,36,0.3)]"
        style={{ left: -18, top: '72%', width: 26, height: 42, borderRadius: 9999 }}
      />
      <Handle
        type="source"
        position={Position.Right}
        id="message:done"
        className="!z-20 !border-2 !border-gray-950 !bg-emerald-500 shadow-[0_0_18px_rgba(16,185,129,0.35)]"
        style={{ right: -18, top: '50%', width: 26, height: 42, borderRadius: 9999 }}
      />
    </div>
  );
});

WhileNode.displayName = 'WhileNode';
