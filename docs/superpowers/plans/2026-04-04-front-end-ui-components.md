# Visual Agent UI Components Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the core UI components for the Visual Agent IDE, including custom nodes, side panel, and main application layout.

**Architecture:**
- Shared `BaseNode` component for common styling and selection.
- Specialized node components (`InputNode`, `LLMNode`, etc.) using React Flow's `Handle`.
- Dynamic `SidePanel` that renders forms based on `selectedNodeId`.
- Tailwind CSS for styling.

**Tech Stack:**
- React, React Flow, Zustand, Tailwind CSS, Lucide React (icons).

---

### Task 1: Update Store with Node Mutation

**Files:**
- Modify: `visual_agent/front_end/src/store/useGraphStore.ts`

- [ ] **Step 1: Add `updateNodeConfig` action.**

```typescript
// Add to GraphState interface:
// updateNodeConfig: (nodeId: string, config: any) => void;

// Add to store implementation:
  updateNodeConfig: (nodeId, config) => {
    set({
      nodes: get().nodes.map((node) =>
        node.id === nodeId ? { ...node, data: { ...node.data, config } } : node
      ),
    });
  },
```

---

### Task 2: Shared Node Components

**Files:**
- Create: `visual_agent/front_end/src/components/nodes/BaseNode.tsx`

- [ ] **Step 1: Implement `BaseNode` wrapper.**

```tsx
import React, { ReactNode } from 'react';

interface BaseNodeProps {
  title: string;
  selected?: boolean;
  children: ReactNode;
  color?: string;
}

export const BaseNode: React.FC<BaseNodeProps> = ({ title, selected, children, color = 'blue' }) => {
  const borderColor = selected ? 'ring-2 ring-blue-500' : 'border-gray-700';
  
  return (
    <div className={`bg-gray-800 rounded-md border ${borderColor} shadow-lg min-w-[180px] overflow-hidden`}>
      <div className={`h-1 w-full bg-${color}-500`} />
      <div className="p-3">
        <div className="text-xs font-semibold text-gray-400 mb-2 uppercase tracking-wider">{title}</div>
        {children}
      </div>
    </div>
  );
};
```

---

### Task 3: Specialized Node Components

**Files:**
- Create: `visual_agent/front_end/src/components/nodes/InputNode.tsx`
- Create: `visual_agent/front_end/src/components/nodes/LLMNode.tsx`
- Create: `visual_agent/front_end/src/components/nodes/ToolboxNode.tsx`

- [ ] **Step 1: Implement `InputNode`.**
- [ ] **Step 2: Implement `LLMNode`.**
- [ ] **Step 3: Implement `ToolboxNode`.**

---

### Task 4: Side Panel

**Files:**
- Create: `visual_agent/front_end/src/components/SidePanel.tsx`

- [ ] **Step 1: Implement dynamic configuration form.**

---

### Task 5: App Shell

**Files:**
- Create: `visual_agent/front_end/src/App.tsx`
- Create: `visual_agent/front_end/src/main.tsx`

- [ ] **Step 1: Wire up React Flow with Custom Nodes.**
- [ ] **Step 2: Add Header and Side Panel layout.**
