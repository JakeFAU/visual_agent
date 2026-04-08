package compiler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"iter"

	"github.com/JakeFAU/visual_agent/internal/graph"
	"github.com/JakeFAU/visual_agent/internal/runtime"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/genai"
)

// MockModel satisfies the model.LLM interface for testing.
type MockModel struct{}

func (m *MockModel) Name() string { return "mock-model" }

func (m *MockModel) GenerateContent(_ context.Context, _ *model.LLMRequest, _ bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		yield(&model.LLMResponse{}, nil)
	}
}

func MockModelFactory(_ context.Context, _ string, _ *genai.ClientConfig) (model.LLM, error) {
	return &MockModel{}, nil
}

type CaptureCompiler struct {
	lastMetadata map[string]interface{}
}

type StubToolboxCompiler struct {
	toolsets []tool.Toolset
}

type BranchCaptureCompiler struct {
	lastMetadata map[string]interface{}
}

type LoopTestCompiler struct {
	runCounts map[string]int
}

type JSONBranchTestCompiler struct {
	runCounts map[string]int
}

func (c *CaptureCompiler) Compile(node graph.Node, metadata map[string]interface{}) (any, error) {
	c.lastMetadata = metadata
	return agent.New(agent.Config{Name: node.ID})
}

func (c *StubToolboxCompiler) Compile(_ graph.Node, _ map[string]interface{}) (any, error) {
	return c.toolsets, nil
}

func (c *BranchCaptureCompiler) Compile(node graph.Node, metadata map[string]interface{}) (any, error) {
	c.lastMetadata = metadata
	return agent.New(agent.Config{Name: node.ID})
}

func (c *LoopTestCompiler) Compile(node graph.Node, metadata map[string]interface{}) (any, error) {
	cfg, ok := node.Config.(graph.LLMNodeConfig)
	if !ok {
		return nil, fmt.Errorf("expected llm config, got %T", node.Config)
	}

	stateKey, _ := metadata["state_key"].(string)
	if c.runCounts == nil {
		c.runCounts = make(map[string]int)
	}

	return agent.New(agent.Config{
		Name: cfg.Name,
		Run: func(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				c.runCounts[node.ID]++

				event := session.NewEvent(ctx.InvocationID())
				switch cfg.Name {
				case "analyze":
					status := "retry"
					if c.runCounts[node.ID] >= 2 {
						status = "pass"
					}
					event.Actions.StateDelta = map[string]any{
						stateKey: map[string]any{
							"status":    status,
							"iteration": c.runCounts[node.ID],
						},
					}
				case "retry":
					event.Actions.StateDelta = map[string]any{
						stateKey: fmt.Sprintf("retry-%d", c.runCounts[node.ID]),
					}
				case "done":
					event.Actions.StateDelta = map[string]any{
						stateKey: "Failures resolved",
					}
				default:
					event.Actions.StateDelta = map[string]any{
						stateKey: cfg.Name,
					}
				}

				yield(event, nil)
			}
		},
	})
}

func (c *JSONBranchTestCompiler) Compile(node graph.Node, metadata map[string]interface{}) (any, error) {
	cfg, ok := node.Config.(graph.LLMNodeConfig)
	if !ok {
		return nil, fmt.Errorf("expected llm config, got %T", node.Config)
	}

	stateKey, _ := metadata["state_key"].(string)
	if c.runCounts == nil {
		c.runCounts = make(map[string]int)
	}

	return agent.New(agent.Config{
		Name: cfg.Name,
		Run: func(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				c.runCounts[cfg.Name]++

				event := session.NewEvent(ctx.InvocationID())
				switch cfg.Name {
				case "classify_ticket":
					event.Actions.StateDelta = map[string]any{
						stateKey: `{"route":"billing","reason":"classified as payroll/billing"}`,
					}
				default:
					event.Actions.StateDelta = map[string]any{
						stateKey: cfg.Name,
					}
				}

				yield(event, nil)
			}
		},
	})
}

type testToolset struct {
	name string
}

type errToolset struct {
	name string
	err  error
}

func (t *testToolset) Name() string {
	return t.name
}

func (t *testToolset) Tools(agent.ReadonlyContext) ([]tool.Tool, error) {
	return nil, nil
}

func (t *errToolset) Name() string {
	return t.name
}

func (t *errToolset) Tools(agent.ReadonlyContext) ([]tool.Tool, error) {
	return nil, t.err
}

