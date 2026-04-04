package compiler

import (
	"github.com/JakeFAU/visual_agent/internal/graph"
	"google.golang.org/adk/agent"
)

type ToolboxNodeCompiler struct{}

func (c *ToolboxNodeCompiler) Compile(node graph.Node, metadata map[string]interface{}) (agent.Agent, error) {
	// For v0, Toolbox nodes are processed during LLM node compilation 
	// or handled as a separate orchestrator. 
	// Returning nil here to satisfy the interface for now.
	return nil, nil
}
