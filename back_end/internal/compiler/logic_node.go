package compiler

import (
	"fmt"
	"github.com/JakeFAU/visual_agent/internal/graph"
	"github.com/google/cel-go/cel"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/workflowagents/loopagent"
	"google.golang.org/adk/session"
	"iter"
)

type IfElseNodeCompiler struct{}

func (c *IfElseNodeCompiler) Compile(node graph.Node, metadata map[string]interface{}) (any, error) {
	cfg, ok := node.Config.(graph.IfElseNodeConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config for if_else_node")
	}

    trueAgent, _ := metadata["true_agent"].(string)
    falseAgent, _ := metadata["false_agent"].(string)

	return agent.New(agent.Config{
		Name: node.ID,
		Run: func(_ agent.InvocationContext) iter.Seq2[*session.Event, error] {			return func(yield func(*session.Event, error) bool) {
				e, err := cel.NewEnv(
					cel.Variable("state", cel.MapType(cel.StringType, cel.AnyType)),
				)
				if err != nil {
					yield(nil, fmt.Errorf("failed to create CEL env: %w", err))
					return
				}

				ast, issues := e.Compile(cfg.Condition)
				if issues.Err() != nil {
					yield(nil, fmt.Errorf("CEL compile error: %w", issues.Err()))
					return
				}

				program, err := e.Program(ast)
				if err != nil {
					yield(nil, fmt.Errorf("failed to create program: %w", err))
					return
				}

				state := ctx.Session().State()
				out, _, err := program.Eval(map[string]interface{}{
					"state": state,
				})
				if err != nil {
					yield(nil, fmt.Errorf("CEL eval error: %w", err))
					return
				}

				result, ok := out.Value().(bool)
				if !ok {
					yield(nil, fmt.Errorf("condition did not return a boolean"))
					return
				}

				target := falseAgent
				if result {
					target = trueAgent
				}

				yield(&session.Event{
					Actions: session.EventActions{
						TransferToAgent: target,
					},
				}, nil)
			}
		},
	})
}

type WhileNodeCompiler struct{}

func (c *WhileNodeCompiler) Compile(node graph.Node, _ map[string]interface{}) (any, error) {
	cfg, ok := node.Config.(graph.WhileNodeConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config for while_node")
	}

	// Body agents would be compiled by the walker and passed here
	// For v0 we'll use a dummy condition agent
	condAgent, err := agent.New(agent.Config{
		Name: node.ID + "_cond",
		Run: func(_ agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				// Simplified loop logic: always escalate to finish loop for now
				yield(&session.Event{
					Actions: session.EventActions{
						Escalate: true,
					},
				}, nil)
			}
		},
	})
	if err != nil {
		return nil, err
	}

	return loopagent.New(loopagent.Config{
		AgentConfig: agent.Config{
			Name:      node.ID,
			SubAgents: []agent.Agent{condAgent},
		},
		MaxIterations: uint(cfg.MaxIterations),
	})
}