func TestCompileSequential(t *testing.T) {
	graphJSON := `{
  "version": "1.0",
  "name": "sequential_workflow",
  "nodes": [
    {
      "id": "input-1",
      "type": "input_node",
      "position": { "x": -250, "y": 0 },
      "config": {
        "name": "user_input",
        "description": "User input"
      }
    },
    {
      "id": "llm-1",
      "type": "llm_node",
      "position": { "x": 0, "y": 0 },
      "config": {
        "name": "agent_1",
        "description": "First agent",
        "model": "gemini-2.0-flash",
        "instruction": "Hello",
        "response_mode": "text"
      }
    },
    {
      "id": "llm-2",
      "type": "llm_node",
      "position": { "x": 300, "y": 0 },
      "config": {
        "name": "agent_2",
        "description": "Second agent",
        "model": "gemini-2.0-flash",
        "instruction": "World",
        "response_mode": "text"
      }
    }
  ],
  "edges": [
    {
      "id": "input-edge",
      "source": "input-1",
      "source_port": "message",
      "target": "llm-1",
      "target_port": "message"
    },
    {
      "id": "edge-1",
      "source": "llm-1",
      "source_port": "message",
      "target": "llm-2",
      "target_port": "message"
    }
  ]
}`

	var g graph.Graph
	if err := json.Unmarshal([]byte(graphJSON), &g); err != nil {
		t.Fatalf("Failed to unmarshal graph: %v", err)
	}

	c := New()
	c.Register("llm_node", &LLMNodeCompiler{NewModel: MockModelFactory})

	compiled, err := c.Compile(g)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	if compiled == nil {
		t.Fatal("Expected compiled agent, got nil")
	}
}

func TestCompileIfElse(t *testing.T) {
	ifElseJSON := `{
  "version": "1.0",
  "name": "branching_workflow",
  "nodes": [
    {
      "id": "input-1",
      "type": "input_node",
      "config": {
        "name": "user_input",
        "description": "User input"
      }
    },
    {
      "id": "if-1",
      "type": "if_else_node",
      "config": {
        "condition_language": "CEL",
        "condition": "state.category == 'billing'"
      }
    },
    {
      "id": "true-branch",
      "type": "llm_node",
      "config": { "name": "billing_agent" }
    },
    {
      "id": "false-branch",
      "type": "llm_node",
      "config": { "name": "general_agent" }
    }
  ],
  "edges": [
    { "id": "input", "source": "input-1", "source_port": "message", "target": "if-1", "target_port": "message" },
    { "id": "e1", "source": "if-1", "source_port": "out_true", "target": "true-branch" },
    { "id": "e2", "source": "if-1", "source_port": "out_false", "target": "false-branch" }
  ]
}`

	var g graph.Graph
	if err := json.Unmarshal([]byte(ifElseJSON), &g); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	c := New()
	c.Register("llm_node", &LLMNodeCompiler{NewModel: MockModelFactory})
	c.Register("if_else_node", &IfElseNodeCompiler{})

	compiled, err := c.Compile(g)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	if compiled == nil {
		t.Fatal("Expected compiled agent, got nil")
	}
}

func TestCompileIfElseReadsStructuredJSONState(t *testing.T) {
	g := graph.Graph{
		Version: "1.0",
		Name:    "structured_branching_workflow",
		Nodes: []graph.Node{
			{
				ID:   "input-1",
				Type: "input_node",
				Config: graph.InputNodeConfig{
					Name:        "user_input",
					Description: "User input",
				},
			},
			{
				ID:   "classify",
				Type: "llm_node",
				Config: graph.LLMNodeConfig{
					Name:         "classify_ticket",
					Model:        "mock-model",
					Instruction:  "classify",
					ResponseMode: "json",
				},
			},
			{
				ID:   "route-gate",
				Type: "if_else_node",
				Config: graph.IfElseNodeConfig{
					ConditionLanguage: "CEL",
					Condition:         `state.classify_ticket.route == "billing"`,
				},
			},
			{
				ID:   "billing-node",
				Type: "llm_node",
				Config: graph.LLMNodeConfig{
					Name:         "billing_agent",
					Model:        "mock-model",
					Instruction:  "billing",
					ResponseMode: "text",
				},
			},
			{
				ID:   "general-node",
				Type: "llm_node",
				Config: graph.LLMNodeConfig{
					Name:         "general_agent",
					Model:        "mock-model",
					Instruction:  "general",
					ResponseMode: "text",
				},
			},
		},
		Edges: []graph.Edge{
			{ID: "e-input", Source: "input-1", SourcePort: "message", Target: "classify", TargetPort: "message"},
			{ID: "e-classify-gate", Source: "classify", SourcePort: "message", Target: "route-gate", TargetPort: "message"},
			{ID: "e-true", Source: "route-gate", SourcePort: "message:true", Target: "billing-node", TargetPort: "message"},
			{ID: "e-false", Source: "route-gate", SourcePort: "message:false", Target: "general-node", TargetPort: "message"},
		},
	}

	c := New()
	branchCompiler := &JSONBranchTestCompiler{}
	c.Register("llm_node", branchCompiler)
	c.Register("if_else_node", &IfElseNodeCompiler{})

	compiled, err := c.Compile(g)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	rt := runtime.NewLocalRuntime()
	for event, err := range rt.Execute(context.Background(), compiled, "Where is my paycheck?") {
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		if event == nil {
			continue
		}
	}

	if branchCompiler.runCounts["billing_agent"] != 1 {
		t.Fatalf("billing branch run count mismatch: got %d want 1", branchCompiler.runCounts["billing_agent"])
	}
	if branchCompiler.runCounts["general_agent"] != 0 {
		t.Fatalf("general branch run count mismatch: got %d want 0", branchCompiler.runCounts["general_agent"])
	}
}

