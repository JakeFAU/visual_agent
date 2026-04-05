package compiler

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/JakeFAU/visual_agent/internal/graph"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
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

type StubToolboxCompiler struct {
	toolsets []tool.Toolset
}

type BranchCaptureCompiler struct {
	lastMetadata map[string]interface{}
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

func TestCompileWiresToolboxHandleToLLMNode(t *testing.T) {
	graphJSON := `{
  "version": "1.0",
  "name": "toolbox_wiring_workflow",
  "nodes": [
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
