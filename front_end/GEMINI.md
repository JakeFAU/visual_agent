# Front-End Context: Visual Agent UI

## Tech Stack

- **Framework:** React + Vite (TypeScript).
- **Canvas:** React Flow.
- **State Management:** Zustand.
- **Schema Validation:** Zod (for strict JSON bidirectional validation).
- **Styling:** Tailwind CSS.

## Architecture Guidelines

1. **Zustand Store:** The single source of truth. It holds `nodes` and `edges`. All updates to the canvas (dragging, connecting, updating configs in the side panel) must dispatch through the store.
2. **Custom Nodes:** Every node type (`input_node`, `llm_node`, `output_node`, `toolbox`) requires a custom React Flow node component.
3. **Strict Port Connectivity:** Use React Flow's `isValidConnection` callback. A connection is only valid if `sourceHandle` type matches `targetHandle` type (e.g., `message` to `message`, `tool_result` to `tool_result`). Map these types to specific CSS variables for consistent port coloring.
4. **Bidirectional Serialization:** - `exportGraph()`: Parses the Zustand store into the canonical JSON schema, stripping out React Flow UI specifics (like `selected`, `dragging`) while retaining `position`.
   - `importGraph(json)`: Validates the incoming JSON via Zod, maps it back into React Flow compatible `Node` and `Edge` arrays, and updates the Zustand store.
5. **Side Panel:** When `onNodesChange` detects a selection, the side panel renders a dynamic form based on the selected node's `type` and mutates the node's `config` object.
