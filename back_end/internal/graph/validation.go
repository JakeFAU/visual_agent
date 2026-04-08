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
	incoming := make(map[string][]Edge, len(g.Nodes))
	outgoing := make(map[string][]Edge, len(g.Nodes))
	var inputNodeID string

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

		if node.Type == "input_node" {
			if inputNodeID != "" {
				return fmt.Errorf("graph must define exactly one input_node")
			}
			inputNodeID = node.ID
		}
	}

	for _, node := range g.Nodes {
		if node.ParentID == "" {
			continue
		}
		if node.ParentID == node.ID {
			return fmt.Errorf("node %q cannot be its own parent", node.ID)
		}
		parent, ok := nodeIDs[node.ParentID]
		if !ok {
			return fmt.Errorf("node %q references unknown parent node %q", node.ID, node.ParentID)
		}
		if parent.Type != "while_node" {
			return fmt.Errorf("node %q can only be parented under a while_node container", node.ID)
		}
		if node.Type == "input_node" {
			return fmt.Errorf("input node %q cannot live inside a while_node container", node.ID)
		}
		if err := validateParentChain(node, nodeIDs); err != nil {
			return err
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

		outgoing[edge.Source] = append(outgoing[edge.Source], edge)
		incoming[edge.Target] = append(incoming[edge.Target], edge)
	}

	if inputNodeID == "" {
		return fmt.Errorf("graph must define exactly one input_node")
	}

	for _, node := range g.Nodes {
		switch node.Type {
		case "input_node":
			controlEdges := 0
			for _, edge := range outgoing[node.ID] {
				target := nodeIDs[edge.Target]
				if isToolboxConnection(edge, node, target) {
					return fmt.Errorf("input node %q cannot connect to toolbox handles", node.ID)
				}
				if !isExecutionNode(target) {
					return fmt.Errorf("input node %q must target an execution node", node.ID)
				}
				controlEdges++
			}
			if controlEdges != 1 {
				return fmt.Errorf("input node %q must define exactly one outgoing execution edge", node.ID)
			}
		case "llm_node":
			successorCount := 0
			for _, edge := range outgoing[node.ID] {
				target := nodeIDs[edge.Target]
				switch {
				case target.Type == "output_node":
					continue
				case isToolboxConnection(edge, node, target):
					return fmt.Errorf("llm node %q cannot originate toolbox edges", node.ID)
				case isExecutionNode(target):
					successorCount++
				default:
					return fmt.Errorf("llm node %q cannot target node %q of type %q", node.ID, target.ID, target.Type)
				}
			}
			if successorCount > 1 {
				return fmt.Errorf("llm node %q must define at most one execution successor", node.ID)
			}
		case "if_else_node":
			var trueCount, falseCount int
			for _, edge := range outgoing[node.ID] {
				target := nodeIDs[edge.Target]
				switch edge.SourcePort {
				case "message:true", "out_true":
					if !isExecutionNode(target) {
						return fmt.Errorf("if_else node %q true branch must target an execution node", node.ID)
					}
					trueCount++
				case "message:false", "out_false":
					if !isExecutionNode(target) {
						return fmt.Errorf("if_else node %q false branch must target an execution node", node.ID)
					}
					falseCount++
				default:
					return fmt.Errorf("if_else node %q only supports true/false branch outputs", node.ID)
				}
			}
			if trueCount != 1 || falseCount != 1 {
				return fmt.Errorf("if_else node %q must define exactly one true branch and one false branch", node.ID)
			}
		case "while_node":
			var loopCount, doneCount int
			for _, edge := range outgoing[node.ID] {
				target := nodeIDs[edge.Target]
				switch edge.SourcePort {
				case "message:loop", "out_loop":
					if !isExecutionNode(target) {
						return fmt.Errorf("while node %q loop branch must target an execution node", node.ID)
					}
					loopCount++
				case "message:done", "out_done":
					if !isExecutionNode(target) && target.Type != "output_node" {
						return fmt.Errorf("while node %q done branch must target an execution node or output node", node.ID)
					}
					doneCount++
				default:
					return fmt.Errorf("while node %q only supports loop/done outputs", node.ID)
				}
			}
			if loopCount != 1 || doneCount != 1 {
				return fmt.Errorf("while node %q must define exactly one loop branch and one done branch", node.ID)
			}
		case "output_node":
			if len(outgoing[node.ID]) != 0 {
				return fmt.Errorf("output node %q cannot have outgoing edges", node.ID)
			}
			if len(incoming[node.ID]) != 1 {
				return fmt.Errorf("output node %q must have exactly one incoming edge", node.ID)
			}
			incomingEdge := incoming[node.ID][0]
			source := nodeIDs[incomingEdge.Source]
			switch source.Type {
			case "llm_node":
			case "while_node":
				if !isWhileDoneEdge(incomingEdge) {
					return fmt.Errorf("output node %q must be driven by a while_node done branch", node.ID)
				}
				if countWhileDoneProducers(source.ID, nodeIDs, incoming) == 0 {
					return fmt.Errorf("output node %q must be backed by at least one execution node feeding the while container output", node.ID)
				}
			default:
				return fmt.Errorf("output node %q must be driven by an llm_node or while_node", node.ID)
			}
		case "toolbox":
			if len(incoming[node.ID]) != 0 {
				return fmt.Errorf("toolbox node %q cannot have incoming edges", node.ID)
			}
			for _, edge := range outgoing[node.ID] {
				target := nodeIDs[edge.Target]
				if !isToolboxConnection(edge, node, target) {
					return fmt.Errorf("toolbox node %q can only connect to llm toolbox handles", node.ID)
				}
			}
		}
	}

	reachable := walkReachableExecutionNodes(inputNodeID, nodeIDs, outgoing)
	for _, node := range g.Nodes {
		if !isExecutionNode(node) {
			continue
		}
		if !reachable[node.ID] {
			return fmt.Errorf("execution node %q is unreachable from the input node", node.ID)
		}
	}

	return nil
}