func TestCompileExecutesControlLoop(t *testing.T) {
	g := graph.Graph{
		Version: "1.0",
		Name:    "loop_workflow",
		Nodes: []graph.Node{
			{
				ID:   "input-1",
				Type: "input_node",
				Config: graph.InputNodeConfig{
					Name:        "user_input",
					Description: "User input",
				},
			},
			{
				ID:   "analyze",
				Type: "llm_node",
				Config: graph.LLMNodeConfig{
					Name:         "analyze",
					Model:        "mock-model",
					Instruction:  "analyze",
					ResponseMode: "json",
				},
			},
			{
				ID:   "gate",
				Type: "if_else_node",
				Config: graph.IfElseNodeConfig{
					ConditionLanguage: "CEL",
					Condition:         `state.analyze.status == "pass"`,
				},
			},
			{
				ID:   "retry",
				Type: "llm_node",
				Config: graph.LLMNodeConfig{
					Name:         "retry",
					Model:        "mock-model",
					Instruction:  "retry",
					ResponseMode: "text",
				},
			},
			{
				ID:   "done",
				Type: "llm_node",
				Config: graph.LLMNodeConfig{
					Name:         "done",
					Model:        "mock-model",
					Instruction:  "done",
					ResponseMode: "text",
				},
			},
			{
				ID:   "output-1",
				Type: "output_node",
				Config: graph.OutputNodeConfig{
					Name:      "final_output",
					OutputKey: "result",
					Format:    "message",
				},
			},
		},
		Edges: []graph.Edge{
			{ID: "e-input", Source: "input-1", SourcePort: "message", Target: "analyze", TargetPort: "message"},
			{ID: "e-analyze-gate", Source: "analyze", SourcePort: "message", Target: "gate", TargetPort: "message"},
			{ID: "e-gate-true", Source: "gate", SourcePort: "message:true", Target: "done", TargetPort: "message"},
			{ID: "e-gate-false", Source: "gate", SourcePort: "message:false", Target: "retry", TargetPort: "message"},
			{ID: "e-retry-loop", Source: "retry", SourcePort: "message", Target: "analyze", TargetPort: "message"},
			{ID: "e-done-output", Source: "done", SourcePort: "message", Target: "output-1", TargetPort: "message"},
		},
	}

	c := New()
	loopCompiler := &LoopTestCompiler{}
	c.Register("llm_node", loopCompiler)
	c.Register("if_else_node", &IfElseNodeCompiler{})

	compiled, err := c.Compile(g)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	rt := runtime.NewLocalRuntime()
	var gotResult string
	for event, err := range rt.Execute(context.Background(), compiled, "fix it") {
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		if event == nil {
			continue
		}
		if value, ok := event.Actions.StateDelta["result"].(string); ok {
			gotResult = value
		}
	}

	if gotResult != "Failures resolved" {
		t.Fatalf("final output mismatch: got %q want %q", gotResult, "Failures resolved")
	}

	if loopCompiler.runCounts["analyze"] != 2 {
		t.Fatalf("analyze run count mismatch: got %d want 2", loopCompiler.runCounts["analyze"])
	}

	if loopCompiler.runCounts["retry"] != 1 {
		t.Fatalf("retry run count mismatch: got %d want 1", loopCompiler.runCounts["retry"])
	}

	if loopCompiler.runCounts["done"] != 1 {
		t.Fatalf("done run count mismatch: got %d want 1", loopCompiler.runCounts["done"])
	}
}

