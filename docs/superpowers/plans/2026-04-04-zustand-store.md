# Zustand Store Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the state management layer for the React front-end using Zustand to manage React Flow state and bidirectional JSON contract serialization.

**Architecture:**
- Use Zustand for a centralized store.
- Use React Flow types for nodes and edges.
- Use Zod (from `src/schema/graph.ts`) for strict serialization and validation.
- Implement specialized handlers for canvas interactions and contract enforcement.

**Tech Stack:**
- React, Zustand, React Flow, Zod, TypeScript.

---

### Task 1: Initialize Front-End Dependencies

**Files:**
- Create: `visual_agent/front_end/package.json`

- [ ] **Step 1: Create package.json and install dependencies.**

Run:
```bash
cd visual_agent/front_end
npm init -y
npm install zustand reactflow zod
```

---

### Task 2: Implement Zustand Store

**Files:**
- Create: `visual_agent/front_end/src/store/useGraphStore.ts`

- [ ] **Step 1: Define the Store state and actions interface.**

```typescript
import { create } from 'zustand';
import {
  Connection,
  Edge,
  EdgeChange,
  Node,
  NodeChange,
  addEdge,
  applyNodeChanges,
  applyEdgeChanges,
  OnNodesChange,
  OnEdgesChange,
  OnConnect,
} from 'reactflow';
import { GraphSchema, Graph, Node as ContractNode, Edge as ContractEdge } from '../schema/graph';

export interface GraphState {
  nodes: Node[];
  edges: Edge[];
  version: string;
  name: string;
  validationErrors: string[];
  selectedNodeId: string | null;

  // React Flow Handlers
  onNodesChange: OnNodesChange;
  onEdgesChange: OnEdgesChange;
  onConnect: OnConnect;

  // Selection
  setSelectedNodeId: (nodeId: string | null) => void;

  // Contract Serialization
  exportGraph: () => Graph;
  importGraph: (json: string) => void;
  validateGraph: () => void;
}
```

- [ ] **Step 2: Implement the store logic.**

```typescript
export const useGraphStore = create<GraphState>((set, get) => ({
  nodes: [],
  edges: [],
  version: "1.0",
  name: "New Workflow",
  validationErrors: [],
  selectedNodeId: null,

  onNodesChange: (changes: NodeChange[]) => {
    set({
      nodes: applyNodeChanges(changes, get().nodes),
    });
    // Track selection
    const selectionChange = changes.find(c => c.type === 'select');
    if (selectionChange && 'selected' in selectionChange) {
        if (selectionChange.selected) {
            set({ selectedNodeId: selectionChange.id });
        } else if (get().selectedNodeId === selectionChange.id) {
            set({ selectedNodeId: null });
        }
    }
  },

  onEdgesChange: (changes: EdgeChange[]) => {
    set({
      edges: applyEdgeChanges(changes, get().edges),
    });
  },

  onConnect: (connection: Connection) => {
    // Validate types before connecting
    if (connection.sourceHandle !== connection.targetHandle) {
        console.warn("Invalid connection: types must match", connection.sourceHandle, connection.targetHandle);
        return;
    }
    set({
      edges: addEdge(connection, get().edges),
    });
  },

  setSelectedNodeId: (nodeId) => set({ selectedNodeId: nodeId }),

  exportGraph: () => {
    const { nodes, edges, version, name } = get();
    
    const contractNodes: ContractNode[] = nodes.map(node => ({
      id: node.id,
      type: node.type as any, // Validated by Zod later
      position: node.position,
      config: node.data.config,
    }));

    const contractEdges: ContractEdge[] = edges.map(edge => ({
      id: edge.id,
      source: edge.source,
      source_port: edge.sourceHandle || "",
      target: edge.target,
      target_port: edge.targetHandle || "",
      data_type: edge.sourceHandle || "", // Assuming handle ID matches data_type
      edge_kind: "data_flow",
    }));

    return {
      version,
      name,
      nodes: contractNodes,
      edges: contractEdges,
    };
  },

  importGraph: (json: string) => {
    try {
      const data = JSON.parse(json);
      const validated = GraphSchema.parse(data);
      
      const rfNodes: Node[] = validated.nodes.map(node => ({
        id: node.id,
        type: node.type,
        position: node.position,
        data: { config: node.config },
      }));

      const rfEdges: Edge[] = validated.edges.map(edge => ({
        id: edge.id,
        source: edge.source,
        sourceHandle: edge.source_port,
        target: edge.target,
        targetHandle: edge.target_port,
      }));

      set({
        nodes: rfNodes,
        edges: rfEdges,
        version: validated.version,
        name: validated.name,
        validationErrors: [],
      });
    } catch (err: any) {
      console.error("Import failed:", err);
      if (err.errors) {
        set({ validationErrors: err.errors.map((e: any) => `${e.path.join('.')}: ${e.message}`) });
      }
    }
  },

  validateGraph: () => {
    const graph = get().exportGraph();
    const result = GraphSchema.safeParse(graph);
    if (!result.success) {
      set({ 
        validationErrors: result.error.errors.map(e => `${e.path.join('.')}: ${e.message}`) 
      });
    } else {
      set({ validationErrors: [] });
    }
  },
}));
```

---

### Task 3: Verification

- [ ] **Step 1: Create a simple test file to verify serialization logic.**

```typescript
import { useGraphStore } from './useGraphStore';

// Mock some state
useGraphStore.setState({
    nodes: [
        { id: '1', type: 'input_node', position: { x: 0, y: 0 }, data: { config: { name: 'test', description: 'test' } } }
    ],
    edges: [],
    name: 'Test Workflow',
    version: '1.0'
});

const exported = useGraphStore.getState().exportGraph();
console.log("Exported:", JSON.stringify(exported, null, 2));

useGraphStore.getState().validateGraph();
console.log("Validation errors:", useGraphStore.getState().validationErrors);
```

Run with `npx tsx src/store/test_store.ts` (after setting up the test file).
