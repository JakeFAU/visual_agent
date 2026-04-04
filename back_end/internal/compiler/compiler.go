package compiler

import (
	"fmt"
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

func (c *Compiler) Compile(g graph.Graph) (interface{}, error) {
	// For v0, we'll implement a simple sequential execution of LLM nodes
	// In v1, this will build a full ADK DAG (SequentialAgent, ParallelAgent, etc.)
	
	compiledNodes := make(map[string]interface{})
	
	for _, node := range g.Nodes {
		nc, ok := c.compilers[node.Type]
		if !ok {
			return nil, fmt.Errorf("no compiler registered for node type: %s", node.Type)
		}
		
		compiled, err := nc.Compile(node)
		if err != nil {
			return nil, fmt.Errorf("failed to compile node %s: %w", node.ID, err)
		}
		
		compiledNodes[node.ID] = compiled
	}
	
	return compiledNodes, nil
}