func TestCompileExecutesWhileLoop(t *testing.T) {
	g := graph.Graph{
		Version: "1.0",
		Name:    "while_workflow",
		Nodes: []graph.Node{
			{
				ID:   "input-1",
				Type: "input_node",
				Config: graph.InputNodeConfig{
					Name:        "user_input",
					Description: "User input",
				},
			},
			{
				ID:   "analyze",
				Type: "llm_node",
				Config: graph.LLMNodeConfig{
					Name:         "analyze",
					Model:        "mock-model",
					Instruction:  "analyze",
					ResponseMode: "json",
				},
			},
			{
				ID:   "loop-gate",
				Type: "while_node",
				Config: graph.WhileNodeConfig{
					Condition:     `state.analyze.status != "pass"`,
					MaxIterations: 3,
				},
			},
			{
				ID:   "retry",
				Type: "llm_node",
				Config: graph.LLMNodeConfig{
					Name:         "retry",
					Model:        "mock-model",
					Instruction:  "retry",
					ResponseMode: "text",
				},
			},
			{
				ID:   "done",
				Type: "llm_node",
				Config: graph.LLMNodeConfig{
					Name:         "done",
					Model:        "mock-model",
					Instruction:  "done",
					ResponseMode: "text",
				},
			},
			{
				ID:   "output-1",
				Type: "output_node",
				Config: graph.OutputNodeConfig{
					Name:      "final_output",
					OutputKey: "result",
					Format:    "message",
				},
			},
		},
		Edges: []graph.Edge{
			{ID: "e-input", Source: "input-1", SourcePort: "message", Target: "analyze", TargetPort: "message"},
			{ID: "e-analyze-while", Source: "analyze", SourcePort: "message", Target: "loop-gate", TargetPort: "message"},
			{ID: "e-while-loop", Source: "loop-gate", SourcePort: "message:loop", Target: "retry", TargetPort: "message"},
			{ID: "e-while-done", Source: "loop-gate", SourcePort: "message:done", Target: "done", TargetPort: "message"},
			{ID: "e-retry-loop", Source: "retry", SourcePort: "message", Target: "analyze", TargetPort: "message"},
			{ID: "e-done-output", Source: "done", SourcePort: "message", Target: "output-1", TargetPort: "message"},
		},
	}

	c := New()
	loopCompiler := &LoopTestCompiler{}
	c.Register("llm_node", loopCompiler)
	c.Register("while_node", &WhileNodeCompiler{})

	compiled, err := c.Compile(g)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	rt := runtime.NewLocalRuntime()
	var gotResult string
	for event, err := range rt.Execute(context.Background(), compiled, "fix it") {
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		if event == nil {
			continue
		}
		if value, ok := event.Actions.StateDelta["result"].(string); ok {
			gotResult = value
		}
	}

	if gotResult != "Failures resolved" {
		t.Fatalf("final output mismatch: got %q want %q", gotResult, "Failures resolved")
	}

	if loopCompiler.runCounts["analyze"] != 2 {
		t.Fatalf("analyze run count mismatch: got %d want 2", loopCompiler.runCounts["analyze"])
	}

	if loopCompiler.runCounts["retry"] != 1 {
		t.Fatalf("retry run count mismatch: got %d want 1", loopCompiler.runCounts["retry"])
	}
}

func TestCompileExecutesWhileLoopWithContainerOutput(t *testing.T) {
	g := graph.Graph{
		Version: "1.0",
		Name:    "while_output_workflow",
		Nodes: []graph.Node{
			{
				ID:   "input-1",
				Type: "input_node",
				Config: graph.InputNodeConfig{
					Name:        "user_input",
					Description: "User input",
				},
			},
			{
				ID:   "analyze",
				Type: "llm_node",
				Config: graph.LLMNodeConfig{
					Name:         "analyze",
					Model:        "mock-model",
					Instruction:  "analyze",
					ResponseMode: "json",
				},
			},
			{
				ID:   "loop-gate",
				Type: "while_node",
				Config: graph.WhileNodeConfig{
					Condition:     `state.analyze.status != "pass"`,
					MaxIterations: 3,
				},
			},
			{
				ID:   "retry",
				Type: "llm_node",
				Config: graph.LLMNodeConfig{
					Name:         "retry",
					Model:        "mock-model",
					Instruction:  "retry",
					ResponseMode: "text",
				},
			},
			{
				ID:   "output-1",
				Type: "output_node",
				Config: graph.OutputNodeConfig{
					Name:      "final_output",
					OutputKey: "result",
					Format:    "message",
				},
			},
		},
		Edges: []graph.Edge{
			{ID: "e-input", Source: "input-1", SourcePort: "message", Target: "analyze", TargetPort: "message"},
			{ID: "e-analyze-out", Source: "analyze", SourcePort: "message", Target: "loop-gate", TargetPort: "message:done"},
			{ID: "e-while-loop", Source: "loop-gate", SourcePort: "message:loop", Target: "retry", TargetPort: "message"},
			{ID: "e-while-done", Source: "loop-gate", SourcePort: "message:done", Target: "output-1", TargetPort: "message"},
			{ID: "e-retry-loop", Source: "retry", SourcePort: "message", Target: "analyze", TargetPort: "message"},
		},
	}

	c := New()
	loopCompiler := &LoopTestCompiler{}
	c.Register("llm_node", loopCompiler)
	c.Register("while_node", &WhileNodeCompiler{})

	compiled, err := c.Compile(g)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	rt := runtime.NewLocalRuntime()
	var gotResult map[string]any
	for event, err := range rt.Execute(context.Background(), compiled, "fix it") {
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		if event == nil {
			continue
		}
		if value, ok := event.Actions.StateDelta["result"].(map[string]any); ok {
			gotResult = value
		}
	}

	if gotResult == nil {
		t.Fatal("expected while container output to populate result")
	}
	if gotResult["status"] != "pass" {
		t.Fatalf("result status mismatch: got %v want %q", gotResult["status"], "pass")
	}
	if gotResult["iteration"] != 2 {
		t.Fatalf("result iteration mismatch: got %v want %d", gotResult["iteration"], 2)
	}

	if loopCompiler.runCounts["analyze"] != 2 {
		t.Fatalf("analyze run count mismatch: got %d want 2", loopCompiler.runCounts["analyze"])
	}
	if loopCompiler.runCounts["retry"] != 1 {
		t.Fatalf("retry run count mismatch: got %d want 1", loopCompiler.runCounts["retry"])
	}
}

