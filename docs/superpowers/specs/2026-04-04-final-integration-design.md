# Spec: Final Front-End Integration & Schema Builder

**Status:** Draft
**Date:** 2026-04-04

## Goal
Connect the Visual Agent IDE to the Go backend for persistence and execution, and provide a user-friendly way to define structured outputs.

## Architecture

### 1. API Client (`src/api/client.ts`)
- **Functions:** `saveGraph(graph)`, `loadGraphs()`, `executeGraph(graph, input)`.
- **Execution:** Uses `fetch` with the `EventSource` pattern (or manual stream reading) to handle SSE from `/api/execute`.

### 2. Execution Logs Panel
- **UI:** A collapsible bottom-docked panel in `App.tsx`.
- **Content:** List of formatted events:
  - `[INFO] Compiling graph...`
  - `[LLM] Instruction: ...`
  - `[DONE] Execution complete.`

### 3. Structured Output Builder
- **UI:** An editable list of "Fields" in the `LLMNode` section of `SidePanel.tsx`.
- **Fields:** `name`, `type` (string, number, boolean, object), `description`.
- **Internal:** Automatically generates the standard JSON Schema `Draft 7` required by the ADK.

### 4. Wiring
- **Header:** "Save" button (calls `saveGraph`), "Export & Deploy" button (opens a prompt for `input`, then calls `executeGraph`).

## Data Flow
1. User clicks "Deploy" -> Prompt for `input`.
2. `executeGraph` sends POST to Go Backend.
3. SSE stream received -> Log Panel appends events.
4. Execution finishes -> Final grounded response displayed.

## Next Steps
1. Create `front_end/src/api/client.ts`.
2. Create `front_end/src/components/LogPanel.tsx`.
3. Create `front_end/src/components/editors/SchemaFieldBuilder.tsx`.
4. Update `SidePanel.tsx` and `App.tsx` to wire everything together.
