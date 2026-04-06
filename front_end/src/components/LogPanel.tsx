import React, { useEffect, useRef } from 'react';
import { Terminal, ChevronUp, ChevronDown } from 'lucide-react';

interface LogEntry {
  type: string;
  content: any;
  timestamp: Date;
}

interface RuntimeStats {
  steps: number;
  elapsedMs: number;
  totalTokens: number;
  currentNode: string | null;
  exitReason: string | null;
  nodeVisits: Record<string, number>;
}

interface LogPanelProps {
  logs: LogEntry[];
  response: string | null;
  stats: RuntimeStats | null;
  isOpen: boolean;
  onToggle: () => void;
  onClear: () => void;
}

export const LogPanel: React.FC<LogPanelProps> = ({ logs, response, stats, isOpen, onToggle, onClear }) => {
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [logs]);

  return (
    <div className={`fixed bottom-0 left-0 right-[320px] bg-gray-900 border-t border-gray-800 transition-all duration-300 z-50 ${isOpen ? 'h-64' : 'h-10'}`}>
      <div className="h-10 flex items-center justify-between px-4 border-b border-gray-800 bg-gray-950/50 cursor-pointer" onClick={onToggle}>
        <div className="flex items-center gap-2">
          <Terminal size={14} className="text-blue-400" />
          <span className="text-[10px] font-bold text-gray-400 uppercase tracking-widest">Execution Logs</span>
          <span className="bg-gray-800 text-gray-500 text-[9px] px-1.5 py-0.5 rounded-full font-mono">{logs.length}</span>
        </div>
        <div className="flex items-center gap-4">
            <button 
                onClick={(e) => { e.stopPropagation(); onClear(); }}
                className="text-[9px] font-bold text-gray-600 hover:text-gray-400 uppercase tracking-wider"
            >
                Clear
            </button>
            {isOpen ? <ChevronDown size={14} className="text-gray-500" /> : <ChevronUp size={14} className="text-gray-500" />}
        </div>
      </div>

      {isOpen && (
        <div ref={scrollRef} className="h-[calc(100%-40px)] overflow-y-auto p-4 font-mono text-[11px] space-y-1 custom-scrollbar">
          {stats && (
            <div className="mb-4 rounded border border-blue-500/20 bg-blue-950/10 p-3">
              <div className="mb-2 text-[10px] font-bold uppercase tracking-widest text-blue-400">Run Diagnostics</div>
              <div className="grid grid-cols-2 gap-2 font-sans text-xs text-gray-300 xl:grid-cols-5">
                <div>
                  <div className="text-[10px] uppercase tracking-widest text-gray-500">Steps</div>
                  <div className="text-sm text-white">{stats.steps}</div>
                </div>
                <div>
                  <div className="text-[10px] uppercase tracking-widest text-gray-500">Tokens</div>
                  <div className="text-sm text-white">{stats.totalTokens}</div>
                </div>
                <div>
                  <div className="text-[10px] uppercase tracking-widest text-gray-500">Elapsed</div>
                  <div className="text-sm text-white">{formatElapsed(stats.elapsedMs)}</div>
                </div>
                <div>
                  <div className="text-[10px] uppercase tracking-widest text-gray-500">Current Node</div>
                  <div className="truncate text-sm text-white">{stats.currentNode || 'idle'}</div>
                </div>
                <div>
                  <div className="text-[10px] uppercase tracking-widest text-gray-500">Exit</div>
                  <div className="truncate text-sm text-white">{stats.exitReason || 'running'}</div>
                </div>
              </div>
              {Object.keys(stats.nodeVisits).length > 0 && (
                <div className="mt-3 text-xs text-gray-400">
                  Visits: {Object.entries(stats.nodeVisits)
                    .sort((a, b) => b[1] - a[1])
                    .map(([node, visits]) => `${node} x${visits}`)
                    .join(', ')}
                </div>
              )}
            </div>
          )}
          {response && (
            <div className="mb-4 rounded border border-green-500/20 bg-green-950/20 p-3">
              <div className="mb-2 text-[10px] font-bold uppercase tracking-widest text-green-400">Final Response</div>
              <div className="font-sans text-sm leading-6 text-gray-100 whitespace-pre-wrap break-words">{response}</div>
            </div>
          )}
          {logs.map((log, i) => (
            <div key={i} className="flex gap-3 animate-in fade-in duration-300">
              <span className="text-gray-600 shrink-0">[{log.timestamp.toLocaleTimeString()}]</span>
              <span className={`font-bold shrink-0 w-24 uppercase ${
                log.type === 'error' ? 'text-red-500' : 
                log.type === 'step' ? 'text-amber-400' :
                log.type === 'route' ? 'text-cyan-400' :
                log.type === 'done' ? 'text-green-500' : 'text-blue-400'
              }`}>
                {log.type}
              </span>
              <span className="text-gray-300 break-all">
                {typeof log.content === 'string' ? log.content : (log.content?.message || JSON.stringify(log.content))}
              </span>
            </div>
          ))}
          {logs.length === 0 && !response && (
            <div className="h-full flex items-center justify-center text-gray-600 italic">
              No execution data. Click "Deploy" to start.
            </div>
          )}
        </div>
      )}
    </div>
  );
};

const formatElapsed = (elapsedMs: number) => {
  if (elapsedMs < 1000) {
    return `${elapsedMs} ms`;
  }
  return `${(elapsedMs / 1000).toFixed(1)} s`;
};
