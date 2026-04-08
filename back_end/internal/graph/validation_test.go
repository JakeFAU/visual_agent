package graph

import "testing"

func TestGraphValidateAcceptsSupportedGraph(t *testing.T) {
	g := Graph{
		Version: SupportedGraphVersion,
		Name:    "Valid Graph",
		Nodes: []Node{
			{
				ID:   "input-1",
				Type: "input_node",
				Config: InputNodeConfig{
					Name:        "user_input",
					Description: "User input",
				},
			},
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
			{ID: "input", Source: "input-1", SourcePort: "message", Target: "if-1", TargetPort: "message"},
			{ID: "e1", Source: "if-1", SourcePort: "message:true", Target: "llm-true", TargetPort: "message"},
			{ID: "e2", Source: "if-1", SourcePort: "message:false", Target: "llm-false", TargetPort: "message"},
		},
	}

	if err := g.Validate(); err != nil {
		t.Fatalf("expected graph to validate, got %v", err)
	}
}

func TestGraphValidateRejectsJSONPathConditions(t *testing.T) {
	if err := (Node{
		ID:   "if-1",
		Type: "if_else_node",
		Config: IfElseNodeConfig{
			ConditionLanguage: "JSONPath",
			Condition:         "$.category == 'billing'",
		},
	}).Validate(); err == nil {
		t.Fatal("expected JSONPath validation error, got nil")
	}
}

func TestGraphValidateAcceptsWhileNode(t *testing.T) {
	g := Graph{
		Version: SupportedGraphVersion,
		Name:    "Valid While",
		Nodes: []Node{
			{
				ID:   "input-1",
				Type: "input_node",
				Config: InputNodeConfig{
					Name:        "user_input",
					Description: "User input",
				},
			},
			{
				ID:   "while-1",
				Type: "while_node",
				Config: WhileNodeConfig{
					Condition:     "state.counter < 3",
					MaxIterations: 3,
				},
			},
			{
				ID:   "loop-body",
				Type: "llm_node",
				Config: LLMNodeConfig{
					Name:         "loop_body",
					Model:        "gemini-2.5-flash",
					Instruction:  "Loop",
					ResponseMode: "text",
				},
			},
			{
				ID:   "done",
				Type: "llm_node",
				Config: LLMNodeConfig{
					Name:         "done",
					Model:        "gemini-2.5-flash",
					Instruction:  "Done",
					ResponseMode: "text",
				},
			},
		},
		Edges: []Edge{
			{ID: "e-input", Source: "input-1", SourcePort: "message", Target: "while-1", TargetPort: "message"},
			{ID: "e-loop", Source: "while-1", SourcePort: "message:loop", Target: "loop-body", TargetPort: "message"},
			{ID: "e-done", Source: "while-1", SourcePort: "message:done", Target: "done", TargetPort: "message"},
			{ID: "e-back", Source: "loop-body", SourcePort: "message", Target: "while-1", TargetPort: "message"},
		},
	}

	if err := g.Validate(); err != nil {
		t.Fatalf("expected while_node graph to validate, got %v", err)
	}
}

func TestGraphValidateAcceptsWhileDoneBranchToOutputNode(t *testing.T) {
	g := Graph{
		Version: SupportedGraphVersion,
		Name:    "While Container Output",
		Nodes: []Node{
			{
				ID:   "input-1",
				Type: "input_node",
				Config: InputNodeConfig{
					Name:        "user_input",
					Description: "User input",
				},
			},
			{
				ID:   "analyze",
				Type: "llm_node",
				Config: LLMNodeConfig{
					Name:         "analyze",
					Model:        "gemini-2.5-flash",
					Instruction:  "Analyze",
					ResponseMode: "json",
				},
			},
			{
				ID:   "while-1",
				Type: "while_node",
				Config: WhileNodeConfig{
					Condition:     `state.analyze.status != "pass"`,
					MaxIterations: 3,
				},
			},
			{
				ID:   "retry",
				Type: "llm_node",
				Config: LLMNodeConfig{
					Name:         "retry",
					Model:        "gemini-2.5-flash",
					Instruction:  "Retry",
					ResponseMode: "text",
				},
			},
			{
				ID:   "output-1",
				Type: "output_node",
				Config: OutputNodeConfig{
					Name:      "final_output",
					OutputKey: "result",
					Format:    "message",
				},
			},
		},
		Edges: []Edge{
			{ID: "e-input", Source: "input-1", SourcePort: "message", Target: "analyze", TargetPort: "message"},
			{ID: "e-analyze-done", Source: "analyze", SourcePort: "message", Target: "while-1", TargetPort: "message:done"},
			{ID: "e-loop", Source: "while-1", SourcePort: "message:loop", Target: "retry", TargetPort: "message"},
			{ID: "e-repeat", Source: "retry", SourcePort: "message", Target: "analyze", TargetPort: "message"},
			{ID: "e-done", Source: "while-1", SourcePort: "message:done", Target: "output-1", TargetPort: "message"},
		},
	}

	if err := g.Validate(); err != nil {
		t.Fatalf("expected while output graph to validate, got %v", err)
	}
}

