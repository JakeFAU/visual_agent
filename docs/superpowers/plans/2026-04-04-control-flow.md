# Control Flow (If/Else & While) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement deterministic conditional branching (`if_else_node`) and iterative loops (`while_node`) using custom ADK agents.

**Architecture:**
- Implement a custom `SwitchAgent` that evaluates CEL conditions and triggers agent transfers.
- Implement `IfElseNodeCompiler` to produce a `SwitchAgent`.
- Implement `WhileNodeCompiler` using ADK's `LoopAgent`.
- Update DAG walker to handle branching and loops.

**Tech Stack:**
- Go 1.25+
- Google ADK, CEL-Go

---

### Task 1: Initialize Dependencies

**Files:**
- Modify: `visual_agent/back_end/go.mod`

- [ ] **Step 1: Add `google/cel-go` dependency.**

Run:
```bash
cd visual_agent/back_end
go get github.com/google/cel-go/cel
```

---

### Task 2: Implement Switch Agent

**Files:**
- Create: `visual_agent/back_end/internal/compiler/logic_node.go`

- [ ] **Step 1: Implement `SwitchAgent` struct and `Run` method.**

```go
package compiler

import (
	"context"
	"fmt"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/session"
	"iter"
)

type SwitchAgent struct {
	agent.Config
	Condition string
	Language  string // "CEL" or "JSONPath"
	TrueAgent  string
	FalseAgent string
}

func (a *SwitchAgent) Run(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
	return func(yield func(*session.Event, error) bool) {
		// 1. Get session state
		state := ctx.Session().State()
		
		// 2. Evaluate condition (simplified for now)
		// result, err := evaluate(a.Condition, state)
		
		target := a.FalseAgent
		// if result { target = a.TrueAgent }

		// 3. Emit transfer event
		yield(&session.Event{
			Type: "transfer",
			Actions: &session.EventActions{
				TransferToAgent: target,
			},
		}, nil)
	}
}
```

---

### Task 3: Node Compilers for Logic

**Files:**
- Modify: `visual_agent/back_end/internal/compiler/logic_node.go`

- [ ] **Step 1: Implement `IfElseNodeCompiler`.**
- [ ] **Step 2: Implement `WhileNodeCompiler`.**

---

### Task 4: Register New Node Types

**Files:**
- Modify: `visual_agent/back_end/cmd/visual-agent/main.go`

- [ ] **Step 1: Register `if_else_node` and `while_node` in the compiler.**

---

### Task 5: Verification

- [ ] **Step 1: Create a test graph with an `if_else_node` and verify it transfers to the correct branch.**
