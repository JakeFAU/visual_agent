# UI Node Palette Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a draggable sidebar palette to allow users to build agent workflows visually.

**Architecture:**
- Update Zustand store with an `addNode` action that handles unique ID generation and default configs.
- Create a `Palette` component with draggable node type items.
- Wrap `App.tsx` in `ReactFlowProvider` to enable `useReactFlow` for coordinate mapping.
- Implement `onDrop` and `onDragOver` handlers on the React Flow canvas.

**Tech Stack:**
- React, React Flow, Zustand, Tailwind CSS, Lucide React.

---

### Task 1: Update Store with Node Creation

**Files:**
- Modify: `visual_agent/front_end/src/store/useGraphStore.ts`

- [ ] **Step 1: Add `addNode` action.**

```typescript
// Add to GraphState interface:
// addNode: (type: string, position: { x: number, y: number }) => void;

// Add to store implementation:
  addNode: (type, position) => {
    const id = `${type}-${Date.now()}`;
    let config: any = {};

    switch (type) {
      case 'input_node':
        config = { name: 'user_input', description: 'The user input' };
        break;
      case 'llm_node':
        config = { 
            name: 'llm_agent', 
            description: 'The Agent', 
            model: 'gemini-2.5-flash', 
            instruction: 'You are a helpful assistant.',
            response_mode: 'text',
            generate_content_config: { temperature: 0.7, max_output_tokens: 4096 }
        };
        break;
      case 'toolbox':
        config = { tools: [], mcp_servers: [], custom_functions: [] };
        break;
      case 'output_node':
        config = { name: 'final_output', output_key: 'result', format: 'message' };
        break;
      case 'if_else_node':
        config = { condition_language: 'CEL', condition: 'state.category == "billing"' };
        break;
      case 'while_node':
        config = { condition: 'state.counter < 5', max_iterations: 10 };
        break;
    }

    const newNode: Node = {
      id,
      type,
      position,
      data: { config },
    };

    set({
      nodes: [...get().nodes, newNode],
    });
  },
```

---

### Task 2: Palette Component

**Files:**
- Create: `visual_agent/front_end/src/components/Palette.tsx`

- [ ] **Step 1: Implement draggable node items.**

```tsx
import React from 'react';
import { MessageSquare, Cpu, Briefcase, Play, Split, Repeat } from 'lucide-react';

const NODE_TYPES = [
  { type: 'input_node', label: 'Input', icon: <MessageSquare size={16} />, color: 'bg-green-500' },
  { type: 'llm_node', label: 'LLM Agent', icon: <Cpu size={16} />, color: 'bg-blue-500' },
  { type: 'toolbox', label: 'Toolbox', icon: <Briefcase size={16} />, color: 'bg-amber-500' },
  { type: 'output_node', label: 'Output', icon: <Play size={16} />, color: 'bg-gray-500' },
  { type: 'if_else_node', label: 'If / Else', icon: <Split size={16} />, color: 'bg-purple-500' },
  { type: 'while_node', label: 'While', icon: <Repeat size={16} />, color: 'bg-purple-500' },
];

export const Palette: React.FC = () => {
  const onDragStart = (event: React.DragEvent, nodeType: string) => {
    event.dataTransfer.setData('application/reactflow', nodeType);
    event.dataTransfer.effectAllowed = 'move';
  };

  return (
    <aside className="w-16 bg-gray-900 border-r border-gray-800 flex flex-col items-center py-4 gap-4 shrink-0">
      <div className="text-[8px] font-bold text-gray-600 uppercase tracking-tighter mb-2">Nodes</div>
      {NODE_TYPES.map((node) => (
        <div
          key={node.type}
          className={`w-10 h-10 rounded-lg ${node.color} flex items-center justify-center cursor-grab active:cursor-grabbing shadow-lg hover:brightness-110 transition-all`}
          onDragStart={(event) => onDragStart(event, node.type)}
          draggable
          title={node.label}
        >
          {node.icon}
        </div>
      ))}
    </aside>
  );
};
```

---

### Task 3: Final App Integration

**Files:**
- Modify: `visual_agent/front_end/src/App.tsx`
- Modify: `visual_agent/front_end/src/main.tsx`

- [ ] **Step 1: Wrap `App` in `ReactFlowProvider` in `main.tsx`.**
- [ ] **Step 2: Add `Palette` to `App.tsx` layout.**
- [ ] **Step 3: Implement `onDrop` and `onDragOver` in `App.tsx`.**