func TestCompileExecutesWhileLoopWithContainerOutputUsingReturnAlias(t *testing.T) {
	g := graph.Graph{
		Version: "1.0",
		Name:    "while_output_return_alias_workflow",
		Nodes: []graph.Node{
			{
				ID:   "input-1",
				Type: "input_node",
				Config: graph.InputNodeConfig{
					Name:        "user_input",
					Description: "User input",
				},
			},
			{
				ID:   "analyze",
				Type: "llm_node",
				Config: graph.LLMNodeConfig{
					Name:         "analyze",
					Model:        "mock-model",
					Instruction:  "analyze",
					ResponseMode: "json",
				},
			},
			{
				ID:   "loop-gate",
				Type: "while_node",
				Config: graph.WhileNodeConfig{
					Condition:     `state.analyze.status != "pass"`,
					MaxIterations: 3,
				},
			},
			{
				ID:   "retry",
				Type: "llm_node",
				Config: graph.LLMNodeConfig{
					Name:         "retry",
					Model:        "mock-model",
					Instruction:  "retry",
					ResponseMode: "text",
				},
			},
			{
				ID:   "output-1",
				Type: "output_node",
				Config: graph.OutputNodeConfig{
					Name:      "final_output",
					OutputKey: "result",
					Format:    "message",
				},
			},
		},
		Edges: []graph.Edge{
			{ID: "e-input", Source: "input-1", SourcePort: "message", Target: "analyze", TargetPort: "message"},
			{ID: "e-analyze-out", Source: "analyze", SourcePort: "message", Target: "loop-gate", TargetPort: "message:return"},
			{ID: "e-while-loop", Source: "loop-gate", SourcePort: "message:loop", Target: "retry", TargetPort: "message"},
			{ID: "e-while-done", Source: "loop-gate", SourcePort: "message:done", Target: "output-1", TargetPort: "message"},
			{ID: "e-retry-loop", Source: "retry", SourcePort: "message", Target: "analyze", TargetPort: "message"},
		},
	}

	c := New()
	loopCompiler := &LoopTestCompiler{}
	c.Register("llm_node", loopCompiler)
	c.Register("while_node", &WhileNodeCompiler{})

	compiled, err := c.Compile(g)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	rt := runtime.NewLocalRuntime()
	var gotResult map[string]any
	for event, err := range rt.Execute(context.Background(), compiled, "fix it") {
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		if event == nil {
			continue
		}
		if value, ok := event.Actions.StateDelta["result"].(map[string]any); ok {
			gotResult = value
		}
	}

	if gotResult == nil {
		t.Fatal("expected while container output to populate result")
	}
	if gotResult["status"] != "pass" {
		t.Fatalf("result status mismatch: got %v want %q", gotResult["status"], "pass")
	}
	if gotResult["iteration"] != 2 {
		t.Fatalf("result iteration mismatch: got %v want %d", gotResult["iteration"], 2)
	}

	if loopCompiler.runCounts["analyze"] != 2 {
		t.Fatalf("analyze run count mismatch: got %d want 2", loopCompiler.runCounts["analyze"])
	}
	if loopCompiler.runCounts["retry"] != 1 {
		t.Fatalf("retry run count mismatch: got %d want 1", loopCompiler.runCounts["retry"])
	}
}

