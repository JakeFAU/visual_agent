# Spec: ADK Expanded Node Types

**Status:** Draft
**Date:** 2026-04-04

## Goal

Expand the Visual Agent JSON contract and IDE capabilities to support the full suite of Google ADK features: Toolboxes (Tools & MCP), Structured Outputs, and Control Flow (If/Else, While).

## Architecture

### 1. Toolbox Node (The Registry)

A standalone node that defines a set of capabilities for an LLM Agent.

- **Type:** `toolbox`
- **Config:**
  - `tools`: Array of built-in ADK tools (e.g., `google_search`).
  - `mcp_servers`: Array of MCP server configurations (command, args, env).
  - `custom_functions`: Array of user-defined function signatures (name, description, parameters).
- **Ports:**
  - Output: `toolbox` (Typed: `toolbox_handle`)

### 2. LLM Agent (Enhanced)

Updated to support Toolboxes and Structured Output.

- **Config:**
  - `response_mode`: `text` | `json`
  - `output_schema`: Optional JSON Schema (Standard Draft 7) for structured output.
- **Ports:**
  - Input: `in_message` (Message)
  - Input: `in_toolbox` (Toolbox Handle)
  - Output: `out_response` (Message/JSON)

### 3. Logic Nodes (Control Flow)

#### If / Else

- **Type:** `if_else_node`
- **Config:**
  - `condition_language`: `CEL` | `JSONPath`
  - `condition`: String (e.g., `$.category == "billing"`)
- **Ports:**
  - Input: `in_data` (Any)
  - Output: `out_true` (Same as input)
  - Output: `out_false` (Same as input)

#### While (Loop)

- **Type:** `while_node`
- **Config:**
  - `condition`: String
  - `max_iterations`: Integer (Safety break)
- **Ports:**
  - Input: `in_start`
  - Output: `out_loop` (Executes the loop body)
  - Output: `out_done` (Final exit)

## Data Flow & Typing

Connection validation will be strictly enforced via `data_type`:

- `message`: Standard LLM text flow.
- `toolbox_handle`: Specific handle for connecting tools to LLMs.
- `json`: Structured data matching a schema.
- `any`: Generic data for logic nodes.

## Next Steps

1. Update `front_end/src/schema/graph.ts` (Zod).
2. Update `back_end/internal/graph/schema.go` (Go Structs).
3. Implement `UnmarshalJSON` logic for new polymorphic configs.
