# JSON Contract Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish a strict bidirectional JSON contract between the React front-end and the Go back-end for the Visual Agent graph.

**Architecture:** 
- Front-end: Use Zod with `z.discriminatedUnion` to define a polymorphic Node schema where the `config` object is strictly typed based on the node's `type`.
- Back-end: Use Go structs with custom `UnmarshalJSON` using `json.RawMessage` to decode the polymorphic `Config` field based on the `Type` string.

**Tech Stack:** 
- TypeScript, Zod (Front-end)
- Go (Back-end)

---

### Task 1: Front-End Zod Schema

**Files:**
- Create: `visual_agent/front_end/src/schema/graph.ts`

- [ ] **Step 1: Define the Zod schema for Graph, Node, and Edge.**

```typescript
import { z } from "zod";

const PositionSchema = z.object({
  x: z.number(),
  y: z.number(),
});

const InputNodeConfigSchema = z.object({
  name: z.string(),
  description: z.string(),
});

const LLMNodeConfigSchema = z.object({
  name: z.string(),
  description: z.string(),
  model: z.string(),
  instruction: z.string(),
  generate_content_config: z.object({
    temperature: z.number().optional(),
    max_output_tokens: z.number().optional(),
  }),
});

const OutputNodeConfigSchema = z.object({
  name: z.string(),
  output_key: z.string(),
  format: z.string(),
});

const NodeSchema = z.discriminatedUnion("type", [
  z.object({
    id: z.string(),
    type: z.literal("input_node"),
    position: PositionSchema,
    config: InputNodeConfigSchema,
  }),
  z.object({
    id: z.string(),
    type: z.literal("llm_node"),
    position: PositionSchema,
    config: LLMNodeConfigSchema,
  }),
  z.object({
    id: z.string(),
    type: z.literal("output_node"),
    position: PositionSchema,
    config: OutputNodeConfigSchema,
  }),
]);

const EdgeSchema = z.object({
  id: z.string(),
  source: z.string(),
  source_port: z.string(),
  target: z.string(),
  target_port: z.string(),
  data_type: z.string(),
  edge_kind: z.string(),
});

export const GraphSchema = z.object({
  version: z.string(),
  name: z.string(),
  nodes: z.array(NodeSchema),
  edges: z.array(EdgeSchema),
});

export type Graph = z.infer<typeof GraphSchema>;
export type Node = z.infer<typeof NodeSchema>;
export type Edge = z.infer<typeof EdgeSchema>;
export type Position = z.infer<typeof PositionSchema>;
export type InputNodeConfig = z.infer<typeof InputNodeConfigSchema>;
export type LLMNodeConfig = z.infer<typeof LLMNodeConfigSchema>;
export type OutputNodeConfig = z.infer<typeof OutputNodeConfigSchema>;
```

- [ ] **Step 2: Verify the schema with the POC JSON.**

Create a temporary test file `visual_agent/front_end/src/schema/test_schema.ts` to validate the POC JSON against the schema.

```typescript
import { GraphSchema } from "./graph";

const pocJson = {
  "version": "1.0",
  "name": "test_workflow",
  "nodes": [
    {
      "id": "input_node-1",
      "type": "input_node",
      "position": { "x": -247, "y": -317 },
      "config": {
        "name": "user_input",
        "description": "The user input"
      }
    },
    {
      "id": "llm_node-2",
      "type": "llm_node",
      "position": { "x": 137.5, "y": -289.5 },
      "config": {
        "name": "llm_agent",
        "description": "The actual Agent",
        "model": "gemini-2.5-flash",
        "instruction": "You are a research assistant with access to Google Search.\\n  When asked a question, search for current information and summarize what you find.",
        "generate_content_config": {
          "temperature": 0.7,
          "max_output_tokens": 4096
        }
      }
    },
    {
      "id": "output_node-4",
      "type": "output_node",
      "position": { "x": 618.5, "y": -288.5 },
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
      "source": "input_node-1",
      "source_port": "out_message",
      "target": "llm_node-2",
      "target_port": "in_message",
      "data_type": "message",
      "edge_kind": "data_flow"
    },
    {
      "id": "edge-4",
      "source": "llm_node-2",
      "source_port": "out_message",
      "target": "output_node-4",
      "target_port": "in_result",
      "data_type": "message",
      "edge_kind": "data_flow"
    }
  ]
};

try {
  GraphSchema.parse(pocJson);
  console.log("POC JSON validated successfully!");
} catch (error) {
  console.error("Validation failed:", error);
}
```

Run: `npx ts-node visual_agent/front_end/src/schema/test_schema.ts` (Note: Ensure dependencies are installed or use a simple node script if ts-node is not available).

---

### Task 2: Back-End Go Structs

**Files:**
- Create: `visual_agent/back_end/internal/graph/schema.go`

- [ ] **Step 1: Define the Go structs for Graph, Node, and Edge.**

