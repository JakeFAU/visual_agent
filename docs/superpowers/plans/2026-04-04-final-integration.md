# Final Front-End Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Connect the IDE to the backend for saving/executing graphs and implement a user-friendly structured output builder.

**Architecture:**
- `api/client.ts`: Handle HTTP requests and SSE streaming.
- `LogPanel.tsx`: Display execution events.
- `SchemaFieldBuilder.tsx`: Visual editor for JSON schemas.
- Integration: Update `App.tsx` and `SidePanel.tsx` to use the new components and API.

**Tech Stack:**
- React, Zustand, SSE, Tailwind CSS.

---

### Task 1: API Client

**Files:**
- Create: `visual_agent/front_end/src/api/client.ts`

- [ ] **Step 1: Implement `saveGraph` and `executeGraph`.**

```typescript
export const API_BASE = 'http://localhost:8080/api';

export const saveGraph = async (graph: any) => {
  const resp = await fetch(`${API_BASE}/graphs`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(graph),
  });
  return resp.json();
};

export const executeGraph = async (graph: any, input: string, onEvent: (ev: any) => void) => {
  const response = await fetch(`${API_BASE}/execute`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ graph, input }),
  });

  const reader = response.body?.getReader();
  if (!reader) return;

  const decoder = new TextDecoder();
  while (true) {
    const { value, done } = await reader.read();
    if (done) break;
    const chunk = decoder.decode(value);
    const lines = chunk.split('\n');
    for (const line of lines) {
        if (line.startsWith('data: ')) {
            try {
                const data = JSON.parse(line.substring(6));
                onEvent(data);
            } catch (e) { /* skip non-json */ }
        }
    }
  }
};
```

---

### Task 2: Log Panel

**Files:**
- Create: `visual_agent/front_end/src/components/LogPanel.tsx`

- [ ] **Step 1: Implement the scrolling log view.**

---

### Task 3: Structured Output Builder

**Files:**
- Create: `visual_agent/front_end/src/components/editors/SchemaFieldBuilder.tsx`

- [ ] **Step 1: Implement field editor that generates JSON schema.**

---

### Task 4: Final Wiring

**Files:**
- Modify: `visual_agent/front_end/src/App.tsx`
- Modify: `visual_agent/front_end/src/components/SidePanel.tsx`

- [ ] **Step 1: Wire "Save" and "Deploy" buttons.**
- [ ] **Step 2: Add LogPanel to layout.**
- [ ] **Step 3: Add SchemaFieldBuilder to LLM node config.**