func TestGraphValidateAcceptsWhileDoneBranchToOutputNodeWithReturnAlias(t *testing.T) {
	g := Graph{
		Version: SupportedGraphVersion,
		Name:    "While Container Output Return Alias",
		Nodes: []Node{
			{
				ID:   "input-1",
				Type: "input_node",
				Config: InputNodeConfig{
					Name:        "user_input",
					Description: "User input",
				},
			},
			{
				ID:   "analyze",
				Type: "llm_node",
				Config: LLMNodeConfig{
					Name:         "analyze",
					Model:        "gemini-2.5-flash",
					Instruction:  "Analyze",
					ResponseMode: "json",
				},
			},
			{
				ID:   "while-1",
				Type: "while_node",
				Config: WhileNodeConfig{
					Condition:     `state.analyze.status != "pass"`,
					MaxIterations: 3,
				},
			},
			{
				ID:   "retry",
				Type: "llm_node",
				Config: LLMNodeConfig{
					Name:         "retry",
					Model:        "gemini-2.5-flash",
					Instruction:  "Retry",
					ResponseMode: "text",
				},
			},
			{
				ID:   "output-1",
				Type: "output_node",
				Config: OutputNodeConfig{
					Name:      "final_output",
					OutputKey: "result",
					Format:    "message",
				},
			},
		},
		Edges: []Edge{
			{ID: "e-input", Source: "input-1", SourcePort: "message", Target: "analyze", TargetPort: "message"},
			{ID: "e-analyze-done", Source: "analyze", SourcePort: "message", Target: "while-1", TargetPort: "message:return"},
			{ID: "e-loop", Source: "while-1", SourcePort: "message:loop", Target: "retry", TargetPort: "message"},
			{ID: "e-repeat", Source: "retry", SourcePort: "message", Target: "analyze", TargetPort: "message"},
			{ID: "e-done", Source: "while-1", SourcePort: "message:done", Target: "output-1", TargetPort: "message"},
		},
	}

	if err := g.Validate(); err != nil {
		t.Fatalf("expected while output graph using message:return to validate, got %v", err)
	}
}

func TestGraphValidateRejectsWhileNodeWithoutPositiveMaxIterations(t *testing.T) {
	g := Graph{
		Version: SupportedGraphVersion,
		Name:    "Invalid While",
		Nodes: []Node{
			{
				ID:   "while-1",
				Type: "while_node",
				Config: WhileNodeConfig{
					Condition:     "true",
					MaxIterations: 0,
				},
			},
		},
	}

	if err := g.Validate(); err == nil {
		t.Fatal("expected while_node validation error, got nil")
	}
}

func TestGraphValidateAcceptsNodeInsideWhileContainer(t *testing.T) {
	g := Graph{
		Version: SupportedGraphVersion,
		Name:    "Nested Loop Body",
		Nodes: []Node{
			{
				ID:   "input-1",
				Type: "input_node",
				Config: InputNodeConfig{
					Name:        "user_input",
					Description: "User input",
				},
			},
			{
				ID:   "while-1",
				Type: "while_node",
				Config: WhileNodeConfig{
					Condition:     "state.review.status != 'pass'",
					MaxIterations: 3,
				},
			},
			{
				ID:       "review",
				Type:     "llm_node",
				ParentID: "while-1",
				Config: LLMNodeConfig{
					Name:         "review",
					Model:        "gemini-2.5-flash",
					Instruction:  "Review",
					ResponseMode: "text",
				},
			},
			{
				ID:   "done",
				Type: "llm_node",
				Config: LLMNodeConfig{
					Name:         "done",
					Model:        "gemini-2.5-flash",
					Instruction:  "Done",
					ResponseMode: "text",
				},
			},
		},
		Edges: []Edge{
			{ID: "e-input", Source: "input-1", SourcePort: "message", Target: "while-1", TargetPort: "message"},
			{ID: "e-loop", Source: "while-1", SourcePort: "message:loop", Target: "review", TargetPort: "message"},
			{ID: "e-repeat", Source: "review", SourcePort: "message", Target: "while-1", TargetPort: "message:return"},
			{ID: "e-done", Source: "while-1", SourcePort: "message:done", Target: "done", TargetPort: "message"},
		},
	}

	if err := g.Validate(); err != nil {
		t.Fatalf("expected parented graph to validate, got %v", err)
	}
}

func TestGraphValidateRejectsUnknownParent(t *testing.T) {
	g := Graph{
		Version: SupportedGraphVersion,
		Name:    "Unknown Parent",
		Nodes: []Node{
			{
				ID:       "review",
				Type:     "llm_node",
				ParentID: "missing",
				Config: LLMNodeConfig{
					Name:         "review",
					Model:        "gemini-2.5-flash",
					Instruction:  "Review",
					ResponseMode: "text",
				},
			},
		},
	}

	if err := g.Validate(); err == nil {
		t.Fatal("expected unknown parent validation error, got nil")
	}
}

