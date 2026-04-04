package compiler

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/JakeFAU/visual_agent/internal/graph"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
	"iter"
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

func (c *CaptureCompiler) Compile(node graph.Node, metadata map[string]interface{}) (any, error) {
	c.lastMetadata = metadata
	return agent.New(agent.Config{Name: node.ID})
}

func TestCompileSequential(t *testing.T) {
	graphJSON := `{
  "version": "1.0",
  "name": "sequential_workflow",
  "nodes": [
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
      "id": "edge-1",
      "source": "llm-1",
      "source_port": "out_message",
      "target": "llm-2",
      "target_port": "in_message"
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

func TestCompileCycle(t *testing.T) {
	cycleJSON := `{
  "nodes": [
    {"id": "n1", "type": "llm_node", "config": {"name": "a1"}},
    {"id": "n2", "type": "llm_node", "config": {"name": "a2"}}
  ],
  "edges": [
    {"id": "e1", "source": "n1", "target": "n2"},
    {"id": "e2", "source": "n2", "target": "n1"}
  ]
}`
	var g graph.Graph
	_ = json.Unmarshal([]byte(cycleJSON), &g)

	c := New()
	c.Register("llm_node", &LLMNodeCompiler{NewModel: MockModelFactory})

	_, err := c.Compile(g)
	if err == nil {
		t.Fatal("Expected error for cyclic graph, got nil")
	}
}

func TestCompileUsesOutputNodeKey(t *testing.T) {
	graphJSON := `{
  "version": "1.0",
  "name": "output_key_workflow",
  "nodes": [
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