func TestCompileWhileNodeEnforcesLocalIterationBudget(t *testing.T) {
	g := graph.Graph{
		Version: "1.0",
		Name:    "while_budget_workflow",
		Nodes: []graph.Node{
			{
				ID:   "input-1",
				Type: "input_node",
				Config: graph.InputNodeConfig{
					Name:        "user_input",
					Description: "User input",
				},
			},
			{
				ID:   "analyze",
				Type: "llm_node",
				Config: graph.LLMNodeConfig{
					Name:         "analyze",
					Model:        "mock-model",
					Instruction:  "analyze",
					ResponseMode: "json",
				},
			},
			{
				ID:   "loop-gate",
				Type: "while_node",
				Config: graph.WhileNodeConfig{
					Condition:     `true`,
					MaxIterations: 1,
				},
			},
			{
				ID:   "retry",
				Type: "llm_node",
				Config: graph.LLMNodeConfig{
					Name:         "retry",
					Model:        "mock-model",
					Instruction:  "retry",
					ResponseMode: "text",
				},
			},
		},
		Edges: []graph.Edge{
			{ID: "e-input", Source: "input-1", SourcePort: "message", Target: "analyze", TargetPort: "message"},
			{ID: "e-analyze-while", Source: "analyze", SourcePort: "message", Target: "loop-gate", TargetPort: "message"},
			{ID: "e-while-loop", Source: "loop-gate", SourcePort: "message:loop", Target: "retry", TargetPort: "message"},
			{ID: "e-while-done", Source: "loop-gate", SourcePort: "message:done", Target: "retry", TargetPort: "message"},
			{ID: "e-retry-loop", Source: "retry", SourcePort: "message", Target: "analyze", TargetPort: "message"},
		},
	}

	c := New()
	loopCompiler := &LoopTestCompiler{}
	c.Register("llm_node", loopCompiler)
	c.Register("while_node", &WhileNodeCompiler{})

	compiled, err := c.Compile(g)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	rt := runtime.NewLocalRuntime()
	var gotErr error
	for _, err := range rt.Execute(context.Background(), compiled, "fix it") {
		if err != nil {
			gotErr = err
			break
		}
	}

	if gotErr == nil {
		t.Fatal("expected while max_iterations error, got nil")
	}

	if want := `while node "loop-gate" exceeded max_iterations 1`; !strings.Contains(gotErr.Error(), want) {
		t.Fatalf("while budget error mismatch: got %q want substring %q", gotErr.Error(), want)
	}
}

func TestCompileWithOptionsEnforcesStepBudget(t *testing.T) {
	g := graph.Graph{
		Version: "1.0",
		Name:    "loop_budget_workflow",
		Nodes: []graph.Node{
			{
				ID:   "input-1",
				Type: "input_node",
				Config: graph.InputNodeConfig{
					Name:        "user_input",
					Description: "User input",
				},
			},
			{
				ID:   "analyze",
				Type: "llm_node",
				Config: graph.LLMNodeConfig{
					Name:         "analyze",
					Model:        "mock-model",
					Instruction:  "analyze",
					ResponseMode: "json",
				},
			},
			{
				ID:   "gate",
				Type: "if_else_node",
				Config: graph.IfElseNodeConfig{
					ConditionLanguage: "CEL",
					Condition:         `state.analyze.status == "pass"`,
				},
			},
			{
				ID:   "retry",
				Type: "llm_node",
				Config: graph.LLMNodeConfig{
					Name:         "retry",
					Model:        "mock-model",
					Instruction:  "retry",
					ResponseMode: "text",
				},
			},
		},
		Edges: []graph.Edge{
			{ID: "e-input", Source: "input-1", SourcePort: "message", Target: "analyze", TargetPort: "message"},
			{ID: "e-analyze-gate", Source: "analyze", SourcePort: "message", Target: "gate", TargetPort: "message"},
			{ID: "e-gate-false", Source: "gate", SourcePort: "message:false", Target: "retry", TargetPort: "message"},
			{ID: "e-gate-true", Source: "gate", SourcePort: "message:true", Target: "retry", TargetPort: "message"},
			{ID: "e-retry-loop", Source: "retry", SourcePort: "message", Target: "analyze", TargetPort: "message"},
		},
	}

	c := New()
	loopCompiler := &LoopTestCompiler{}
	c.Register("llm_node", loopCompiler)
	c.Register("if_else_node", &IfElseNodeCompiler{})

	compiled, err := c.CompileWithOptions(g, CompileOptions{MaxSteps: 3})
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	rt := runtime.NewLocalRuntime()
	var gotErr error
	for _, err := range rt.Execute(context.Background(), compiled, "fix it") {
		if err != nil {
			gotErr = err
			break
		}
	}

	if gotErr == nil {
		t.Fatal("expected step budget error, got nil")
	}

	if want := "possible infinite loop"; !strings.Contains(gotErr.Error(), want) {
		t.Fatalf("step budget error mismatch: got %q want substring %q", gotErr.Error(), want)
	}
}

