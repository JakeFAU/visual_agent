# Spec: Abstract Agent Runtime

**Status:** Draft
**Date:** 2026-04-04

## Goal
Implement a provider-agnostic execution engine for Visual Agent agents that leverages Google ADK's `Runner` architecture while remaining decoupled from specific backends like Vertex AI.

## Architecture

### 1. Core Runtime Interface
The `AgentRuntime` interface defines the contract for executing agents.
```go
type AgentRuntime interface {
    Execute(ctx context.Context, agent agent.Agent, input string) iter.Seq2[*session.Event, error]
}
```

### 2. Base ADK Runner
A shared implementation that wraps `google.golang.org/adk/runner`.
- **Initialization:** Takes an `agent.Agent` and a `session.Service`.
- **Execution:** Handles standard ADK session creation, event streaming, and error handling.

### 3. Concrete Implementations
- **VertexRuntime:** Injects `google.golang.org/adk/session/vertexai.Service`. Uses GCP credentials (ADC).
- **LocalRuntime:** Injects `google.golang.org/adk/session.InMemoryService()`. Ideal for local development and unit testing.

### 4. Configuration (Viper)
- Runtime selection is controlled via `VISUAL_AGENT_RUNTIME_TYPE` (e.g., `vertex`, `local`).
- Backend-specific configs (ProjectID, Location) are loaded conditionally.

## Data Flow
1. **Request:** Compiled `agent.Agent` + user `input`.
2. **Setup:** Runtime ensures a session exists (ADK `AutoCreateSession`).
3. **Stream:** `Runner.Run()` produces an iterator of `session.Event` objects.
4. **Completion:** Results are returned to the caller (e.g., the CLI or a web response).

## Next Steps
1. Implement `BaseRuntime` in `internal/runtime`.
2. Implement `VertexRuntime` factory.
3. Implement `LocalRuntime` factory.
4. Update `cmd/visual-agent` to support an `execute` command.
