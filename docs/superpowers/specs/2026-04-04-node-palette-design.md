# Spec: UI Node Palette (Drag & Drop)

**Status:** Draft
**Date:** 2026-04-04

## Goal
Implement a sidebar palette that allows users to drag and drop new nodes onto the canvas to build workflows visually.

## Architecture

### 1. Palette Sidebar (`src/components/Palette.tsx`)
- **UI:** A slim vertical sidebar on the left.
- **Items:** Draggable blocks for each node type:
  - Input Node
  - LLM Agent
  - Toolbox
  - Output Node
  - If / Else
  - While / Loop
- **Logic:** Uses `onDragStart` to set the `application/reactflow` data type.

### 2. Canvas Integration (`App.tsx`)
- **ReactFlowProvider:** Wrapped around the App to enable `useReactFlow` hook.
- **Drop Handling:** 
  - `onDragOver`: Prevents default to allow drop.
  - `onDrop`: Uses `screenToFlowPosition` to get canvas coordinates and calls `addNode`.

### 3. Store Update (`src/store/useGraphStore.ts`)
- **New Action:** `addNode(type, position)`
- **Logic:** 
  - Generates a unique ID.
  - Sets default `config` based on the node type (e.g., default Gemini model for LLM nodes).
  - Appends to the `nodes` array.

## Data Flow
1. User drags item from Palette.
2. `onDragStart` sets node type in `dataTransfer`.
3. User drops item on React Flow canvas.
4. `onDrop` calculates position and triggers `addNode`.
5. Store updates -> Canvas re-renders with the new node.

## Next Steps
1. Add `addNode` to `useGraphStore.ts`.
2. Create `src/components/Palette.tsx`.
3. Wrap `App.tsx` in `ReactFlowProvider` and implement drop logic.
