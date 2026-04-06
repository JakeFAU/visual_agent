package compiler

import (
	"fmt"

	"github.com/JakeFAU/visual_agent/internal/graph"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/tool"
)

func isToolboxEdge(edge graph.Edge) bool {
	return edge.TargetPort == "in_toolbox" || edge.TargetPort == "toolbox_handle"
}

// Both handle conventions are supported so saved graphs created before and
// after the port rename continue to compile.
func isIfElseTrueEdge(edge graph.Edge) bool {
	return edge.SourcePort == "message:true" || edge.SourcePort == "out_true"
}

func isIfElseFalseEdge(edge graph.Edge) bool {
	return edge.SourcePort == "message:false" || edge.SourcePort == "out_false"
}

func isWhileLoopEdge(edge graph.Edge) bool {
	return edge.SourcePort == "message:loop" || edge.SourcePort == "out_loop"
}

func isWhileDoneEdge(edge graph.Edge) bool {
	return edge.SourcePort == "message:done" || edge.SourcePort == "out_done"
}

func isWhileDoneTarget(edge graph.Edge) bool {
	return edge.TargetPort == "message:done" || edge.TargetPort == "out_done"
}

func isExecutionNode(node graph.Node) bool {
	_, ok := node.AgentName()
	return ok
}

func appendUnique(values []string, candidate string) []string {
	if candidate == "" {
		return values
	}
	for _, existing := range values {
		if existing == candidate {
			return values
		}
	}
	return append(values, candidate)
}

func defaultStateKey(node graph.Node) string {
	switch cfg := node.Config.(type) {
	case graph.LLMNodeConfig:
		return cfg.Name
	default:
		return ""
	}
}

// NodeCompiler translates a specific graph node into its ADK representation
// interface{} (any) allows nodes to return agents, toolsets, or other configs.
type NodeCompiler interface {
	Compile(node graph.Node, metadata map[string]interface{}) (any, error)
}

// Compiler orchestrates the translation of a full Graph
type Compiler struct {
	compilers map[string]NodeCompiler
}

// CompileOptions controls per-execution orchestration behavior for the graph
// runtime that wraps compiled node agents.
type CompileOptions struct {
	MaxSteps int
}

// New constructs a compiler with an empty node-compiler registry.
func New() *Compiler {
	return &Compiler{
		compilers: make(map[string]NodeCompiler),
	}
}

// Register associates a graph node type with the compiler responsible for
// translating it into an ADK representation.
func (c *Compiler) Register(nodeType string, nc NodeCompiler) {
	c.compilers[nodeType] = nc
}

// Compile translates an entire validated graph into a runnable ADK agent.
//
// The compiler materializes each execution node as a child ADK agent, then
// wraps them in a custom root agent that walks the graph dynamically at run
// time. That orchestration layer is what allows explicit cycles in the graph
// while still reusing ADK agents for the leaf execution work.
func (c *Compiler) Compile(g graph.Graph) (agent.Agent, error) {
	return c.CompileWithOptions(g, CompileOptions{})
}

