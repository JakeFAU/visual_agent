# API Server & Execution Flow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a Gin-based API server for graph persistence and real-time execution via SSE.

**Architecture:**
- `internal/storage`: Local file-based JSON storage for graphs.
- `internal/server`: Gin server with endpoints for CRUD on graphs and a POST `/execute` endpoint.
- Use SSE (Server-Sent Events) to stream ADK events to the front-end.
- Enable CORS for local development.

**Tech Stack:**
- Go 1.25+
- Gin, Google ADK

---

### Task 1: Initialize Dependencies

**Files:**
- Modify: `visual_agent/back_end/go.mod`

- [ ] **Step 1: Add `gin` and `cors` dependencies.**

Run:
```bash
cd visual_agent/back_end
go get github.com/gin-gonic/gin
go get github.com/gin-contrib/cors
```

---

### Task 2: Implement Persistence Layer

**Files:**
- Create: `visual_agent/back_end/internal/storage/storage.go`

- [ ] **Step 1: Implement `SaveGraph`, `LoadGraph`, and `ListGraphs`.**

```go
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"github.com/JakeFAU/visual_agent/internal/graph"
)

type Storage struct {
	dataDir string
}

func New(dir string) (*Storage, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &Storage{dataDir: dir}, nil
}

func (s *Storage) Save(g graph.Graph) error {
	path := filepath.Join(s.dataDir, g.Name+".json")
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
// ... Load and List
```

---

### Task 3: Implement API Server

**Files:**
- Create: `visual_agent/back_end/internal/server/server.go`

- [ ] **Step 1: Define the server struct and routes.**
- [ ] **Step 2: Implement `/api/execute` with SSE streaming.**

```go
func (s *Server) Execute(c *gin.Context) {
    // 1. Bind JSON { graph, input }
    // 2. Compile graph
    // 3. rt := runtime.NewLocalRuntime()
    // 4. Set SSE headers
    c.Stream(func(w io.Writer) bool {
        for event, err := range rt.Execute(ctx, compiled, input) {
            // Send JSON event string
            c.SSEvent("message", event)
        }
        return false
    })
}
```

---

### Task 4: CLI "Serve" Command

**Files:**
- Modify: `visual_agent/back_end/cmd/visual-agent/main.go`

- [ ] **Step 1: Add a `serve` command to start the Gin server.**

---

### Task 5: Verification

- [ ] **Step 1: Save a graph via `curl` and verify it appears in `data/graphs/`.**
- [ ] **Step 2: Execute a simple graph via `curl` and verify the SSE stream.**
