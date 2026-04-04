# Back-End Context: Visual Agent Compiler

## Tech Stack

- **Language:** Go (1.25+)
- **AI Framework:** Google GenAI SDK (ADK) / Vertex AI.
- **Authentication:** Google Application Default Credentials (ADC).

## Architecture Guidelines

1. **JSON Unmarshaling & Validation:** The Go backend must define strong structs matching the front-end JSON schema. Use polymorphic unmarshaling for the `nodes` array based on the `type` field to properly decode the `config` objects.
2. **Graph Compilation (The Walker):** - Translate the parsed nodes and edges into an execution Directed Acyclic Graph (DAG) or a state machine.
   - Map the `llm_node` config to the corresponding ADK Model setup (e.g., configuring `SystemInstruction`, `GenerationConfig.Temperature`).
   - Map `toolbox` nodes to ADK Tool declarations (e.g., `Google Search` built-in tool).
3. **Execution Engine:** v0 runs the compiled agent against Vertex AI using the `gemini-2.5-flash` model. The engine handles the standard input/output loop, injecting the user message, handling tool call loops, and returning the final grounded response metadata.
4. **Extensibility:** Wrap the Vertex AI specific calls in an interface (e.g., `AgentRuntime`). Do not leak Vertex-specific types into the core graph compilation logic.
