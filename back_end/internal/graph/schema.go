package graph

import (
	"encoding/json"
	"fmt"
)

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type InputNodeConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type LLMNodeConfig struct {
	Name                  string                 `json:"name"`
	Description           string                 `json:"description"`
	Model                 string                 `json:"model"`
	Instruction           string                 `json:"instruction"`
	ResponseMode          string                 `json:"response_mode"` // "text" or "json"
	OutputSchema          map[string]interface{} `json:"output_schema,omitempty"`
	GenerateContentConfig GenerateContentConfig `json:"generate_content_config"`
}

type GenerateContentConfig struct {
	Temperature      float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"max_output_tokens,omitempty"`
}

type OutputNodeConfig struct {
	Name      string `json:"name"`
	OutputKey string `json:"output_key"`
	Format    string `json:"format"`
}

type MCPServerConfig struct {
	Name    string            `json:"name"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

type CustomFunctionConfig struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type ToolboxNodeConfig struct {
	Tools           []string               `json:"tools"`
	MCPServers      []MCPServerConfig      `json:"mcp_servers"`
	CustomFunctions []CustomFunctionConfig `json:"custom_functions"`
}

type IfElseNodeConfig struct {
	ConditionLanguage string `json:"condition_language"` // "CEL" or "JSONPath"
	Condition         string `json:"condition"`
}

type WhileNodeConfig struct {
	Condition     string `json:"condition"`
	MaxIterations int    `json:"max_iterations"`
}

type Node struct {
	ID       string      `json:"id"`
	Type     string      `json:"type"`
	Position Position    `json:"position"`
	Config   interface{} `json:"config"`
}

type Edge struct {
	ID         string `json:"id" mapstructure:"id"`
	Source     string `json:"source" mapstructure:"source"`
	SourcePort string `json:"source_port" mapstructure:"source_port"`
	Target     string `json:"target" mapstructure:"target"`
	TargetPort string `json:"target_port" mapstructure:"target_port"`
	DataType   string `json:"data_type" mapstructure:"data_type"`
	EdgeKind   string `json:"edge_kind" mapstructure:"edge_kind"`
}

type Graph struct {
	Version string `json:"version"`
	Name    string `json:"name"`
	Nodes   []Node `json:"nodes"`
	Edges   []Edge `json:"edges"`
}

func (n *Node) UnmarshalJSON(data []byte) error {
	type Alias Node
	aux := &struct {
		Config json.RawMessage `json:"config"`
		*Alias
	}{
		Alias: (*Alias)(n),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch n.Type {
	case "input_node":
		var config InputNodeConfig
		if err := json.Unmarshal(aux.Config, &config); err != nil {
			return err
		}
		n.Config = config
	case "llm_node":
		var config LLMNodeConfig
		if err := json.Unmarshal(aux.Config, &config); err != nil {
			return err
		}
		n.Config = config
	case "output_node":
		var config OutputNodeConfig
		if err := json.Unmarshal(aux.Config, &config); err != nil {
			return err
		}
		n.Config = config
	case "toolbox":
		var config ToolboxNodeConfig
		if err := json.Unmarshal(aux.Config, &config); err != nil {
			return err
		}
		n.Config = config
	case "if_else_node":
		var config IfElseNodeConfig
		if err := json.Unmarshal(aux.Config, &config); err != nil {
			return err
		}
		n.Config = config
	case "while_node":
		var config WhileNodeConfig
		if err := json.Unmarshal(aux.Config, &config); err != nil {
			return err
		}
		n.Config = config
	default:
		return fmt.Errorf("unknown node type: %s", n.Type)
	}

	return nil
}
