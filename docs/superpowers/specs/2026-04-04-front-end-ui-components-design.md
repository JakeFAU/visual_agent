# Spec: Visual Agent UI Components

**Status:** Draft
**Date:** 2026-04-04

## Goal

Implement the core UI components for the Visual Agent IDE, including custom React Flow nodes and a dynamic configuration side panel.

## Architecture

### 1. Custom React Flow Nodes

Each node type defined in the JSON contract will have a corresponding React component.

- **Node Container:** A shared wrapper for styling (Tailwind CSS), selection state, and header icons.
- **Port Typing (Visual):**
  - `message`: Green Handle (`#4caf50`)
  - `toolbox_handle`: Amber Handle (`#ffc107`)
  - `structured_output`: Purple Handle (`#9c27b0`)
  - `any`: Gray Handle (`#666`)

### 2. Side Panel (Configuration Form)

A dynamic panel that renders based on the `selectedNodeId` from the Zustand store.

- **Selection Binding:** `const selectedNode = useGraphStore(s => s.nodes.find(n => n.id === s.selectedNodeId))`.
- **Form Generation:**
  - `input_node`: Inputs for `name`, `description`.
  - `llm_node`: Inputs for `model`, `instruction`, `response_mode`, and sliders for `temperature`.
  - `toolbox`: Dynamic list editor for MCP servers and Custom Functions.
  - `if_else_node` / `while_node`: Code editor (simple textarea) for CEL/JSONPath conditions.
- **Live Update:** Changes in the form dispatch immediately to the Zustand store via a `updateNodeConfig(id, config)` action.

### 3. App Layout

- **Header:** Title, "Import" button (file picker), "Export" button (JSON download).
- **Canvas:** Full-bleed React Flow component.
- **Side Panel:** Fixed-width (320px) right sidebar.

## Data Flow

1. User clicks node -> `onNodesChange` (selection) -> `selectedNodeId` updated in Store.
2. Side Panel reads `selectedNodeId` -> Renders corresponding form.
3. User types in input -> `updateNodeConfig` -> Store updates -> Node on canvas re-renders.
4. User clicks "Export" -> `exportGraph()` -> Browser downloads JSON.

## Next Steps

1. Create `front_end/src/components/nodes/` for custom node types.
2. Create `front_end/src/components/SidePanel.tsx`.
3. Create `front_end/src/App.tsx` and `front_end/src/main.tsx` to mount the IDE.
