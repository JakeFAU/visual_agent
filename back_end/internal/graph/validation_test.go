package graph

import "testing"

func TestGraphValidateAcceptsSupportedGraph(t *testing.T) {
	g := Graph{
		Version: SupportedGraphVersion,
		Name:    "Valid Graph",
		Nodes: []Node{
			{
				ID:   "if-1",
				Type: "if_else_node",
				Config: IfElseNodeConfig{
					ConditionLanguage: "CEL",
					Condition:         `state.category == "billing"`,
				},
			},
			{
				ID:   "llm-true",
				Type: "llm_node",
				Config: LLMNodeConfig{
					Name:         "billing_agent",
					Model:        "gemini-2.5-flash",
					Instruction:  "Handle billing",
					ResponseMode: "text",
				},
			},
			{
				ID:   "llm-false",
				Type: "llm_node",
				Config: LLMNodeConfig{
					Name:         "general_agent",
					Model:        "gemini-2.5-flash",
					Instruction:  "Handle general",
					ResponseMode: "text",
				},
			},
		},
		Edges: []Edge{
			{ID: "e1", Source: "if-1", SourcePort: "message:true", Target: "llm-true", TargetPort: "message"},
			{ID: "e2", Source: "if-1", SourcePort: "message:false", Target: "llm-false", TargetPort: "message"},
		},
	}

	if err := g.Validate(); err != nil {
		t.Fatalf("expected graph to validate, got %v", err)
	}
}

func TestGraphValidateRejectsJSONPathConditions(t *testing.T) {
	g := Graph{
		Version: SupportedGraphVersion,
		Name:    "Invalid IfElse",
		Nodes: []Node{
			{
				ID:   "if-1",
				Type: "if_else_node",
				Config: IfElseNodeConfig{
					ConditionLanguage: "JSONPath",
					Condition:         "$.category == 'billing'",
				},
			},
		},
	}

	if err := g.Validate(); err == nil {
		t.Fatal("expected JSONPath validation error, got nil")
	}
}

func TestGraphValidateRejectsUnsupportedWhileNode(t *testing.T) {
	g := Graph{
		Version: SupportedGraphVersion,
		Name:    "Invalid While",
		Nodes: []Node{
			{
				ID:   "while-1",
				Type: "while_node",
				Config: WhileNodeConfig{
					Condition:     "state.counter < 3",
					MaxIterations: 3,
				},
			},
		},
	}

	if err := g.Validate(); err == nil {
		t.Fatal("expected while_node validation error, got nil")
	}
}

func TestGraphValidateRejectsUnsupportedBuiltInTool(t *testing.T) {
	g := Graph{
		Version: SupportedGraphVersion,
		Name:    "Invalid Toolbox",
		Nodes: []Node{
			{
				ID:   "toolbox-1",
				Type: "toolbox",
				Config: ToolboxNodeConfig{
					Tools: []string{"code_interpreter"},
				},
			},
		},
	}

	if err := g.Validate(); err == nil {
		t.Fatal("expected unsupported tool validation error, got nil")
	}
}

func TestGraphValidateRejectsDuplicateAgentNames(t *testing.T) {
	g := Graph{
		Version: SupportedGraphVersion,
		Name:    "Duplicate Agents",
		Nodes: []Node{
			{
				ID:   "llm-1",
				Type: "llm_node",
				Config: LLMNodeConfig{
					Name:         "same_name",
					Model:        "gemini-2.5-flash",
					Instruction:  "One",
					ResponseMode: "text",
				},
			},
			{
				ID:   "llm-2",
				Type: "llm_node",
				Config: LLMNodeConfig{
					Name:         "same_name",
					Model:        "gemini-2.5-flash",
					Instruction:  "Two",
					ResponseMode: "text",
				},
			},
		},
	}

	if err := g.Validate(); err == nil {
		t.Fatal("expected duplicate agent name validation error, got nil")
	}
}
