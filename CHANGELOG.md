# Changelog

## v0.1.0 - 2026-04-08

First public release of Visual Agent.

### Highlights

- Local-first visual IDE for building Google ADK workflows with React Flow on the frontend and Go on the backend
- Shared graph contract across TypeScript and Go for validation, persistence, and execution
- Control-flow primitives including `if_else_node` and `while_node`
- LLM nodes with text and structured JSON output modes
- Toolbox nodes with built-in tools and MCP server wiring
- Execution diagnostics including steps, token usage, node visits, and exit reasons
- In-app graph library with saved workflows, curated examples, and JSON import

### Included Example Workflows

- `Simple Chat`: minimal input -> LLM -> output path
- `Structured Router`: structured classification followed by explicit branching
- `While Loop Review`: looped review/rewrite flow with a local iteration cap

### Scope Notes

- Visual Agent is currently optimized for trusted local use, not hosted multi-tenant deployment.
- The built-in tool catalog is intentionally narrow in `v0.1.0`.
- Custom Functions remain a planned feature and are not executable in this release.

### Upgrade Notes

- Run-wide execution budgets are now exposed directly in the app shell.
- Workflow naming, saved-graph loading, and example loading are now handled in-app rather than through browser prompt/alert flows.
- JSON-mode state used by control-flow nodes is normalized before CEL evaluation, which fixes structured-router style workflows.
