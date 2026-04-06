package compiler

import (
	"errors"
	"fmt"
	"iter"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
)

const (
	graphRuntimeAuthor        = "graph_runtime"
	graphRuntimeDiagnosticKey = "graph_runtime_diagnostic"
)

type compiledGraph struct {
	startNodeID       string
	nodes             map[string]compiledExecutionNode
	agentNameToNodeID map[string]string
	maxSteps          int
}

type compiledExecutionNode struct {
	id                      string
	nodeType                string
	agent                   agent.Agent
	nextNodeID              string
	stateKey                string
	outputKeys              []string
	allowTerminalNoTransfer bool
}

func newGraphRuntimeAgent(name string, graph compiledGraph, subAgents []agent.Agent) (agent.Agent, error) {
	return agent.New(agent.Config{
		Name:      name,
		SubAgents: subAgents,
		Run: func(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				currentNodeID := graph.startNodeID
				visitCounts := make(map[string]int)

				for step := 1; currentNodeID != ""; step++ {
					if graph.maxSteps > 0 && step > graph.maxSteps {
						yield(nil, fmt.Errorf("graph exceeded %d execution steps; possible infinite loop", graph.maxSteps))
						return
					}

					node, ok := graph.nodes[currentNodeID]
					if !ok {
						yield(nil, fmt.Errorf("graph routed to unknown node %q", currentNodeID))
						return
					}

					visitCounts[node.id]++
					if !yield(graphRuntimeDiagnosticEvent(ctx, map[string]any{
						"kind":        "node_enter",
						"step":        step,
						"node_id":     node.id,
						"node_type":   node.nodeType,
						"agent_name":  node.agent.Name(),
						"visit_count": visitCounts[node.id],
					}), nil) {
						return
					}

					transferTarget := ""
					for event, err := range node.agent.Run(ctx) {
						if err != nil {
							yield(nil, err)
							return
						}
						if event == nil {
							continue
						}

						if err := applyStateDelta(ctx.Session().State(), event.Actions.StateDelta); err != nil {
							yield(nil, fmt.Errorf("failed to apply state delta for node %q: %w", node.id, err))
							return
						}

						if event.Actions.TransferToAgent != "" {
							transferTarget = event.Actions.TransferToAgent
						}

						if !yield(event, nil) {
							return
						}

						if event.Actions.Escalate {
							return
						}
					}

					if ok, err := emitOutputAliases(ctx, node, yield); err != nil {
						yield(nil, err)
						return
					} else if !ok {
						return
					}

					nextNodeID, err := graph.nextNodeID(node, transferTarget)
					if err != nil {
						yield(nil, err)
						return
					}

					if nextNodeID != "" {
						nextNode := graph.nodes[nextNodeID]
						if !yield(graphRuntimeDiagnosticEvent(ctx, map[string]any{
							"kind":            "transition",
							"step":            step,
							"from_node_id":    node.id,
							"from_agent_name": node.agent.Name(),
							"to_node_id":      nextNodeID,
							"to_agent_name":   nextNode.agent.Name(),
							"reason":          transitionReason(node, transferTarget),
						}), nil) {
							return
						}
					}
					currentNodeID = nextNodeID
				}
			}
		},
	})
}

func (g compiledGraph) nextNodeID(node compiledExecutionNode, transferTarget string) (string, error) {
	if transferTarget != "" {
		nextNodeID, ok := g.agentNameToNodeID[transferTarget]
		if !ok {
			return "", fmt.Errorf("node %q transferred to unknown agent %q", node.id, transferTarget)
		}
		return nextNodeID, nil
	}

	if node.nodeType == "while_node" && node.allowTerminalNoTransfer {
		return "", nil
	}

	if node.nodeType == "if_else_node" || node.nodeType == "while_node" {
		return "", fmt.Errorf("%s %q did not choose a branch", node.nodeType, node.id)
	}

	return node.nextNodeID, nil
}

func emitOutputAliases(
	ctx agent.InvocationContext,
	node compiledExecutionNode,
	yield func(*session.Event, error) bool,
) (bool, error) {
	if node.stateKey == "" || len(node.outputKeys) == 0 {
		return true, nil
	}

	value, err := ctx.Session().State().Get(node.stateKey)
	if err != nil {
		if errors.Is(err, session.ErrStateKeyNotExist) {
			return true, nil
		}
		return false, fmt.Errorf("failed to read state key %q for node %q: %w", node.stateKey, node.id, err)
	}

	delta := make(map[string]any)
	for _, outputKey := range node.outputKeys {
		if outputKey == "" || outputKey == node.stateKey {
			continue
		}
		if err := ctx.Session().State().Set(outputKey, value); err != nil {
			return false, fmt.Errorf("failed to write output key %q for node %q: %w", outputKey, node.id, err)
		}
		delta[outputKey] = value
	}

	if len(delta) == 0 {
		return true, nil
	}

	event := session.NewEvent(ctx.InvocationID())
	event.Author = node.agent.Name()
	event.Actions.StateDelta = delta
	return yield(event, nil), nil
}

func applyStateDelta(state session.State, delta map[string]any) error {
	if state == nil || len(delta) == 0 {
		return nil
	}

	for key, value := range delta {
		if err := state.Set(key, value); err != nil {
			return err
		}
	}

	return nil
}

func graphRuntimeDiagnosticEvent(ctx agent.InvocationContext, payload map[string]any) *session.Event {
	event := session.NewEvent(ctx.InvocationID())
	event.Author = graphRuntimeAuthor
	event.LLMResponse = model.LLMResponse{
		CustomMetadata: map[string]any{
			graphRuntimeDiagnosticKey: payload,
		},
	}
	return event
}

func transitionReason(node compiledExecutionNode, transferTarget string) string {
	if transferTarget != "" && node.nodeType == "if_else_node" {
		return "branch"
	}
	if transferTarget != "" && node.nodeType == "while_node" {
		return "loop_control"
	}
	if transferTarget != "" {
		return "transfer"
	}
	return "edge"
}
