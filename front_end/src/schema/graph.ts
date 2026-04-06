import { z } from "zod";

const PositionSchema = z.object({
  x: z.number(),
  y: z.number(),
});

const BaseNodeFields = {
  id: z.string(),
  position: PositionSchema,
  parent_id: z.string().optional(),
  width: z.number().optional(),
  height: z.number().optional(),
};

const InputNodeConfigSchema = z.object({
  name: z.string(),
  description: z.string(),
});

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

const OutputNodeConfigSchema = z.object({
  name: z.string(),
  output_key: z.string(),
  format: z.string(),
});

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
  condition_language: z.literal("CEL"),
  condition: z.string(),
});

const WhileNodeConfigSchema = z.object({
  condition: z.string(),
  max_iterations: z.number().int().positive(),
});

const NodeSchema = z.discriminatedUnion("type", [
  z.object({
    ...BaseNodeFields,
    type: z.literal("input_node"),
    config: InputNodeConfigSchema,
  }),
  z.object({
    ...BaseNodeFields,
    type: z.literal("llm_node"),
    config: LLMNodeConfigSchema,
  }),
  z.object({
    ...BaseNodeFields,
    type: z.literal("output_node"),
    config: OutputNodeConfigSchema,
  }),
  z.object({
    ...BaseNodeFields,
    type: z.literal("toolbox"),
    config: ToolboxNodeConfigSchema,
  }),
  z.object({
    ...BaseNodeFields,
    type: z.literal("if_else_node"),
    config: IfElseNodeConfigSchema,
  }),
  z.object({
    ...BaseNodeFields,
    type: z.literal("while_node"),
    config: WhileNodeConfigSchema,
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
export type ToolboxNodeConfig = z.infer<typeof ToolboxNodeConfigSchema>;
export type IfElseNodeConfig = z.infer<typeof IfElseNodeConfigSchema>;
export type WhileNodeConfig = z.infer<typeof WhileNodeConfigSchema>;
