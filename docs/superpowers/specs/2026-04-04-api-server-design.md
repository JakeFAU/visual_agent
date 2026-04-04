# Spec: Visual Agent API Server & Execution Flow

**Status:** Draft
**Date:** 2026-04-04

## Goal
Implement a robust Go-based API server using Gin to manage graph persistence and provide a real-time execution environment for the Visual Agent IDE.

## Architecture

### 1. API Framework (Gin)
- **CORS:** Enabled for local development (defaulting to front-end port 5173).
- **Middleware:** JSON logging and recovery.

### 2. Persistence Layer
- **Storage:** Local file-based storage in `back_end/data/graphs/`.
- **Naming:** Files stored as `<graph_name>.json`.

### 3. Endpoints
- **GET `/api/graphs`**: List all saved graphs.
- **GET `/api/graphs/:name`**: Retrieve a specific graph.
- **POST `/api/graphs`**: Save or update a graph.
- **POST `/api/execute`**:
  - **Input:** `{ graph: JSON, input: string }`
  - **Process:** Compiles the graph on-the-fly and runs it using `runtime.LocalRuntime`.
  - **Output:** Streamed `session.Event` objects (SSE).

### 4. Real-time Streaming (SSE)
- Use `Content-Type: text/event-stream`.
- Stream ADK events (e.g., `model_request`, `tool_call`, `final_response`) directly to the front-end.

## Data Flow
1. **Save:** Front-end -> `POST /api/graphs` -> Go saves to disk.
2. **Execute:** Front-end -> `POST /api/execute` -> Go Compiles -> `LocalRuntime.Execute` -> Stream events back to front-end.

## Next Steps
1. Add `github.com/gin-gonic/gin` to `go.mod`.
2. Implement `internal/server/server.go`.
3. Implement `internal/storage/storage.go`.
4. Update `main.go` to launch the API server.
