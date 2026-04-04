package compiler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/JakeFAU/visual_agent/internal/graph"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/tool"
	"google.golang.org/genai"
)

// ModelFactory is a function that creates a new LLM model instance.
type ModelFactory func(ctx context.Context, modelName string, cfg *genai.ClientConfig) (model.LLM, error)

type LLMNodeCompiler struct {
	// NewModel allows overriding the model creation for testing.
	NewModel ModelFactory
}

func (c *LLMNodeCompiler) Compile(node graph.Node, metadata map[string]interface{}) (any, error) {
	cfg, ok := node.Config.(graph.LLMNodeConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config for llm_node")
	}

	apiKey := os.Getenv("GOOGLE_API_KEY")
	clientCfg := &genai.ClientConfig{}
	if apiKey != "" {
		clientCfg.APIKey = apiKey
	} else {
		// Use Vertex AI with ADC
		clientCfg.Backend = genai.BackendVertexAI
		clientCfg.Project = os.Getenv("GOOGLE_CLOUD_PROJECT")
		clientCfg.Location = os.Getenv("GOOGLE_CLOUD_LOCATION")
		if clientCfg.Location == "" {
			clientCfg.Location = "us-central1"
		}
	}

	ctx := context.Background()
	newModel := c.NewModel
	if newModel == nil {
		newModel = gemini.NewModel
	}

	modelInst, err := newModel(ctx, cfg.Model, clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	llmCfg := llmagent.Config{
		Name:        cfg.Name,
		Description: cfg.Description,
		Model:       modelInst,
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
