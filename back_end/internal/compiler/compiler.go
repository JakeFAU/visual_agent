package compiler

import (
	"fmt"
	"github.com/JakeFAU/visual_agent/internal/graph"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/workflowagents/sequentialagent"
	"google.golang.org/adk/tool"
)

// NodeCompiler translates a specific graph node into its ADK representation
// interface{} (any) allows nodes to return agents, toolsets, or other configs.
type NodeCompiler interface {
	Compile(node graph.Node, metadata map[string]interface{}) (any, error)
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

func (c *Compiler) Compile(g graph.Graph) (agent.Agent, error) {
	fmt.Printf("[DEBUG] Starting compilation for graph: %s\n", g.Name)

	// 1. Sort nodes topologically to ensure valid execution order
	sorted, err := c.sortNodes(g)
	if err != nil {
		fmt.Printf("[DEBUG] Sort failed: %v\n", err)
		return nil, fmt.Errorf("graph validation failed: %w", err)
	}
	fmt.Printf("[DEBUG] Topologically sorted %d nodes\n", len(sorted))

	// 2. Pre-pass: Identify data flow state keys and toolset connections
	outputMappings := make(map[string][]string)
	toolsetMappings := make(map[string][]tool.Toolset)
	toolboxResults := make(map[string][]tool.Toolset)

	for _, edge := range g.Edges {
		if edge.TargetPort == "in_toolbox" {
			continue
		}
		outputMappings[edge.Source] = append(outputMappings[edge.Source], edge.TargetPort)
	}

	// 3. First Pass: Compile toolbox nodes
	for _, node := range g.Nodes {
		if node.Type == "toolbox" {
			fmt.Printf("[DEBUG] Compiling toolbox node: %s\n", node.ID)
			nc, ok := c.compilers[node.Type]
			if !ok {
				fmt.Printf("[DEBUG] No compiler for node type: %s\n", node.Type)
				continue
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

	// 4. Map toolsets to LLM nodes
	for _, edge := range g.Edges {
		if edge.TargetPort == "in_toolbox" {
			if toolsets, ok := toolboxResults[edge.Source]; ok {
				toolsetMappings[edge.Target] = append(toolsetMappings[edge.Target], toolsets...)
				fmt.Printf("[DEBUG] Wired toolbox %s to node %s\n", edge.Source, edge.Target)
			}
		}
	}

	// 5. Second Pass: Compile execution nodes (LLM, If/Else, etc.)
	var agents []agent.Agent
	for _, node := range sorted {
		if node.Type == "toolbox" || node.Type == "input_node" || node.Type == "output_node" {
			fmt.Printf("[DEBUG] Skipping node %s during execution pass (type: %s)\n", node.ID, node.Type)
			continue
		}

		fmt.Printf("[DEBUG] Compiling execution node: %s (type: %s)\n", node.ID, node.Type)
		nc, ok := c.compilers[node.Type]
		if !ok {
			fmt.Printf("[DEBUG] WARNING: No compiler registered for node type: %s\n", node.Type)
			continue
		}

		metadata := map[string]interface{}{
			"output_keys": outputMappings[node.ID],
			"toolsets":    toolsetMappings[node.ID],
		}

		res, err := nc.Compile(node, metadata)
		if err != nil {
			fmt.Printf("[DEBUG] Failed to compile node %s: %v\n", node.ID, err)
			return nil, fmt.Errorf("failed to compile node %s: %w", node.ID, err)
		}

		if a, ok := res.(agent.Agent); ok {
			agents = append(agents, a)
			fmt.Printf("[DEBUG] Node %s compiled to agent successfully\n", node.ID)
		} else {
			fmt.Printf("[DEBUG] WARNING: Node %s did not produce an agent.Agent\n", node.ID)
		}
	}

	// 6. Wrap in a SequentialAgent
	if len(agents) == 0 {
		fmt.Printf("[DEBUG] No execution agents found in graph\n")
		return nil, fmt.Errorf("no execution nodes found in graph")
	}

	if len(agents) == 1 {
		fmt.Printf("[DEBUG] Returning single agent for graph\n")
		return agents[0], nil
	}

	fmt.Printf("[DEBUG] Wrapping %d agents in SequentialAgent\n", len(agents))
	return sequentialagent.New(sequentialagent.Config{
		AgentConfig: agent.Config{
			Name:      g.Name,
			SubAgents: agents,
		},
	})
}

func (c *Compiler) sortNodes(g graph.Graph) ([]graph.Node, error) {
	adj := make(map[string][]string)
	inDegree := make(map[string]int)
	nodeMap := make(map[string]graph.Node)

	for _, node := range g.Nodes {
		inDegree[node.ID] = 0
		nodeMap[node.ID] = node
	}

	for _, edge := range g.Edges {
		adj[edge.Source] = append(adj[edge.Source], edge.Target)
		inDegree[edge.Target]++
	}

	var queue []string
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	var sorted []graph.Node
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		sorted = append(sorted, nodeMap[u])

		for _, v := range adj[u] {
			inDegree[v]--
			if inDegree[v] == 0 {
				queue = append(queue, v)
			}
		}
	}

	if len(sorted) != len(g.Nodes) {
		return nil, fmt.Errorf("cycle detected in graph")
	}

	return sorted, nil
}