func validateParentChain(node Node, nodeIDs map[string]Node) error {
	seen := map[string]bool{node.ID: true}
	currentParentID := node.ParentID

	for currentParentID != "" {
		if seen[currentParentID] {
			return fmt.Errorf("node %q participates in a parent cycle", node.ID)
		}
		seen[currentParentID] = true

		parent, ok := nodeIDs[currentParentID]
		if !ok {
			return fmt.Errorf("node %q references unknown parent node %q", node.ID, currentParentID)
		}
		currentParentID = parent.ParentID
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
		if strings.TrimSpace(cfg.Condition) == "" {
			return fmt.Errorf("condition cannot be empty")
		}
		if cfg.MaxIterations <= 0 {
			return fmt.Errorf("max_iterations must be greater than zero")
		}
	default:
		return fmt.Errorf("unsupported config type %T", n.Config)
	}

	return nil
}

func isExecutionNode(node Node) bool {
	_, ok := node.AgentName()
	return ok
}

func isToolboxConnection(edge Edge, source, target Node) bool {
	if source.Type != "toolbox" {
		return false
	}
	return (edge.TargetPort == "toolbox_handle" || edge.TargetPort == "in_toolbox") && target.Type == "llm_node"
}

func isWhileDoneEdge(edge Edge) bool {
	return edge.SourcePort == "message:done" || edge.SourcePort == "out_done"
}

func isWhileDoneTarget(edge Edge) bool {
	return edge.TargetPort == "message:done" || edge.TargetPort == "message:return" || edge.TargetPort == "out_done"
}

func countWhileDoneProducers(whileNodeID string, nodeIDs map[string]Node, incoming map[string][]Edge) int {
	count := 0
	for _, edge := range incoming[whileNodeID] {
		source, ok := nodeIDs[edge.Source]
		if !ok {
			continue
		}
		if !isExecutionNode(source) || !isWhileDoneTarget(edge) {
			continue
		}
		count++
	}
	return count
}

func walkReachableExecutionNodes(inputNodeID string, nodeIDs map[string]Node, outgoing map[string][]Edge) map[string]bool {
	reachable := make(map[string]bool)
	queue := []string{inputNodeID}
	seen := map[string]bool{inputNodeID: true}

	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]

		for _, edge := range outgoing[nodeID] {
			target := nodeIDs[edge.Target]
			if target.Type == "output_node" || isToolboxConnection(edge, nodeIDs[nodeID], target) {
				continue
			}
			if !isExecutionNode(target) || seen[target.ID] {
				continue
			}
			seen[target.ID] = true
			reachable[target.ID] = true
			queue = append(queue, target.ID)
		}
	}

	return reachable
}
