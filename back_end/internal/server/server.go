package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/JakeFAU/visual_agent/internal/compiler"
	"github.com/JakeFAU/visual_agent/internal/graph"
	"github.com/JakeFAU/visual_agent/internal/runtime"
	"github.com/JakeFAU/visual_agent/internal/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"google.golang.org/adk/session"
)

type Server struct {
	router   *gin.Engine
	storage  *storage.Storage
	compiler *compiler.Compiler
}

// New constructs the HTTP server and registers the node compilers required by
// the API execution path.
func New(s *storage.Storage) *Server {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		// The browser UI is intended to run on the same machine as the backend, so
		// only loopback origins are accepted by default.
		AllowOriginFunc: func(origin string) bool {
			u, err := url.Parse(origin)
			if err != nil {
				return false
			}
			switch u.Hostname() {
			case "localhost", "127.0.0.1":
				return true
			default:
				return false
			}
		},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
		MaxAge:       12 * time.Hour,
	}))

	c := compiler.New()
	c.Register("llm_node", &compiler.LLMNodeCompiler{})
	c.Register("toolbox", &compiler.ToolboxNodeCompiler{})
	c.Register("if_else_node", &compiler.IfElseNodeCompiler{})
	c.Register("while_node", &compiler.WhileNodeCompiler{})

	srv := &Server{
		router:   r,
		storage:  s,
		compiler: c,
	}

	srv.routes()
	return srv
}

// routes registers the REST and SSE endpoints consumed by the frontend.
func (s *Server) routes() {
	s.router.GET("/api/graphs", s.ListGraphs)
	s.router.GET("/api/graphs/:name", s.GetGraph)
	s.router.POST("/api/graphs", s.SaveGraph)
	s.router.POST("/api/execute", s.Execute)
}

// Run starts the HTTP server at the configured address.
func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

// ListGraphs returns the names of all saved workflow documents.
func (s *Server) ListGraphs(c *gin.Context) {
	names, err := s.storage.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, names)
}

// GetGraph loads a single saved graph by name.
func (s *Server) GetGraph(c *gin.Context) {
	name := c.Param("name")
	g, err := s.storage.Load(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "graph not found"})
		return
	}
	c.JSON(http.StatusOK, g)
}

// SaveGraph validates and persists a workflow document sent by the frontend.
func (s *Server) SaveGraph(c *gin.Context) {
	var g graph.Graph
	if err := c.ShouldBindJSON(&g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := g.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.storage.Save(g); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "saved"})
}

// ExecuteRequest is the JSON payload accepted by POST /api/execute.
type ExecuteRequest struct {
	Graph  graph.Graph     `json:"graph"`
	Input  string          `json:"input"`
	Budget ExecutionBudget `json:"budget"`
}

// ExecutionBudget declares per-run safety limits for a single execution.
type ExecutionBudget struct {
	MaxSteps       int `json:"max_steps,omitempty"`
	MaxDurationMS  int `json:"max_duration_ms,omitempty"`
	MaxTotalTokens int `json:"max_total_tokens,omitempty"`
}

