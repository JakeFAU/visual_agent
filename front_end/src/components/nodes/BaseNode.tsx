import React, { ReactNode } from 'react';

interface BaseNodeProps {
  title: string;
  selected?: boolean;
  children: ReactNode;
  color?: string;
}

export const BaseNode: React.FC<BaseNodeProps> = ({ title, selected, children, color = 'blue' }) => {
  const borderColor = selected ? 'ring-2 ring-blue-500' : 'border-gray-700';
  
  // Mapping color names to Tailwind classes
  const colorMap: Record<string, string> = {
    blue: 'bg-blue-500',
    green: 'bg-green-500',
    amber: 'bg-amber-500',
    purple: 'bg-purple-500',
    gray: 'bg-gray-500',
  };

  const barColor = colorMap[color] || 'bg-blue-500';

  return (
    <div className={`bg-gray-800 rounded-md border ${borderColor} shadow-lg min-w-[180px] overflow-hidden`}>
      <div className={`h-1 w-full ${barColor}`} />
      <div className="p-3">
        <div className="text-[10px] font-bold text-gray-400 mb-2 uppercase tracking-wider">{title}</div>
        <div className="text-sm text-gray-200">
            {children}
        </div>
      </div>
    </div>
  );
};
