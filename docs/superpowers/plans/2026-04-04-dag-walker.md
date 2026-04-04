# DAG Walker Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a sophisticated DAG analysis and ADK orchestration builder in the Go compiler.

**Architecture:**
- Use Kahn's algorithm for topological sorting and cycle detection.
- Analyze edges to determine input/output dependencies between nodes.
- Map linear chains to `SequentialAgent`.
- Implement state injection using ADK's shared state and `{placeholder}` syntax.

**Tech Stack:**
- Go 1.25+
- Google ADK

---

### Task 1: Topological Sort & Validation

**Files:**
- Modify: `visual_agent/back_end/internal/compiler/compiler.go`

- [ ] **Step 1: Implement dependency analysis and topological sort.**

```go
func (c *Compiler) sortNodes(g graph.Graph) ([]graph.Node, error) {
    // 1. Build adjacency list
    // 2. Count in-degrees
    // 3. Kahn's algorithm
    // 4. Return sorted nodes or error if cycle detected
}
```

---

### Task 2: Edge Analysis & State Mapping

**Files:**
- Modify: `visual_agent/back_end/internal/compiler/compiler.go`

- [ ] **Step 1: Map output ports to state keys.**

---

### Task 3: Pattern Detection (Sequential)

**Files:**
- Modify: `visual_agent/back_end/internal/compiler/compiler.go`

- [ ] **Step 1: Group sorted nodes into a `SequentialAgent` if they form a chain.**

```go
// Example logic:
// If node A -> node B, wrap them in SequentialAgent.New("workflow").WithAgents(agentA, agentB)
```

---

### Task 4: Node Compiler Enhancements

**Files:**
- Modify: `visual_agent/back_end/internal/compiler/llm_node.go`

- [ ] **Step 1: Update `LLMNodeCompiler` to accept output key information from the walker.**

---

### Task 5: Verification

- [ ] **Step 1: Create a test graph with 3 sequential nodes and verify it compiles into a SequentialAgent.**
