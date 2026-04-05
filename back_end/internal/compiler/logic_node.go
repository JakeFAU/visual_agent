package compiler

import (
	"fmt"
	"github.com/JakeFAU/visual_agent/internal/graph"
	"github.com/google/cel-go/cel"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/session"
	"iter"
)

type IfElseNodeCompiler struct{}

// Compile turns an if_else_node into a lightweight ADK agent whose only job is
// to evaluate the configured CEL expression and transfer control to the chosen
// downstream agent.
func (c *IfElseNodeCompiler) Compile(node graph.Node, metadata map[string]interface{}) (any, error) {
	cfg, ok := node.Config.(graph.IfElseNodeConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config for if_else_node")
	}
	if cfg.ConditionLanguage != "CEL" {
		return nil, fmt.Errorf("unsupported condition language %q", cfg.ConditionLanguage)
	}

	trueAgent, _ := metadata["true_agent"].(string)
	falseAgent, _ := metadata["false_agent"].(string)
	if trueAgent == "" || falseAgent == "" {
		return nil, fmt.Errorf("if_else_node requires both true and false branch targets")
	}

	env, err := cel.NewEnv(
		cel.Variable("state", cel.MapType(cel.StringType, cel.AnyType)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL env: %w", err)
	}

	ast, issues := env.Compile(cfg.Condition)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("CEL compile error: %w", issues.Err())
	}

	program, err := env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL program: %w", err)
	}

	return agent.New(agent.Config{
		Name: node.ID,
		Run: func(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				state := sessionStateToMap(ctx.Session().State())
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

// Compile rejects while_node usage until loop semantics are implemented in the
// backend.
func (c *WhileNodeCompiler) Compile(node graph.Node, _ map[string]interface{}) (any, error) {
	if _, ok := node.Config.(graph.WhileNodeConfig); !ok {
		return nil, fmt.Errorf("invalid config for while_node")
	}
	return nil, fmt.Errorf("while_node is not supported in v0")
}

// sessionStateToMap copies ADK session state into a regular Go map so it can be
// passed into CEL evaluation.
func sessionStateToMap(state session.ReadonlyState) map[string]any {
	values := make(map[string]any)
	if state == nil {
		return values
	}
	for key, value := range state.All() {
		values[key] = value
	}
	return values
}
