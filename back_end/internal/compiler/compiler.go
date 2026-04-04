package compiler

import (
	"fmt"
	"github.com/JakeFAU/visual_agent/internal/graph"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/workflowagents/sequentialagent"
)

// NodeCompiler translates a specific graph node into its ADK representation
type NodeCompiler interface {
	Compile(node graph.Node, metadata map[string]interface{}) (agent.Agent, error)
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
	// 1. Sort nodes topologically to ensure no cycles and valid execution order
	sorted, err := c.sortNodes(g)
	if err != nil {
		return nil, fmt.Errorf("graph validation failed: %w", err)
	}

	// 2. Map edges to determine data flow state keys
	// nodeID -> list of output keys it should set
	outputMappings := make(map[string][]string)
	for _, edge := range g.Edges {
		outputMappings[edge.Source] = append(outputMappings[edge.Source], edge.TargetPort)
	}

	// 3. Compile individual agents
	var agents []agent.Agent
	for _, node := range sorted {
		nc, ok := c.compilers[node.Type]
		if !ok {
			// Skip nodes like input/output for now if not registered
			continue
		}

		metadata := map[string]interface{}{
			"output_keys": outputMappings[node.ID],
		}

		compiled, err := nc.Compile(node, metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to compile node %s: %w", node.ID, err)
		}
		agents = append(agents, compiled)
	}

	// 4. Wrap in a SequentialAgent for now (simplest pattern)
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
