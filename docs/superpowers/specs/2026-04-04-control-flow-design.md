# Spec: Visual Agent Control Flow (If/Else & While)

**Status:** Draft
**Date:** 2026-04-04

## Goal
Implement support for deterministic conditional branching and iterative loops in the Go compiler and execution engine.

## Architecture

### 1. If/Else Logic (The Switch Agent)
Instead of a simple if/else, we will implement a polymorphic `SwitchAgent` that can handle multiple branches (True/False or eventually a full Switch).
- **Implementation:** A custom Go struct that implements `agent.Agent`.
- **Condition Evaluation:** 
  - **CEL:** Use `google/cel-go` to evaluate expressions against the current session state.
  - **JSONPath:** Use a standard Go JSONPath library.
- **Routing:** Upon evaluation, the agent returns an event with `Actions.TransferToAgent` set to the target node's ID.

### 2. While Logic (Loop Agent)
Map `while_node` directly to ADK's `loopagent`.
- **Structure:**
  - `LoopAgent` wraps a `SequentialAgent` containing the nodes in the loop body.
  - Termination is handled by a special `ConditionNode` at the end of the loop body that emits an `escalate=true` action if the condition is no longer met.

### 3. Intermediate Representation Updates
- **Compiler:** Add `logic_node.go` to handle `if_else_node` and `while_node` types.
- **DAG Walker:** Update to recognize branching paths and nested loop structures.

## Data Flow
1. **If/Else:** 
   - `SwitchAgent.Run` -> Evaluate Condition -> Emit `Transfer(TargetAgent)`.
2. **While:** 
   - `LoopAgent.Run` -> Execute Body -> `ConditionAgent` evaluates -> Repeat or Exit.

## Next Steps
1. Add `google/cel-go` to `go.mod`.
2. Implement `SwitchAgent` in `internal/compiler/logic_node.go`.
3. Implement `WhileNodeCompiler`.
4. Update `Compiler.Compile` to register these new node types.
