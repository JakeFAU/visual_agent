package compiler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/JakeFAU/visual_agent/internal/graph"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/tool"
	"google.golang.org/genai"
	"os"
)

type LLMNodeCompiler struct{}

func (c *LLMNodeCompiler) Compile(node graph.Node, metadata map[string]interface{}) (any, error) {
	cfg, ok := node.Config.(graph.LLMNodeConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config for llm_node")
	}

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		apiKey = "dummy"
	}

	ctx := context.Background()
	model, err := gemini.NewModel(ctx, cfg.Model, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	llmCfg := llmagent.Config{
		Name:        cfg.Name,
		Description: cfg.Description,
		Model:       model,
		Instruction: cfg.Instruction,
	}

	if cfg.ResponseMode == "json" && cfg.OutputSchema != nil {
		schemaBytes, _ := json.Marshal(cfg.OutputSchema)
		var schema genai.Schema
		if err := json.Unmarshal(schemaBytes, &schema); err != nil {
			return nil, fmt.Errorf("failed to parse output schema: %w", err)
		}
		llmCfg.OutputSchema = &schema
	}

	// Apply output keys from walker if present
	if keys, ok := metadata["output_keys"].([]string); ok && len(keys) > 0 {
		llmCfg.OutputKey = keys[0]
	}

	// Apply toolsets if present
	if toolsets, ok := metadata["toolsets"].([]tool.Toolset); ok {
		llmCfg.Toolsets = toolsets
	}

	agentInstance, err := llmagent.New(llmCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create llmagent: %w", err)
	}

	return agentInstance, nil
}
