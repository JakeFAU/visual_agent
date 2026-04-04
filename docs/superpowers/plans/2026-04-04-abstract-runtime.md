# Abstract Agent Runtime Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a provider-agnostic execution engine for Visual Agent using Google ADK's `Runner` architecture.

**Architecture:**
- Refine `AgentRuntime` interface to use `agent.Agent`.
- Implement `BaseRuntime` that wraps ADK `Runner`.
- Provide `VertexRuntime` and `LocalRuntime` factories that inject different `session.Service` implementations.
- Update configuration to support runtime selection.

**Tech Stack:**
- Go 1.25+
- Google ADK

---

### Task 1: Refine Interface & Base Implementation

**Files:**
- Modify: `visual_agent/back_end/internal/runtime/runtime.go`

- [ ] **Step 1: Update interface and implement `BaseRuntime`.**

```go
package runtime

import (
	"context"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"iter"
    "google.golang.org/genai"
)

type AgentRuntime interface {
	Execute(ctx context.Context, a agent.Agent, input string) iter.Seq2[*session.Event, error]
}

type BaseRuntime struct {
	sessionService session.Service
}

func (r *BaseRuntime) Execute(ctx context.Context, a agent.Agent, input string) iter.Seq2[*session.Event, error] {
	adkRunner, err := runner.New(runner.Config{
		AppName:        "VisualAgent",
		Agent:          a,
		SessionService: r.sessionService,
		AutoCreateSession: true,
	})
    if err != nil {
        return func(yield func(*session.Event, error) bool) {
            yield(nil, err)
        }
    }

	return adkRunner.Run(ctx, "default-user", "default-session", &genai.Content{
        Parts: []*genai.Part{{Text: input}},
    }, agent.RunConfig{})
}
```

---

### Task 2: Concrete Runtime Factories

**Files:**
- Modify: `visual_agent/back_end/internal/runtime/runtime.go`

- [ ] **Step 1: Implement `NewVertexRuntime` and `NewLocalRuntime`.**

---

### Task 3: Update Config & CLI

**Files:**
- Modify: `visual_agent/back_end/internal/config/config.go`
- Modify: `visual_agent/back_end/cmd/visual-agent/main.go`

- [ ] **Step 1: Add `RuntimeType` to `Config`.**
- [ ] **Step 2: Update CLI to initialize the correct runtime and add an `execute` command.**

---

### Task 4: Verification

- [ ] **Step 1: Create a unit test for `LocalRuntime` using `InMemoryService`.**