func TestGraphValidateRejectsParentThatIsNotWhileContainer(t *testing.T) {
	g := Graph{
		Version: SupportedGraphVersion,
		Name:    "Invalid Parent Type",
		Nodes: []Node{
			{
				ID:   "llm-1",
				Type: "llm_node",
				Config: LLMNodeConfig{
					Name:         "a",
					Model:        "gemini-2.5-flash",
					Instruction:  "A",
					ResponseMode: "text",
				},
			},
			{
				ID:       "llm-2",
				Type:     "llm_node",
				ParentID: "llm-1",
				Config: LLMNodeConfig{
					Name:         "b",
					Model:        "gemini-2.5-flash",
					Instruction:  "B",
					ResponseMode: "text",
				},
			},
		},
	}

	if err := g.Validate(); err == nil {
		t.Fatal("expected invalid parent type error, got nil")
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

func TestGraphValidateAcceptsReachableControlLoop(t *testing.T) {
	g := Graph{
		Version: SupportedGraphVersion,
		Name:    "Retry Loop",
		Nodes: []Node{
			{
				ID:   "input-1",
				Type: "input_node",
				Config: InputNodeConfig{
					Name:        "user_input",
					Description: "User input",
				},
			},
			{
				ID:   "analyze",
				Type: "llm_node",
				Config: LLMNodeConfig{
					Name:         "analyze",
					Model:        "gemini-2.5-flash",
					Instruction:  "Analyze failures",
					ResponseMode: "json",
				},
			},
			{
				ID:   "gate",
				Type: "if_else_node",
				Config: IfElseNodeConfig{
					ConditionLanguage: "CEL",
					Condition:         `state.analyze.status == "pass"`,
				},
			},
			{
				ID:   "retry",
				Type: "llm_node",
				Config: LLMNodeConfig{
					Name:         "retry",
					Model:        "gemini-2.5-flash",
					Instruction:  "Try again",
					ResponseMode: "text",
				},
			},
			{
				ID:   "done",
				Type: "llm_node",
				Config: LLMNodeConfig{
					Name:         "done",
					Model:        "gemini-2.5-flash",
					Instruction:  "Return the final result",
					ResponseMode: "text",
				},
			},
		},
		Edges: []Edge{
			{ID: "e-input", Source: "input-1", SourcePort: "message", Target: "analyze", TargetPort: "message"},
			{ID: "e-analyze-gate", Source: "analyze", SourcePort: "message", Target: "gate", TargetPort: "message"},
			{ID: "e-gate-true", Source: "gate", SourcePort: "message:true", Target: "done", TargetPort: "message"},
			{ID: "e-gate-false", Source: "gate", SourcePort: "message:false", Target: "retry", TargetPort: "message"},
			{ID: "e-retry-loop", Source: "retry", SourcePort: "message", Target: "analyze", TargetPort: "message"},
		},
	}

	if err := g.Validate(); err != nil {
		t.Fatalf("expected loop graph to validate, got %v", err)
	}
}

func TestGraphValidateRejectsMultipleExecutionSuccessors(t *testing.T) {
	g := Graph{
		Version: SupportedGraphVersion,
		Name:    "Forked LLM",
		Nodes: []Node{
			{
				ID:   "input-1",
				Type: "input_node",
				Config: InputNodeConfig{
					Name:        "user_input",
					Description: "User input",
				},
			},
			{
				ID:   "llm-1",
				Type: "llm_node",
				Config: LLMNodeConfig{
					Name:         "router",
					Model:        "gemini-2.5-flash",
					Instruction:  "Route",
					ResponseMode: "text",
				},
			},
			{
				ID:   "llm-2",
				Type: "llm_node",
				Config: LLMNodeConfig{
					Name:         "a",
					Model:        "gemini-2.5-flash",
					Instruction:  "A",
					ResponseMode: "text",
				},
			},
			{
				ID:   "llm-3",
				Type: "llm_node",
				Config: LLMNodeConfig{
					Name:         "b",
					Model:        "gemini-2.5-flash",
					Instruction:  "B",
					ResponseMode: "text",
				},
			},
		},
		Edges: []Edge{
			{ID: "e-input", Source: "input-1", SourcePort: "message", Target: "llm-1", TargetPort: "message"},
			{ID: "e-a", Source: "llm-1", SourcePort: "message", Target: "llm-2", TargetPort: "message"},
			{ID: "e-b", Source: "llm-1", SourcePort: "message", Target: "llm-3", TargetPort: "message"},
		},
	}

	if err := g.Validate(); err == nil {
		t.Fatal("expected multiple-successor validation error, got nil")
	}
}
