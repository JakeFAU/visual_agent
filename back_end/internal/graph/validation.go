package graph

import (
	"fmt"
	"strings"
)

const SupportedGraphVersion = "1.0"

var supportedBuiltInTools = map[string]struct{}{
	"google_search": {},
}

// Validate enforces the backend-supported subset of the graph contract.
//
// This is intentionally stricter than "the JSON parsed successfully": it
// checks node uniqueness, edge references, reserved names, and control-flow
// wiring so invalid workflows fail early before compilation or execution.
func (g Graph) Validate() error {
	if g.Version != SupportedGraphVersion {
		return fmt.Errorf("unsupported graph version %q", g.Version)
	}

	if strings.TrimSpace(g.Name) == "" {
		return fmt.Errorf("graph name cannot be empty")
	}
	if strings.Contains(g.Name, "/") || strings.Contains(g.Name, "\\") {
		return fmt.Errorf("graph name cannot contain path separators")
	}

	nodeIDs := make(map[string]Node, len(g.Nodes))
	agentNames := make(map[string]string)

	for _, node := range g.Nodes {
		if strings.TrimSpace(node.ID) == "" {
			return fmt.Errorf("node id cannot be empty")
		}
		if _, exists := nodeIDs[node.ID]; exists {
			return fmt.Errorf("duplicate node id %q", node.ID)
		}
		nodeIDs[node.ID] = node

		if err := node.Validate(); err != nil {
			return fmt.Errorf("node %q: %w", node.ID, err)
		}

		if name, ok := node.AgentName(); ok {
			if existing, exists := agentNames[name]; exists {
				return fmt.Errorf("duplicate agent name %q used by nodes %q and %q", name, existing, node.ID)
			}
			agentNames[name] = node.ID
		}
	}

	for _, edge := range g.Edges {
		if strings.TrimSpace(edge.ID) == "" {
			return fmt.Errorf("edge id cannot be empty")
		}
		if _, ok := nodeIDs[edge.Source]; !ok {
			return fmt.Errorf("edge %q references unknown source node %q", edge.ID, edge.Source)
		}
		if _, ok := nodeIDs[edge.Target]; !ok {
			return fmt.Errorf("edge %q references unknown target node %q", edge.ID, edge.Target)
		}
		if strings.TrimSpace(edge.SourcePort) == "" {
			return fmt.Errorf("edge %q source_port cannot be empty", edge.ID)
		}
		if strings.TrimSpace(edge.TargetPort) == "" {
			return fmt.Errorf("edge %q target_port cannot be empty", edge.ID)
		}
	}

	for _, node := range g.Nodes {
		if node.Type != "if_else_node" {
			continue
		}

		var trueCount, falseCount int
		for _, edge := range g.Edges {
			if edge.Source != node.ID {
				continue
			}
			switch edge.SourcePort {
			case "message:true", "out_true":
				if _, ok := nodeIDs[edge.Target].AgentName(); !ok {
					return fmt.Errorf("if_else node %q true branch must target an execution node", node.ID)
				}
				trueCount++
			case "message:false", "out_false":
				if _, ok := nodeIDs[edge.Target].AgentName(); !ok {
					return fmt.Errorf("if_else node %q false branch must target an execution node", node.ID)
				}
				falseCount++
			}
		}

		if trueCount != 1 || falseCount != 1 {
			return fmt.Errorf("if_else node %q must define exactly one true branch and one false branch", node.ID)
		}
	}

	return nil
}

// Validate checks a node's configuration according to its concrete type.
func (n Node) Validate() error {
	switch cfg := n.Config.(type) {
	case InputNodeConfig:
		if strings.TrimSpace(cfg.Name) == "" {
			return fmt.Errorf("input node name cannot be empty")
		}
	case LLMNodeConfig:
		if strings.TrimSpace(cfg.Name) == "" {
			return fmt.Errorf("llm node name cannot be empty")
		}
		if cfg.Name == "user" {
			return fmt.Errorf("llm node name %q is reserved", cfg.Name)
		}
		if strings.TrimSpace(cfg.Model) == "" {
			return fmt.Errorf("llm node model cannot be empty")
		}
		switch cfg.ResponseMode {
		case "text", "json":
		default:
			return fmt.Errorf("unsupported response_mode %q", cfg.ResponseMode)
		}
		if cfg.GenerateContentConfig.MaxOutputTokens < 0 {
			return fmt.Errorf("max_output_tokens cannot be negative")
		}
	case OutputNodeConfig:
		if strings.TrimSpace(cfg.Name) == "" {
			return fmt.Errorf("output node name cannot be empty")
		}
		if strings.TrimSpace(cfg.OutputKey) == "" {
			return fmt.Errorf("output_key cannot be empty")
		}
		if cfg.Format != "message" {
			return fmt.Errorf("unsupported output format %q", cfg.Format)
		}
	case ToolboxNodeConfig:
		for _, toolID := range cfg.Tools {
			if _, ok := supportedBuiltInTools[toolID]; !ok {
				return fmt.Errorf("unsupported built-in tool %q", toolID)
			}
		}
		for _, server := range cfg.MCPServers {
			if strings.TrimSpace(server.Name) == "" {
				return fmt.Errorf("mcp server name cannot be empty")
			}
			if strings.TrimSpace(server.Command) == "" {
				return fmt.Errorf("mcp server command cannot be empty")
			}
		}
	case IfElseNodeConfig:
		if cfg.ConditionLanguage != "CEL" {
			return fmt.Errorf("unsupported condition_language %q", cfg.ConditionLanguage)
		}
		if strings.TrimSpace(cfg.Condition) == "" {
			return fmt.Errorf("condition cannot be empty")
		}
	case WhileNodeConfig:
		return fmt.Errorf("while_node is not supported in v0")
	default:
		return fmt.Errorf("unsupported config type %T", n.Config)
	}

	return nil
}
