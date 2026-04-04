# Sync and Validate Schema

Analyze the TypeScript Zod schema in `front_end/src/schema/graph.ts` and the Go structs in `back_end/internal/graph/schema.go`.

1. Identify any discrepancies in field names, data types, or required/optional flags between the front-end definitions and the back-end definitions.
2. Provide the code to fix any mismatches, treating the TypeScript Zod schema as the primary source of truth.
3. Ensure the Go custom unmarshaler `UnmarshalJSON` for the Node struct correctly handles all current polymorphic `config` types based on the `type` string.
