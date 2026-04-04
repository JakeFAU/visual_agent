# ADK Expanded Node Types Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expand the Visual Agent JSON contract to support Toolboxes, Structured Outputs, and Control Flow nodes.

**Architecture:**
- Update Zod schemas in the front-end to include new node types and enhanced `llm_node` config.
- Update Go structs in the back-end to match the new front-end schemas.
- Update Go `UnmarshalJSON` to handle polymorphic decoding for the new node types.

**Tech Stack:**
- TypeScript, Zod (Front-end)
- Go (Back-end)

---

### Task 1: Front-End Schema Expansion

**Files:**
- Modify: `visual_agent/front_end/src/schema/graph.ts`

- [ ] **Step 1: Update `LLMNodeConfigSchema` to include `response_mode` and `output_schema`.**

```typescript
const LLMNodeConfigSchema = z.object({
  name: z.string(),
  description: z.string(),
  model: z.string(),
  instruction: z.string(),
  response_mode: z.enum(["text", "json"]).default("text"),
  output_schema: z.record(z.any()).optional(), // JSON Schema
  generate_content_config: z.object({
    temperature: z.number().optional(),
    max_output_tokens: z.number().optional(),
  }),
});
```

- [ ] **Step 2: Define schemas for `ToolboxNode`, `IfElseNode`, and `WhileNode`.**

```typescript
const ToolboxNodeConfigSchema = z.object({
  tools: z.array(z.string()), // Built-in tool IDs
  mcp_servers: z.array(z.object({
    name: z.string(),
    command: z.string(),
    args: z.array(z.string()),
    env: z.record(z.string()).optional(),
  })),
  custom_functions: z.array(z.object({
    name: z.string(),
    description: z.string(),
    parameters: z.record(z.any()), // JSON Schema for parameters
  })),
});

const IfElseNodeConfigSchema = z.object({
  condition_language: z.enum(["CEL", "JSONPath"]),
  condition: z.string(),
});

const WhileNodeConfigSchema = z.object({
  condition: z.string(),
  max_iterations: z.number().default(10),
});
```

- [ ] **Step 3: Update `NodeSchema` (discriminatedUnion) with new types.**

```typescript
const NodeSchema = z.discriminatedUnion("type", [
  // ... existing types
  z.object({
    id: z.string(),
    type: z.literal("toolbox"),
    position: PositionSchema,
    config: ToolboxNodeConfigSchema,
  }),
  z.object({
    id: z.string(),
    type: z.literal("if_else_node"),
    position: PositionSchema,
    config: IfElseNodeConfigSchema,
  }),
  z.object({
    id: z.string(),
    type: z.literal("while_node"),
    position: PositionSchema,
    config: WhileNodeConfigSchema,
  }),
]);
```

- [ ] **Step 4: Update exported types.**

```typescript
export type ToolboxNodeConfig = z.infer<typeof ToolboxNodeConfigSchema>;
export type IfElseNodeConfig = z.infer<typeof IfElseNodeConfigSchema>;
export type WhileNodeConfig = z.infer<typeof WhileNodeConfigSchema>;
```

---

### Task 2: Back-End Schema Expansion

**Files:**
- Modify: `visual_agent/back_end/internal/graph/schema.go`

- [ ] **Step 1: Update `LLMNodeConfig` struct.**

```go
type LLMNodeConfig struct {
	Name                  string                 `json:"name"`
	Description           string                 `json:"description"`
	Model                 string                 `json:"model"`
	Instruction           string                 `json:"instruction"`
	ResponseMode          string                 `json:"response_mode"` // "text" or "json"
	OutputSchema          map[string]interface{} `json:"output_schema,omitempty"`
	GenerateContentConfig GenerateContentConfig `json:"generate_content_config"`
}
```

- [ ] **Step 2: Add new config structs.**

```go
type MCPServerConfig struct {
	Name    string            `json:"name"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

type CustomFunctionConfig struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type ToolboxNodeConfig struct {
	Tools           []string               `json:"tools"`
	MCPServers      []MCPServerConfig      `json:"mcp_servers"`
	CustomFunctions []CustomFunctionConfig `json:"custom_functions"`
}

type IfElseNodeConfig struct {
	ConditionLanguage string `json:"condition_language"` // "CEL" or "JSONPath"
	Condition         string `json:"condition"`
}

type WhileNodeConfig struct {
	Condition     string `json:"condition"`
	MaxIterations int    `json:"max_iterations"`
}
```

- [ ] **Step 3: Update `UnmarshalJSON` for `Node` to handle new types.**

```go
func (n *Node) UnmarshalJSON(data []byte) error {
	// ... existing logic to unmarshal aux
	switch n.Type {
	case "input_node":
		// ...
	case "llm_node":
		var config LLMNodeConfig
		if err := json.Unmarshal(aux.Config, &config); err != nil {
			return err
		}
		n.Config = config
	case "output_node":
		// ...
	case "toolbox":
		var config ToolboxNodeConfig
		if err := json.Unmarshal(aux.Config, &config); err != nil {
			return err
		}
		n.Config = config
	case "if_else_node":
		var config IfElseNodeConfig
		if err := json.Unmarshal(aux.Config, &config); err != nil {
			return err
		}
		n.Config = config
	case "while_node":
		var config WhileNodeConfig
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

---

### Task 3: Verification

- [ ] **Step 1: Create a comprehensive test JSON containing all new node types.**
- [ ] **Step 2: Verify front-end Zod validation.**
- [ ] **Step 3: Verify back-end Go unmarshaling with a unit test.**

```go
func TestUnmarshalComplexGraph(t *testing.T) {
    complexJSON := `{
        "version": "1.0",
        "name": "complex_workflow",
        "nodes": [
            {
                "id": "toolbox-1",
                "type": "toolbox",
                "position": {"x": 0, "y": 0},
                "config": {
                    "tools": ["google_search"],
                    "mcp_servers": [{"name": "fs", "command": "npx", "args": ["mcp-fs"]}],
                    "custom_functions": []
                }
            },
            {
                "id": "llm-1",
                "type": "llm_node",
                "position": {"x": 200, "y": 0},
                "config": {
                    "name": "agent",
                    "description": "desc",
                    "model": "gemini-2.5-flash",
                    "instruction": "inst",
                    "response_mode": "json",
                    "output_schema": {"type": "object"},
                    "generate_content_config": {}
                }
            },
            {
                "id": "if-1",
                "type": "if_else_node",
                "position": {"x": 400, "y": 0},
                "config": {
                    "condition_language": "CEL",
                    "condition": "true"
                }
            }
        ],
        "edges": []
    }`
    var graph Graph
    if err := json.Unmarshal([]byte(complexJSON), &graph); err != nil {
        t.Fatalf("Failed to unmarshal: %v", err)
    }
}
```
