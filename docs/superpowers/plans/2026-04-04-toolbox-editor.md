# Toolbox Detail Editor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the inline editors for the Toolbox node configuration in the right side panel.

**Architecture:**
- Create modular editor components for Tools, MCP Servers, and Custom Functions.
- Use a shared `Accordion` component to organize sections.
- All editors perform live updates to the Zustand store.

**Tech Stack:**
- React, Zustand, Tailwind CSS, Lucide React (icons).

---

### Task 1: Shared UI Components

**Files:**
- Create: `visual_agent/front_end/src/components/ui/Accordion.tsx`

- [ ] **Step 1: Implement the Accordion component.**

```tsx
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
```

---

### Task 2: Built-in Tools Editor

**Files:**
- Create: `visual_agent/front_end/src/components/editors/ToolListEditor.tsx`

- [ ] **Step 1: Implement the tool toggle list.**

---

### Task 3: MCP Servers Editor

**Files:**
- Create: `visual_agent/front_end/src/components/editors/MCPServerEditor.tsx`

- [ ] **Step 1: Implement list of MCP configurations with add/remove actions.**

---

### Task 4: Custom Functions Editor

**Files:**
- Create: `visual_agent/front_end/src/components/editors/CustomFunctionEditor.tsx`

- [ ] **Step 1: Implement function signature editor with JSON schema validation for parameters.**

---

### Task 5: Integration

**Files:**
- Modify: `visual_agent/front_end/src/components/SidePanel.tsx`

- [ ] **Step 1: Integrate all editors into the Toolbox configuration section.**
