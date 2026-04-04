# Add New Node Type

I need to add a new node type to the Visual Agent system.

Node Type Name: {{NODE_NAME}}
Node Description: {{NODE_DESCRIPTION}}
Config Fields: {{CONFIG_FIELDS}}

Please execute the following steps:

1. **Front-End (Zod & Types):** Update the Zod schema and TypeScript interfaces in `front_end/src/schema/graph.ts` to include the new node type and its config.
2. **Front-End (React Flow):** Scaffold a new custom React Flow node component in `front_end/src/components/nodes/{{NODE_NAME}}.tsx`. Ensure it has the correct Handle types (ports).
3. **Front-End (Side Panel):** Add the configuration form for this node in the side panel component.
4. **Back-End (Structs):** Update the Go structs in `back_end/internal/graph/schema.go` to handle the JSON unmarshaling for this new `type` and its `config`.
5. **Back-End (Compiler):** Add the compilation case for this node in the `back_end/internal/compiler/build.go` switch statement.
