package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

func New(s *storage.Storage) *Server {
	r := gin.Default()
	r.Use(cors.Default())

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

func (s *Server) routes() {
	s.router.GET("/api/graphs", s.ListGraphs)
	s.router.GET("/api/graphs/:name", s.GetGraph)
	s.router.POST("/api/graphs", s.SaveGraph)
	s.router.POST("/api/execute", s.Execute)
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

func (s *Server) ListGraphs(c *gin.Context) {
	names, err := s.storage.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, names)
}

func (s *Server) GetGraph(c *gin.Context) {
	name := c.Param("name")
	g, err := s.storage.Load(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "graph not found"})
		return
	}
	c.JSON(http.StatusOK, g)
}

func (s *Server) SaveGraph(c *gin.Context) {
	var g graph.Graph
	if err := c.ShouldBindJSON(&g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.storage.Save(g); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "saved"})
}

type ExecuteRequest struct {
	Graph graph.Graph `json:"graph"`
	Input string      `json:"input"`
}

func (s *Server) Execute(c *gin.Context) {
	fmt.Println("[DEBUG] Execute endpoint called")
	var req ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("[DEBUG] Bind JSON failed: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("[DEBUG] Compiling graph: %s\n", req.Graph.Name)
	compiled, err := s.compiler.Compile(req.Graph)
	if err != nil {
		fmt.Printf("[DEBUG] Compilation failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("compilation failed: %v", err)})
		return
	}

	rt := runtime.NewLocalRuntime()

	fmt.Println("[DEBUG] Starting SSE stream")
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	c.Stream(func(_ io.Writer) bool {
		fmt.Println("[DEBUG] Entering execution loop")
		for event, err := range rt.Execute(c.Request.Context(), compiled, req.Input) {
			if err != nil {
				fmt.Printf("[DEBUG] Execution error: %v\n", err)
				data, _ := json.Marshal(gin.H{"type": "error", "content": err.Error()})
				c.SSEvent("message", string(data))
				c.Writer.Flush()
				return false
			}
			if event != nil {
				fmt.Printf("[DEBUG] Sending event from author: %s\n", event.Author)
				data, _ := json.Marshal(gin.H{
					"type":    "agent_event",
					"content": event,
					"author":  event.Author,
				})
				c.SSEvent("message", string(data))
				c.Writer.Flush()
			}
		}
		fmt.Println("[DEBUG] Execution loop finished, sending done")
		data, _ := json.Marshal(gin.H{"type": "done", "content": "execution complete"})
		c.SSEvent("message", string(data))
		c.Writer.Flush()
		return false
	})
}