// Execute validates the submitted graph, compiles it, and streams execution
// events back to the client using server-sent events.
func (s *Server) Execute(c *gin.Context) {
	fmt.Println("[DEBUG] Execute endpoint called")
	var req ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("[DEBUG] Bind JSON failed: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := req.Graph.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := req.Budget.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("[DEBUG] Compiling graph: %s\n", req.Graph.Name)
	compiled, err := s.compiler.CompileWithOptions(req.Graph, compiler.CompileOptions{
		MaxSteps: req.Budget.MaxSteps,
	})
	if err != nil {
		fmt.Printf("[DEBUG] Compilation failed: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("compilation failed: %v", err)})
		return
	}

	// API execution uses the local in-memory runtime because the compiled agent
	// already encapsulates model access and tool wiring.
	rt := runtime.NewLocalRuntime()
	execCtx := c.Request.Context()
	var cancel context.CancelFunc
	if req.Budget.MaxDurationMS > 0 {
		execCtx, cancel = context.WithTimeout(execCtx, time.Duration(req.Budget.MaxDurationMS)*time.Millisecond)
		defer cancel()
	}
	tracker := newExecutionTracker(req.Budget)

	fmt.Println("[DEBUG] Starting SSE stream")
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Prevent proxy buffering (Nginx)

	c.Stream(func(_ io.Writer) bool {
		fmt.Println("[DEBUG] Entering execution loop")
		for event, err := range rt.Execute(execCtx, compiled, req.Input) {
			if err != nil {
				exitReason, message := tracker.errorPayload(execCtx, err)
				fmt.Printf("[DEBUG] Execution error: %v\n", err)
				writeSSE(c, gin.H{
					"type":    "diagnostic",
					"content": tracker.summary(exitReason),
				})
				writeSSE(c, gin.H{"type": "error", "content": message})
				return false
			}
			if event != nil {
				if diagnostic, ok := extractDiagnostic(event); ok {
					tracker.observeDiagnostic(diagnostic)
					writeSSE(c, gin.H{
						"type":    "diagnostic",
						"content": diagnostic,
					})
					continue
				}

				fmt.Printf("[DEBUG] Sending event from author: %s\n", event.Author)
				tracker.observeAgentEvent(event)
				writeSSE(c, gin.H{
					"type":    "agent_event",
					"content": event,
					"author":  event.Author,
				})

				if req.Budget.MaxTotalTokens > 0 && tracker.TotalTokens > req.Budget.MaxTotalTokens {
					writeSSE(c, gin.H{
						"type":    "diagnostic",
						"content": tracker.summary("token_budget_exceeded"),
					})
					writeSSE(c, gin.H{
						"type":    "error",
						"content": fmt.Sprintf("token budget exceeded: used %d tokens against limit %d", tracker.TotalTokens, req.Budget.MaxTotalTokens),
					})
					return false
				}
			}
		}
		fmt.Println("[DEBUG] Execution loop finished, sending done")
		writeSSE(c, gin.H{
			"type":    "diagnostic",
			"content": tracker.summary("completed"),
		})
		writeSSE(c, gin.H{"type": "done", "content": "execution complete"})
		return false
	})
}

func (b ExecutionBudget) Validate() error {
	switch {
	case b.MaxSteps < 0:
		return fmt.Errorf("budget.max_steps cannot be negative")
	case b.MaxDurationMS < 0:
		return fmt.Errorf("budget.max_duration_ms cannot be negative")
	case b.MaxTotalTokens < 0:
		return fmt.Errorf("budget.max_total_tokens cannot be negative")
	default:
		return nil
	}
}

type executionTracker struct {
	Budget      ExecutionBudget
	StartedAt   time.Time
	Steps       int
	TotalTokens int
	CurrentNode string
	NodeVisits  map[string]int
}

func newExecutionTracker(budget ExecutionBudget) executionTracker {
	return executionTracker{
		Budget:     budget,
		StartedAt:  time.Now(),
		NodeVisits: make(map[string]int),
	}
}

func (t *executionTracker) observeDiagnostic(payload map[string]any) {
	if payload["kind"] != "node_enter" {
		return
	}

	t.Steps = intValue(payload["step"])
	if currentNode, ok := payload["agent_name"].(string); ok {
		t.CurrentNode = currentNode
	}
	if visitCount := intValue(payload["visit_count"]); visitCount > 0 {
		t.NodeVisits[t.CurrentNode] = visitCount
	}
}

func (t *executionTracker) observeAgentEvent(event *session.Event) {
	if event == nil || event.Partial || event.UsageMetadata == nil {
		return
	}
	t.TotalTokens += int(event.UsageMetadata.TotalTokenCount)
}

func (t executionTracker) summary(exitReason string) map[string]any {
	return map[string]any{
		"kind":         "summary",
		"exit_reason":  exitReason,
		"steps":        t.Steps,
		"elapsed_ms":   time.Since(t.StartedAt).Milliseconds(),
		"total_tokens": t.TotalTokens,
		"current_node": t.CurrentNode,
		"node_visits":  t.NodeVisits,
		"budget": map[string]any{
			"max_steps":        t.Budget.MaxSteps,
			"max_duration_ms":  t.Budget.MaxDurationMS,
			"max_total_tokens": t.Budget.MaxTotalTokens,
		},
	}
}

func (t executionTracker) errorPayload(execCtx context.Context, err error) (string, string) {
	if execCtx.Err() == context.DeadlineExceeded || strings.Contains(strings.ToLower(err.Error()), "deadline exceeded") {
		return "duration_budget_exceeded", fmt.Sprintf("duration budget exceeded after %d ms", t.Budget.MaxDurationMS)
	}
	if strings.Contains(err.Error(), "possible infinite loop") {
		return "step_budget_exceeded", err.Error()
	}
	return "error", err.Error()
}

func extractDiagnostic(event *session.Event) (map[string]any, bool) {
	if event == nil || event.CustomMetadata == nil {
		return nil, false
	}
	payload, ok := event.CustomMetadata["graph_runtime_diagnostic"].(map[string]any)
	return payload, ok
}

func intValue(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

func writeSSE(c *gin.Context, payload any) {
	data, _ := json.Marshal(payload)
	fmt.Fprintf(c.Writer, "data: %s\n\n", string(data))
	c.Writer.Flush()
}
