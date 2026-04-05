package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"

	"github.com/JakeFAU/visual_agent/internal/compiler"
	"github.com/JakeFAU/visual_agent/internal/graph"
	"github.com/JakeFAU/visual_agent/internal/runtime"
	"github.com/JakeFAU/visual_agent/internal/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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
	Graph graph.Graph `json:"graph"`
	Input string      `json:"input"`
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

	fmt.Printf("[DEBUG] Compiling graph: %s\n", req.Graph.Name)
	compiled, err := s.compiler.Compile(req.Graph)
	if err != nil {
		fmt.Printf("[DEBUG] Compilation failed: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("compilation failed: %v", err)})
		return
	}

	// API execution uses the local in-memory runtime because the compiled agent
	// already encapsulates model access and tool wiring.
	rt := runtime.NewLocalRuntime()

	fmt.Println("[DEBUG] Starting SSE stream")
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Prevent proxy buffering (Nginx)
	c.Status(http.StatusOK)
	c.Writer.WriteHeaderNow()
	writeSSEComment(c, "stream-open")

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[DEBUG] Execution panic: %v\n%s\n", r, debug.Stack())
			writeSSE(c, gin.H{
				"type":    "error",
				"content": fmt.Sprintf("execution panic: %v", r),
			})
		}
	}()

	fmt.Println("[DEBUG] Entering execution loop")
	for event, err := range rt.Execute(c.Request.Context(), compiled, req.Input) {
		if err != nil {
			fmt.Printf("[DEBUG] Execution error: %v\n", err)
			writeSSE(c, gin.H{"type": "error", "content": err.Error()})
			return
		}
		if event != nil {
			fmt.Printf("[DEBUG] Sending event from author: %s\n", event.Author)
			writeSSE(c, gin.H{
				"type":    "agent_event",
				"content": event,
				"author":  event.Author,
			})
		}
	}
	fmt.Println("[DEBUG] Execution loop finished, sending done")
	writeSSE(c, gin.H{"type": "done", "content": "execution complete"})
}

func writeSSE(c *gin.Context, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("[DEBUG] Failed to marshal SSE payload: %v\n", err)
		return
	}
	fmt.Fprintf(c.Writer, "data: %s\n\n", string(data))
	c.Writer.Flush()
}

func writeSSEComment(c *gin.Context, comment string) {
	fmt.Fprintf(c.Writer, ": %s\n\n", comment)
	c.Writer.Flush()
}
