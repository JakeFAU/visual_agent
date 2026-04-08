# Documents

Public-facing supporting docs, screenshots, and example graphs for Visual Agent live in this directory.

## Screenshots

- `images/sample_agent.png`: current canvas overview image used in the root README

## Example Graphs

- `examples/simple-chat.json`: minimal single-agent workflow
- `examples/structured-router.json`: structured classification plus `if_else_node` branching
- `examples/while-loop-review.json`: looped review/rewrite workflow using a `while_node`
- `examples/neo-sigma-control-loop.json`: larger experimental control-loop graph

The frontend ships a mirrored subset of these examples through `front_end/public/examples/` so they can be loaded directly from the in-app example library.
