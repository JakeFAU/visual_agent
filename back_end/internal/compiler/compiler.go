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
	// 1. Sort nodes topologically to ensure valid execution order
	sorted, err := c.sortNodes(g)
	if err != nil {
		return nil, fmt.Errorf("graph validation failed: %w", err)
	}

	// 2. Pre-pass: Identify data flow state keys and toolset connections
	outputMappings := make(map[string][]string)
	toolsetMappings := make(map[string][]tool.Toolset)
    
    // Intermediate storage for compiled toolsets
    toolboxResults := make(map[string][]tool.Toolset)

	for _, edge := range g.Edges {
        if edge.TargetPort == "in_toolbox" {
            // This is a toolset connection
            continue 
        }
		outputMappings[edge.Source] = append(outputMappings[edge.Source], edge.TargetPort)
	}

	// 3. First Pass: Compile toolbox nodes
	for _, node := range g.Nodes {
		if node.Type == "toolbox" {
			nc, ok := c.compilers[node.Type]
			if !ok { continue }
			
			res, err := nc.Compile(node, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to compile toolbox %s: %w", node.ID, err)
			}
			if toolsets, ok := res.([]tool.Toolset); ok {
				toolboxResults[node.ID] = toolsets
			}
		}
	}

	// 4. Map toolsets to LLM nodes
	for _, edge := range g.Edges {
		if edge.TargetPort == "in_toolbox" {
			if toolsets, ok := toolboxResults[edge.Source]; ok {
				toolsetMappings[edge.Target] = append(toolsetMappings[edge.Target], toolsets...)
			}
		}
	}

	// 5. Second Pass: Compile execution nodes (LLM, If/Else, etc.)
	var agents []agent.Agent
	for _, node := range sorted {
        // Skip toolbox nodes in execution path
        if node.Type == "toolbox" { continue }

		nc, ok := c.compilers[node.Type]
		if !ok { continue }

		metadata := map[string]interface{}{
			"output_keys": outputMappings[node.ID],
			"toolsets":    toolsetMappings[node.ID],
		}

		res, err := nc.Compile(node, metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to compile node %s: %w", node.ID, err)
		}
		
		if a, ok := res.(agent.Agent); ok {
			agents = append(agents, a)
		}
	}

	// 6. Wrap in a SequentialAgent
	if len(agents) == 1 {
		return agents[0], nil
	}

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