```go
package graph

import (
	"encoding/json"
	"fmt"
)

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type InputNodeConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type LLMNodeConfig struct {
	Name                  string                 `json:"name"`
	Description           string                 `json:"description"`
	Model                 string                 `json:"model"`
	Instruction           string                 `json:"instruction"`
	GenerateContentConfig GenerateContentConfig `json:"generate_content_config"`
}

type GenerateContentConfig struct {
	Temperature      float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"max_output_tokens,omitempty"`
}

type OutputNodeConfig struct {
	Name      string `json:"name"`
	OutputKey string `json:"output_key"`
	Format    string `json:"format"`
}

type Node struct {
	ID       string      `json:"id"`
	Type     string      `json:"type"`
	Position Position    `json:"position"`
	Config   interface{} `json:"config"`
}

type Edge struct {
	ID         string `json:"id"`
	Source     string `json:"source"`
	SourcePort string `json:"source_port"`
	Target     string `json:"target"`
	TargetPort string `json:"target_port"`
	DataType   string `json:"data_type"`
	EdgeKind   string `json:"edge_kind"`
}

type Graph struct {
	Version string `json:"version"`
	Name    string `json:"name"`
	Nodes   []Node `json:"nodes"`
	Edges   []Edge `json:"edges"`
}
```

- [ ] **Step 2: Implement custom `UnmarshalJSON` for polymorphic `Node`.**

```go
func (n *Node) UnmarshalJSON(data []byte) error {
	type Alias Node
	aux := &struct {
		Config json.RawMessage `json:"config"`
		*Alias
	}{
		Alias: (*Alias)(n),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch n.Type {
	case "input_node":
		var config InputNodeConfig
		if err := json.Unmarshal(aux.Config, &config); err != nil {
			return err
		}
		n.Config = config
	case "llm_node":
		var config LLMNodeConfig
		if err := json.Unmarshal(aux.Config, &config); err != nil {
			return err
		}
		n.Config = config
	case "output_node":
		var config OutputNodeConfig
		if err := json.Unmarshal(aux.Config, &config); err != nil {
			return err
		}
		n.Config = config
	default:
		return fmt.Errorf("unknown node type: %s", n.Type)
	}

	return nil
}
```

- [ ] **Step 3: Verify the Go structs with the POC JSON.**

Create a test file `visual_agent/back_end/internal/graph/schema_test.go` to validate the POC JSON.

```go
package graph

import (
	"encoding/json"
	"testing"
)

func TestUnmarshalGraph(t *testing.T) {
	pocJSON := `{
  "version": "1.0",
  "name": "test_workflow",
  "nodes": [
    {
      "id": "input_node-1",
      "type": "input_node",
      "position": { "x": -247, "y": -317 },
      "config": {
        "name": "user_input",
        "description": "The user input"
      }
    },
    {
      "id": "llm_node-2",
      "type": "llm_node",
      "position": { "x": 137.5, "y": -289.5 },
      "config": {
        "name": "llm_agent",
        "description": "The actual Agent",
        "model": "gemini-2.5-flash",
        "instruction": "You are a research assistant with access to Google Search.\\n  When asked a question, search for current information and summarize what you find.",
        "generate_content_config": {
          "temperature": 0.7,
          "max_output_tokens": 4096
        }
      }
    },
    {
      "id": "output_node-4",
      "type": "output_node",
      "position": { "x": 618.5, "y": -288.5 },
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
      "source": "input_node-1",
      "source_port": "out_message",
      "target": "llm_node-2",
      "target_port": "in_message",
      "data_type": "message",
      "edge_kind": "data_flow"
    },
    {
      "id": "edge-4",
      "source": "llm_node-2",
      "source_port": "out_message",
      "target": "output_node-4",
      "target_port": "in_result",
      "data_type": "message",
      "edge_kind": "data_flow"
    }
  ]
}`

	var graph Graph
	if err := json.Unmarshal([]byte(pocJSON), &graph); err != nil {
		t.Fatalf("Failed to unmarshal POC JSON: %v", err)
	}

	if len(graph.Nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(graph.Nodes))
	}

	for _, node := range graph.Nodes {
		switch node.Type {
		case "input_node":
			if _, ok := node.Config.(InputNodeConfig); !ok {
				t.Errorf("Expected InputNodeConfig for input_node, got %T", node.Config)
			}
		case "llm_node":
			if _, ok := node.Config.(LLMNodeConfig); !ok {
				t.Errorf("Expected LLMNodeConfig for llm_node, got %T", node.Config)
			}
		case "output_node":
			if _, ok := node.Config.(OutputNodeConfig); !ok {
				t.Errorf("Expected OutputNodeConfig for output_node, got %T", node.Config)
			}
		}
	}
}
```

Run: `go test -v ./visual_agent/back_end/internal/graph/...`
