# Spec: Visual Agent Back-End Compiler

**Status:** Draft
**Date:** 2026-04-04

## Goal

Implement a robust Go-based compiler and execution engine that translates the Visual Agent Graph JSON into executable Google ADK Agents.

## Architecture

### 1. Compiler (The Walker)

The compiler's responsibility is to validate the graph and translate it into a runnable structure.

- **Node Compilation:** Use a polymorphic strategy where each node `type` is mapped to a specific compiler logic.
- **DAG Construction:** Validate that the graph is a Directed Acyclic Graph (DAG).
- **ADK Mapping:**
  - `llm_node` -> `llmagent.New(...)`
  - `toolbox` -> `mcptoolset.New...` or `functiontool.New(...)`
  - `if_else_node` -> Deterministic Go logic or a custom "router" agent.
  - `while_node` -> `loopagent.New(...)`

### 2. Execution Engine

A lightweight runtime that manages the lifecycle of an agent invocation.

- **Interface:** `AgentRuntime` to abstract Vertex AI / Google Cloud details.
- **Input/Output:** Handles injecting user messages into the start node and collecting results from `output_node`s.
- **State Management:** Leverages ADK's session state to pass data between nodes.

### 3. Package Structure (Google Best Practices)

- `internal/graph`: JSON schemas and unmarshaling logic (already started).
- `internal/compiler`: The graph walker and ADK factory.
- `internal/runtime`: Vertex AI client and execution loop.
- `pkg/api`: Public types for external consumption.

## Data Flow

1. **Unmarshal:** JSON -> `graph.Graph` (Go Structs).
2. **Compile:** `graph.Graph` -> `compiler.CompiledAgent` (DAG of ADK Agents).
3. **Execute:** `CompiledAgent.Run(ctx, input)` -> `iter.Seq2[*session.Event, error]`.

## Next Steps

1. Implement the `compiler` package.
2. Implement node-to-ADK mapping for the core 6 node types.
3. Build a simple CLI or test harness to run a compiled graph.
