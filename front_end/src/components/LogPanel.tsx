import React, { useEffect, useRef } from 'react';
import { Terminal, ChevronUp, ChevronDown } from 'lucide-react';

interface LogEntry {
  type: string;
  content: any;
  timestamp: Date;
}

interface LogPanelProps {
  logs: LogEntry[];
  response: string | null;
  isOpen: boolean;
  onToggle: () => void;
  onClear: () => void;
}

export const LogPanel: React.FC<LogPanelProps> = ({ logs, response, isOpen, onToggle, onClear }) => {
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
