package compiler

import (
	"fmt"
	"github.com/JakeFAU/visual_agent/internal/graph"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
)

type LLMNodeCompiler struct{}

func (c *LLMNodeCompiler) Compile(node graph.Node) (interface{}, error) {
	cfg, ok := node.Config.(graph.LLMNodeConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config for llm_node")
	}

	// In a real implementation, the model would be configured based on cfg.Model
	// and GenerationConfig. For now, we'll use a default Gemini setup.
	model := gemini.New(cfg.Model)
	
	options := []llmagent.Option{
		llmagent.WithInstructions(cfg.Instruction),
	}

	if cfg.ResponseMode == "json" {
		// ADK supports output schemas
		options = append(options, llmagent.WithOutputSchema(cfg.OutputSchema))
	}

	agent := llmagent.New(cfg.Name, cfg.Description, model, options...)
	
	return agent, nil
}
