package compiler

import (
	"encoding/json"
	"testing"

	"github.com/JakeFAU/visual_agent/internal/graph"
)

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
	c.Register("llm_node", &LLMNodeCompiler{})

	compiled, err := c.Compile(g)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	if compiled == nil {
		t.Fatal("Expected compiled agent, got nil")
	}
    
    t.Logf("Compiled agent type: %T", compiled)
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
	c.Register("llm_node", &LLMNodeCompiler{})

	_, err := c.Compile(g)
	if err == nil {
		t.Fatal("Expected error for cyclic graph, got nil")
	}
	t.Logf("Caught expected error: %v", err)
}
