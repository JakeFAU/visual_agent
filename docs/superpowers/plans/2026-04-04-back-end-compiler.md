# Visual Agent Back-End Compiler Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a robust Go-based compiler and execution engine for Visual Agent graphs, using Google-quality code patterns, Viper for config, and Cobra for the CLI.

**Architecture:**
- `AgentRuntime` interface to abstract Vertex AI details.
- Polymorphic `NodeCompiler` pattern for translating graph nodes to ADK declarations.
- Dependency injection for runtime and configuration.
- Cobra-based CLI for running and validating graphs.

**Tech Stack:**
- Go 1.25+
- Google ADK (Agent Development Kit)
- Cobra (CLI), Viper (Config)

---

### Task 1: Initialize Dependencies

**Files:**
- Modify: `visual_agent/back_end/go.mod`

- [ ] **Step 1: Add necessary dependencies.**

Run:
```bash
cd visual_agent/back_end
go get github.com/spf13/cobra
go get github.com/spf13/viper
# Note: Assuming Google ADK import paths based on research
go get google.golang.org/adk
```

---

### Task 2: Runtime Abstraction

**Files:**
- Create: `visual_agent/back_end/internal/runtime/runtime.go`

- [ ] **Step 1: Define the `AgentRuntime` interface.**

```go
package runtime

import (
	"context"
	"google.golang.org/adk/session"
	"iter"
)

// AgentRuntime abstracts the underlying LLM provider (e.g. Vertex AI)
type AgentRuntime interface {
	// Execute runs a compiled agent with the given input
	Execute(ctx context.Context, agent interface{}, input string) iter.Seq2[*session.Event, error]
}
```

---

### Task 3: Compiler Core & Interfaces

**Files:**
- Create: `visual_agent/back_end/internal/compiler/compiler.go`

- [ ] **Step 1: Define compiler types and interfaces.**

```go
package compiler

import (
	"github.com/JakeFAU/visual_agent/internal/graph"
)

// NodeCompiler translates a specific graph node into its ADK representation
type NodeCompiler interface {
	Compile(node graph.Node) (interface{}, error)
}

// Compiler orchestrates the translation of a full Graph
type Compiler struct {
	compilers map[string]NodeCompiler
}

func New() *Compiler {
	return &Compiler{
		compilers: make(map[string]NodeCompiler),
	}
}

func (c *Compiler) Register(nodeType string, nc NodeCompiler) {
	c.compilers[nodeType] = nc
}
```

---

### Task 4: Node Compilers Implementation

**Files:**
- Create: `visual_agent/back_end/internal/compiler/llm_node.go`
- Create: `visual_agent/back_end/internal/compiler/toolbox_node.go`

- [ ] **Step 1: Implement `LLMNodeCompiler`.**
- [ ] **Step 2: Implement `ToolboxNodeCompiler`.**

---

### Task 5: CLI & Config Scaffolding

**Files:**
- Create: `visual_agent/back_end/cmd/visual-agent/main.go`
- Create: `visual_agent/back_end/internal/config/config.go`

- [ ] **Step 1: Implement Viper-based config loader.**
- [ ] **Step 2: Implement Cobra "run" command to compile and execute a JSON file.**

```go
// Example CLI logic:
// 1. Load JSON graph
// 2. compiler.New().Compile(graph)
// 3. runtime.NewVertexRuntime().Execute(compiled, input)
```
