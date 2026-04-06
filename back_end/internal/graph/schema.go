package graph

import (
	"encoding/json"
	"fmt"
)

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// InputNodeConfig describes the synthetic "user input" entry point shown in the
// editor. It is primarily documentation for humans; runtime execution receives
// the actual input string separately.
type InputNodeConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// LLMNodeConfig contains the author-facing settings for an LLM execution node.
type LLMNodeConfig struct {
	Name                  string                 `json:"name"`
	Description           string                 `json:"description"`
	Model                 string                 `json:"model"`
	Instruction           string                 `json:"instruction"`
	ResponseMode          string                 `json:"response_mode"` // "text" or "json"
	OutputSchema          map[string]interface{} `json:"output_schema,omitempty"`
	GenerateContentConfig GenerateContentConfig  `json:"generate_content_config"`
}

// GenerateContentConfig mirrors the subset of generation controls currently
// surfaced by the frontend and supported by the backend compiler.
type GenerateContentConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"max_output_tokens,omitempty"`
}

// OutputNodeConfig tells the compiler which session-state key should be treated
// as the user-facing result for a branch or full workflow.
type OutputNodeConfig struct {
	Name      string `json:"name"`
	OutputKey string `json:"output_key"`
	Format    string `json:"format"`
}

// MCPServerConfig declares an external MCP command that should be launched and
// exposed to the model as a toolset.
type MCPServerConfig struct {
	Name    string            `json:"name"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// CustomFunctionConfig reserves space in the contract for future custom tool
// declarations. The backend currently preserves the data but does not execute
// it in v0.
type CustomFunctionConfig struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolboxNodeConfig groups built-in tools, MCP servers, and future custom
// functions into the contract shape used by toolbox nodes.
type ToolboxNodeConfig struct {
	Tools           []string               `json:"tools"`
	MCPServers      []MCPServerConfig      `json:"mcp_servers"`
	CustomFunctions []CustomFunctionConfig `json:"custom_functions"`
}

// IfElseNodeConfig defines a control-flow node that evaluates a condition and
// transfers execution to one of two downstream agents.
type IfElseNodeConfig struct {
	ConditionLanguage string `json:"condition_language"` // Currently only "CEL" is supported.
	Condition         string `json:"condition"`
}

// WhileNodeConfig defines a bounded loop gate.
//
// The node evaluates Condition on each visit. When the condition returns true,
// execution follows the loop branch. When it returns false, execution follows
// the done branch. MaxIterations is a required local safety cap for that
// specific loop block and is separate from any global per-run execution budget.
type WhileNodeConfig struct {
	Condition     string `json:"condition"`
	MaxIterations int    `json:"max_iterations"`
}

// Node is the discriminated graph node used throughout the backend.
//
// Config is unmarshaled into one of the concrete config structs above based on
// Type.
type Node struct {
	ID       string      `json:"id"`
	Type     string      `json:"type"`
	Position Position    `json:"position"`
	ParentID string      `json:"parent_id,omitempty"`
	Width    float64     `json:"width,omitempty"`
	Height   float64     `json:"height,omitempty"`
	Config   interface{} `json:"config"`
}

// Edge connects two node handles in the serialized graph contract.
type Edge struct {
	ID         string `json:"id" mapstructure:"id"`
	Source     string `json:"source" mapstructure:"source"`
	SourcePort string `json:"source_port" mapstructure:"source_port"`
	Target     string `json:"target" mapstructure:"target"`
	TargetPort string `json:"target_port" mapstructure:"target_port"`
	DataType   string `json:"data_type" mapstructure:"data_type"`
	EdgeKind   string `json:"edge_kind" mapstructure:"edge_kind"`
}

// Graph is the top-level workflow document exchanged between frontend and
// backend.
type Graph struct {
	Version string `json:"version"`
	Name    string `json:"name"`
	Nodes   []Node `json:"nodes"`
	Edges   []Edge `json:"edges"`
}

// AgentName returns the runtime-visible agent name for nodes that participate
// in execution.
//
// Non-execution nodes such as input, output, and toolbox nodes do not have
// agent identities and therefore return false.
func (n Node) AgentName() (string, bool) {
	switch cfg := n.Config.(type) {
	case LLMNodeConfig:
		return cfg.Name, true
	case IfElseNodeConfig:
		return n.ID, true
	case WhileNodeConfig:
		return n.ID, true
	default:
		return "", false
	}
}

// UnmarshalJSON decodes a node and then materializes Config into the concrete
// struct that matches the node's type discriminator.
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