// CompileWithOptions mirrors Compile but allows the caller to override runtime
// limits such as the maximum number of graph steps for a single execution.
func (c *Compiler) CompileWithOptions(g graph.Graph, opts CompileOptions) (agent.Agent, error) {
	fmt.Printf("[DEBUG] Starting compilation for graph: %s\n", g.Name)

	outputMappings := make(map[string][]string)
	toolsetMappings := make(map[string][]tool.Toolset)
	toolboxResults := make(map[string][]tool.Toolset)
	trueAgentMappings := make(map[string]string)
	falseAgentMappings := make(map[string]string)
	loopAgentMappings := make(map[string]string)
	doneAgentMappings := make(map[string]string)
	doneTerminalMappings := make(map[string]bool)
	staticNextMappings := make(map[string]string)
	whileDoneProducerMappings := make(map[string][]string)
	whileOutputKeyMappings := make(map[string][]string)
	outputNodeKeys := make(map[string]string)
	nodeByID := make(map[string]graph.Node, len(g.Nodes))
	agentNameToNodeID := make(map[string]string)
	startNodeID := ""

	for _, node := range g.Nodes {
		nodeByID[node.ID] = node

		if name, ok := node.AgentName(); ok {
			agentNameToNodeID[name] = node.ID
		}

		if node.Type != "output_node" {
			continue
		}

		cfg, ok := node.Config.(graph.OutputNodeConfig)
		if !ok || cfg.OutputKey == "" {
			continue
		}

		outputNodeKeys[node.ID] = cfg.OutputKey
	}

	for _, edge := range g.Edges {
		sourceNode, sourceOK := nodeByID[edge.Source]
		targetNode, targetOK := nodeByID[edge.Target]
		if !sourceOK || !targetOK || isToolboxEdge(edge) {
			continue
		}

		if targetNode.Type == "while_node" && isExecutionNode(sourceNode) && isWhileDoneTarget(edge) {
			whileDoneProducerMappings[targetNode.ID] = appendUnique(whileDoneProducerMappings[targetNode.ID], sourceNode.ID)
		}

		if targetNode.Type == "output_node" {
			if outputKey, ok := outputNodeKeys[targetNode.ID]; ok {
				if sourceNode.Type == "while_node" {
					whileOutputKeyMappings[sourceNode.ID] = appendUnique(whileOutputKeyMappings[sourceNode.ID], outputKey)
				} else {
					outputMappings[sourceNode.ID] = appendUnique(outputMappings[sourceNode.ID], outputKey)
				}
			}
			continue
		}

		if sourceNode.Type == "input_node" {
			if !isExecutionNode(targetNode) {
				return nil, fmt.Errorf("input edge %s targets non-execution node %s", edge.ID, targetNode.ID)
			}
			if startNodeID != "" && startNodeID != targetNode.ID {
				return nil, fmt.Errorf("graph must have exactly one entry execution node")
			}
			startNodeID = targetNode.ID
			continue
		}

		if sourceNode.Type == "if_else_node" || sourceNode.Type == "while_node" {
			continue
		}

		if !isExecutionNode(sourceNode) {
			continue
		}

		if !isExecutionNode(targetNode) {
			return nil, fmt.Errorf("edge %s targets non-execution node %s", edge.ID, targetNode.ID)
		}

		if existing, ok := staticNextMappings[sourceNode.ID]; ok && existing != targetNode.ID {
			return nil, fmt.Errorf("node %s has multiple execution successors", sourceNode.ID)
		}

		staticNextMappings[sourceNode.ID] = targetNode.ID
	}

	for whileNodeID, outputKeys := range whileOutputKeyMappings {
		producerIDs := whileDoneProducerMappings[whileNodeID]
		if len(producerIDs) == 0 {
			return nil, fmt.Errorf("while node %s routes to an output node but has no execution producer wired to its done boundary", whileNodeID)
		}
		for _, producerID := range producerIDs {
			for _, outputKey := range outputKeys {
				outputMappings[producerID] = appendUnique(outputMappings[producerID], outputKey)
			}
		}
	}

	for _, edge := range g.Edges {
		if !isIfElseTrueEdge(edge) && !isIfElseFalseEdge(edge) && !isWhileLoopEdge(edge) && !isWhileDoneEdge(edge) {
			continue
		}

		sourceNode, ok := nodeByID[edge.Source]
		if !ok {
			return nil, fmt.Errorf("edge %s references unknown source node %s", edge.ID, edge.Source)
		}

		targetNode, ok := nodeByID[edge.Target]
		if !ok {
			return nil, fmt.Errorf("edge %s references unknown target node %s", edge.ID, edge.Target)
		}

		if isIfElseTrueEdge(edge) {
			targetAgentName, ok := targetNode.AgentName()
			if !ok {
				return nil, fmt.Errorf("if_else edge %s targets non-execution node %s", edge.ID, edge.Target)
			}
			trueAgentMappings[edge.Source] = targetAgentName
		}
		if isIfElseFalseEdge(edge) {
			targetAgentName, ok := targetNode.AgentName()
			if !ok {
				return nil, fmt.Errorf("if_else edge %s targets non-execution node %s", edge.ID, edge.Target)
			}
			falseAgentMappings[edge.Source] = targetAgentName
		}
		if sourceNode.Type == "while_node" && isWhileLoopEdge(edge) {
			targetAgentName, ok := targetNode.AgentName()
			if !ok {
				return nil, fmt.Errorf("while edge %s targets non-execution node %s", edge.ID, edge.Target)
			}
			loopAgentMappings[edge.Source] = targetAgentName
		}
		if sourceNode.Type == "while_node" && isWhileDoneEdge(edge) {
			if targetNode.Type == "output_node" {
				doneTerminalMappings[edge.Source] = true
				continue
			}
			targetAgentName, ok := targetNode.AgentName()
			if !ok {
				return nil, fmt.Errorf("while edge %s targets non-execution node %s", edge.ID, edge.Target)
			}
			doneAgentMappings[edge.Source] = targetAgentName
		}
	}

	if startNodeID == "" {
		return nil, fmt.Errorf("graph has no entry execution node")
	}

	// 1. Compile toolbox nodes so their toolsets can be attached to LLM nodes.
	for _, node := range g.Nodes {
		if node.Type == "toolbox" {
			fmt.Printf("[DEBUG] Compiling toolbox node: %s\n", node.ID)
			nc, ok := c.compilers[node.Type]
			if !ok {
				return nil, fmt.Errorf("no compiler registered for toolbox node type %s", node.Type)
			}

			res, err := nc.Compile(node, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to compile toolbox %s: %w", node.ID, err)
			}
			if toolsets, ok := res.([]tool.Toolset); ok {
				toolboxResults[node.ID] = toolsets
				fmt.Printf("[DEBUG] Toolbox %s compiled successfully with %d toolsets\n", node.ID, len(toolsets))
			}
		}
	}

	// 2. Attach compiled toolsets to the LLM nodes they feed.
	for _, edge := range g.Edges {
		if isToolboxEdge(edge) {
			if toolsets, ok := toolboxResults[edge.Source]; ok {
				toolsetMappings[edge.Target] = append(toolsetMappings[edge.Target], toolsets...)
				fmt.Printf("[DEBUG] Wired toolbox %s to node %s\n", edge.Source, edge.Target)
			}
		}
	}

	// 3. Compile execution nodes into child agents the graph runtime can call.
	compiledNodes := make(map[string]compiledExecutionNode)
	var subAgents []agent.Agent

	for _, node := range g.Nodes {
		if !isExecutionNode(node) {
			fmt.Printf("[DEBUG] Skipping node %s during execution pass (type: %s)\n", node.ID, node.Type)
			continue
		}

		fmt.Printf("[DEBUG] Compiling execution node: %s (type: %s)\n", node.ID, node.Type)
		nc, ok := c.compilers[node.Type]
		if !ok {
			return nil, fmt.Errorf("no compiler registered for node type %s", node.Type)
		}

		metadata := map[string]interface{}{
			"output_keys":   outputMappings[node.ID],
			"state_key":     defaultStateKey(node),
			"toolsets":      toolsetMappings[node.ID],
			"true_agent":    trueAgentMappings[node.ID],
			"false_agent":   falseAgentMappings[node.ID],
			"loop_agent":    loopAgentMappings[node.ID],
			"done_agent":    doneAgentMappings[node.ID],
			"done_terminal": doneTerminalMappings[node.ID],
		}

		res, err := nc.Compile(node, metadata)
		if err != nil {
			fmt.Printf("[DEBUG] Failed to compile node %s: %v\n", node.ID, err)
			return nil, fmt.Errorf("failed to compile node %s: %w", node.ID, err)
		}

		a, ok := res.(agent.Agent)
		if !ok {
			return nil, fmt.Errorf("node %s did not produce an agent.Agent", node.ID)
		}

		compiledNodes[node.ID] = compiledExecutionNode{
			id:                      node.ID,
			nodeType:                node.Type,
			agent:                   a,
			nextNodeID:              staticNextMappings[node.ID],
			stateKey:                defaultStateKey(node),
			outputKeys:              outputMappings[node.ID],
			allowTerminalNoTransfer: doneTerminalMappings[node.ID],
		}
		subAgents = append(subAgents, a)
		fmt.Printf("[DEBUG] Node %s compiled to agent successfully\n", node.ID)
	}

	if len(compiledNodes) == 0 {
		fmt.Printf("[DEBUG] No execution agents found in graph\n")
		return nil, fmt.Errorf("no execution nodes found in graph")
	}

	maxSteps := opts.MaxSteps
	if maxSteps <= 0 {
		maxSteps = maxInt(len(compiledNodes)*32, 512)
	}

	fmt.Printf("[DEBUG] Building graph runtime with %d execution agents\n", len(compiledNodes))
	return newGraphRuntimeAgent(g.Name, compiledGraph{
		startNodeID:       startNodeID,
		nodes:             compiledNodes,
		agentNameToNodeID: agentNameToNodeID,
		maxSteps:          maxSteps,
	}, subAgents)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