func TestCompileUsesOutputNodeKey(t *testing.T) {
	graphJSON := `{
  "version": "1.0",
  "name": "output_key_workflow",
  "nodes": [
    {
      "id": "input-1",
      "type": "input_node",
      "position": { "x": -300, "y": 0 },
      "config": {
        "name": "user_input",
        "description": "User input"
      }
    },
    {
      "id": "llm-1",
      "type": "llm_node",
      "position": { "x": 0, "y": 0 },
      "config": {
        "name": "agent_1",
        "description": "Agent",
        "model": "gemini-2.0-flash",
        "instruction": "Hello",
        "response_mode": "text",
        "generate_content_config": {}
      }
    },
    {
      "id": "output-1",
      "type": "output_node",
      "position": { "x": 300, "y": 0 },
      "config": {
        "name": "final_output",
        "output_key": "result",
        "format": "message"
      }
    }
  ],
  "edges": [
    {
      "id": "input-edge",
      "source": "input-1",
      "source_port": "message",
      "target": "llm-1",
      "target_port": "message",
      "data_type": "message",
      "edge_kind": "data_flow"
    },
    {
      "id": "edge-1",
      "source": "llm-1",
      "source_port": "message",
      "target": "output-1",
      "target_port": "message",
      "data_type": "message",
      "edge_kind": "data_flow"
    }
  ]
}`

	var g graph.Graph
	if err := json.Unmarshal([]byte(graphJSON), &g); err != nil {
		t.Fatalf("Failed to unmarshal graph: %v", err)
	}

	c := New()
	capture := &CaptureCompiler{}
	c.Register("llm_node", capture)

	compiled, err := c.Compile(g)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	if compiled == nil {
		t.Fatal("Expected compiled agent, got nil")
	}

	got, ok := capture.lastMetadata["output_keys"].([]string)
	if !ok {
		t.Fatalf("Expected output_keys metadata, got %#v", capture.lastMetadata["output_keys"])
	}

	want := []string{"result"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("output_keys mismatch: got %v want %v", got, want)
	}
}

func TestCompileWiresToolboxHandleToLLMNode(t *testing.T) {
	graphJSON := `{
  "version": "1.0",
  "name": "toolbox_wiring_workflow",
  "nodes": [
    {
      "id": "input-1",
      "type": "input_node",
      "config": {
        "name": "user_input",
        "description": "User input"
      }
    },
    {
      "id": "toolbox-1",
      "type": "toolbox",
      "config": {
        "tools": [],
        "mcp_servers": [],
        "custom_functions": []
      }
    },
    {
      "id": "llm-1",
      "type": "llm_node",
      "config": {
        "name": "agent_1"
      }
    }
  ],
  "edges": [
    {
      "id": "edge-0",
      "source": "input-1",
      "source_port": "message",
      "target": "llm-1",
      "target_port": "message",
      "data_type": "message",
      "edge_kind": "data_flow"
    },
    {
      "id": "edge-1",
      "source": "toolbox-1",
      "source_port": "toolbox_handle",
      "target": "llm-1",
      "target_port": "toolbox_handle",
      "data_type": "toolbox_handle",
      "edge_kind": "data_flow"
    }
  ]
}`

	var g graph.Graph
	if err := json.Unmarshal([]byte(graphJSON), &g); err != nil {
		t.Fatalf("Failed to unmarshal graph: %v", err)
	}

	c := New()
	capture := &CaptureCompiler{}
	c.Register("llm_node", capture)
	c.Register("toolbox", &StubToolboxCompiler{
		toolsets: []tool.Toolset{&testToolset{name: "stub-toolset"}},
	})

	compiled, err := c.Compile(g)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	if compiled == nil {
		t.Fatal("Expected compiled agent, got nil")
	}

	got, ok := capture.lastMetadata["toolsets"].([]tool.Toolset)
	if !ok {
		t.Fatalf("Expected toolsets metadata, got %#v", capture.lastMetadata["toolsets"])
	}

	if len(got) != 1 {
		t.Fatalf("toolsets length mismatch: got %d want 1", len(got))
	}

	if got[0].Name() != "stub-toolset" {
		t.Fatalf("toolset mismatch: got %q want %q", got[0].Name(), "stub-toolset")
	}
}

func TestToolboxNodeCompilerBuildsGoogleSearchToolset(t *testing.T) {
	node := graph.Node{
		ID:   "toolbox-1",
		Type: "toolbox",
		Config: graph.ToolboxNodeConfig{
			Tools:           []string{"google_search"},
			MCPServers:      []graph.MCPServerConfig{},
			CustomFunctions: []graph.CustomFunctionConfig{},
		},
	}

	compiler := &ToolboxNodeCompiler{}
	res, err := compiler.Compile(node, nil)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	toolsets, ok := res.([]tool.Toolset)
	if !ok {
		t.Fatalf("Expected []tool.Toolset, got %#v", res)
	}

	if len(toolsets) != 1 {
		t.Fatalf("toolsets length mismatch: got %d want 1", len(toolsets))
	}

	tools, err := toolsets[0].Tools(nil)
	if err != nil {
		t.Fatalf("Tools() failed: %v", err)
	}

	if len(tools) != 1 {
		t.Fatalf("tools length mismatch: got %d want 1", len(tools))
	}

	if tools[0].Name() != "google_search" {
		t.Fatalf("tool mismatch: got %q want %q", tools[0].Name(), "google_search")
	}
}

