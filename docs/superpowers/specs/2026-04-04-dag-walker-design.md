# Spec: Visual Agent DAG Walker & Pattern Construction

**Status:** Draft
**Date:** 2026-04-04

## Goal
Implement a sophisticated compiler phase that analyzes the Visual Agent Graph and constructs a hierarchy of optimized Google ADK orchestrators (`SequentialAgent`, `ParallelAgent`, `LoopAgent`).

## Architecture

### 1. Topological Sorting & Validation
- **Validation:** Ensure the graph is a Directed Acyclic Graph (DAG), except for explicit `while_node` loops.
- **Sorting:** Determine the base execution order using Kahn's algorithm or DFS.

### 2. Pattern Detection Logic
The walker will group nodes into execution "blocks":
- **Linear Chain:** A series of nodes where each has exactly one dependency (the previous node). Map to `SequentialAgent`.
- **Parallel Fan-out:** Multiple independent nodes branching from a single source. Map to `ParallelAgent`.
- **Conditional Branch:** Nodes connected to `if_else_node`. Map to a custom "Router" agent or ADK transfer logic.
- **Iterative Loop:** Cycles managed by `while_node`. Map to `LoopAgent`.

### 3. State & Data Flow Mapping
- **Input Mapping:** Map edge `target_port` to ADK state placeholders (e.g., `{in_message}`).
- **Output Mapping:** Map edge `source_port` to ADK `WithOutputKey` declarations.
- **Typed Connections:** Ensure data types match during the compilation phase.

## Data Structures
- `TaskGraph`: Intermediate representation of the sorted nodes and their dependencies.
- `ADKFactory`: Produces the actual ADK agent instances from the `TaskGraph`.

## Next Steps
1. Implement topological sort in `internal/compiler/compiler.go`.
2. Implement pattern detection logic.
3. Update `LLMNodeCompiler` to support `WithOutputKey` based on edges.
