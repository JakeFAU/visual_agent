import React, { useState, ReactNode } from 'react';
import { ChevronDown, ChevronRight } from 'lucide-react';

interface AccordionProps {
  title: string;
  count?: number;
  children: ReactNode;
  defaultOpen?: boolean;
}

export const Accordion: React.FC<AccordionProps> = ({ title, count, children, defaultOpen = false }) => {
  const [isOpen, setIsOpen] = useState(defaultOpen);

  return (
    <div className="border-b border-gray-800">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full py-3 flex items-center justify-between text-left hover:bg-gray-800/50 transition-colors"
      >
        <div className="flex items-center gap-2">
          {isOpen ? <ChevronDown size={14} className="text-gray-500" /> : <ChevronRight size={14} className="text-gray-500" />}
          <span className="text-xs font-bold text-gray-400 uppercase tracking-wider">{title}</span>
          {count !== undefined && (
            <span className="bg-gray-800 text-blue-400 text-[10px] px-1.5 py-0.5 rounded-full font-mono">
              {count}
            </span>
          )}
        </div>
      </button>
      {isOpen && <div className="pb-4 space-y-4">{children}</div>}
    </div>
  );
};
