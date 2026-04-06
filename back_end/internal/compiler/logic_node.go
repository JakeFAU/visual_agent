package compiler

import (
	"errors"
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

const whileIterationStateKeyPrefix = "__while_iteration__"

// Compile turns a while_node into a control agent that evaluates its CEL
// condition, enforces the loop-local iteration cap, and transfers execution to
// either the loop body or the done branch.
func (c *WhileNodeCompiler) Compile(node graph.Node, metadata map[string]interface{}) (any, error) {
	cfg, ok := node.Config.(graph.WhileNodeConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config for while_node")
	}

	loopAgent, _ := metadata["loop_agent"].(string)
	doneAgent, _ := metadata["done_agent"].(string)
	doneTerminal, _ := metadata["done_terminal"].(bool)
	if loopAgent == "" || (doneAgent == "" && !doneTerminal) {
		return nil, fmt.Errorf("while_node requires a loop branch and either a done branch target or terminal done exit")
	}

	env, err := cel.NewEnv(
		cel.Variable("state", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("iteration", cel.IntType),
		cel.Variable("max_iterations", cel.IntType),
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

	iterationKey := whileIterationStateKey(node.ID)
	return agent.New(agent.Config{
		Name: node.ID,
		Run: func(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				iteration, err := whileIterationCount(ctx.Session().State(), iterationKey)
				if err != nil {
					yield(nil, fmt.Errorf("failed to read while iteration counter for %q: %w", node.ID, err))
					return
				}

				state := sessionStateToMap(ctx.Session().State())
				out, _, err := program.Eval(map[string]interface{}{
					"state":          state,
					"iteration":      iteration,
					"max_iterations": cfg.MaxIterations,
				})
				if err != nil {
					yield(nil, fmt.Errorf("CEL eval error: %w", err))
					return
				}

				shouldContinue, ok := out.Value().(bool)
				if !ok {
					yield(nil, fmt.Errorf("condition did not return a boolean"))
					return
				}

				if !shouldContinue {
					if doneTerminal && doneAgent == "" {
						return
					}
					yield(&session.Event{
						Actions: session.EventActions{
							TransferToAgent: doneAgent,
						},
					}, nil)
					return
				}

				if iteration >= int64(cfg.MaxIterations) {
					yield(nil, fmt.Errorf("while node %q exceeded max_iterations %d", node.ID, cfg.MaxIterations))
					return
				}

				if err := ctx.Session().State().Set(iterationKey, iteration+1); err != nil {
					yield(nil, fmt.Errorf("failed to update while iteration counter for %q: %w", node.ID, err))
					return
				}

				event := session.NewEvent(ctx.InvocationID())
				event.Actions.TransferToAgent = loopAgent
				yield(event, nil)
			}
		},
	})
}

func whileIterationStateKey(nodeID string) string {
	return whileIterationStateKeyPrefix + nodeID
}

func whileIterationCount(state session.ReadonlyState, key string) (int64, error) {
	if state == nil {
		return 0, nil
	}

	value, err := state.Get(key)
	if err != nil {
		if errors.Is(err, session.ErrStateKeyNotExist) {
			return 0, nil
		}
		return 0, err
	}

	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("unexpected iteration counter type %T", value)
	}
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