func TestToolboxNodeCompilerRejectsUnsupportedBuiltInTool(t *testing.T) {
	node := graph.Node{
		ID:   "toolbox-1",
		Type: "toolbox",
		Config: graph.ToolboxNodeConfig{
			Tools:           []string{"code_interpreter"},
			MCPServers:      []graph.MCPServerConfig{},
			CustomFunctions: []graph.CustomFunctionConfig{},
		},
	}

	compiler := &ToolboxNodeCompiler{}
	if _, err := compiler.Compile(node, nil); err == nil {
		t.Fatal("Expected error for unsupported built-in tool, got nil")
	}
}

func TestAnnotatedMCPToolsetIncludesStderrInErrors(t *testing.T) {
	stderrTail := newStderrTailBuffer(1024)
	_, _ = stderrTail.Write([]byte("Error accessing directory /missing: ENOENT"))

	toolset := &annotatedMCPToolset{
		inner: &errToolset{
			name: "mcp_tool_set",
			err:  errors.New("failed to list MCP tools: failed to init MCP session: calling \"initialize\": EOF"),
		},
		serverName: "filesystem",
		command:    "npx",
		args:       []string{"-y", "@modelcontextprotocol/server-filesystem", "/missing"},
		stderrTail: stderrTail,
	}

	_, err := toolset.Tools(nil)
	if err == nil {
		t.Fatal("Expected toolset error, got nil")
	}

	errText := err.Error()
	if !strings.Contains(errText, `mcp server "filesystem" command: npx -y @modelcontextprotocol/server-filesystem /missing`) {
		t.Fatalf("expected command details in error, got %q", errText)
	}
	if !strings.Contains(errText, "Error accessing directory /missing: ENOENT") {
		t.Fatalf("expected stderr details in error, got %q", errText)
	}
}

func TestToolboxNodeCompilerRejectsMissingMCPCommand(t *testing.T) {
	node := graph.Node{
		ID:   "toolbox-1",
		Type: "toolbox",
		Config: graph.ToolboxNodeConfig{
			MCPServers: []graph.MCPServerConfig{{
				Name:    "missing",
				Command: "definitely-not-a-real-command",
			}},
		},
	}

	compiler := &ToolboxNodeCompiler{}
	_, err := compiler.Compile(node, nil)
	if err == nil {
		t.Fatal("Expected missing command error, got nil")
	}
	if !strings.Contains(err.Error(), `mcp server missing command "definitely-not-a-real-command" is not available`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompileIfElseWiresBranchTargets(t *testing.T) {
	graphJSON := `{
  "version": "1.0",
  "name": "if_else_wiring",
  "nodes": [
    {
      "id": "input-1",
      "type": "input_node",
      "config": {
        "name": "user_input",
        "description": "User input"
      }
    },
    {
      "id": "if-1",
      "type": "if_else_node",
      "config": {
        "condition_language": "CEL",
        "condition": "state.category == 'billing'"
      }
    },
    {
      "id": "llm-true",
      "type": "llm_node",
      "config": {
        "name": "billing_agent",
        "model": "gemini-2.0-flash",
        "instruction": "billing",
        "response_mode": "text"
      }
    },
    {
      "id": "llm-false",
      "type": "llm_node",
      "config": {
        "name": "general_agent",
        "model": "gemini-2.0-flash",
        "instruction": "general",
        "response_mode": "text"
      }
    }
  ],
  "edges": [
    {
      "id": "e0",
      "source": "input-1",
      "source_port": "message",
      "target": "if-1",
      "target_port": "message"
    },
    {
      "id": "e1",
      "source": "if-1",
      "source_port": "message:true",
      "target": "llm-true",
      "target_port": "message"
    },
    {
      "id": "e2",
      "source": "if-1",
      "source_port": "message:false",
      "target": "llm-false",
      "target_port": "message"
    }
  ]
}`

	var g graph.Graph
	if err := json.Unmarshal([]byte(graphJSON), &g); err != nil {
		t.Fatalf("Failed to unmarshal graph: %v", err)
	}

	c := New()
	capture := &BranchCaptureCompiler{}
	c.Register("if_else_node", capture)
	c.Register("llm_node", &CaptureCompiler{})

	compiled, err := c.Compile(g)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	if compiled == nil {
		t.Fatal("Expected compiled agent, got nil")
	}

	if got := capture.lastMetadata["true_agent"]; got != "billing_agent" {
		t.Fatalf("true_agent mismatch: got %#v want %q", got, "billing_agent")
	}

	if got := capture.lastMetadata["false_agent"]; got != "general_agent" {
		t.Fatalf("false_agent mismatch: got %#v want %q", got, "general_agent")
	}
}
