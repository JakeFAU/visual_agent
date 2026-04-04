# Contributing to Visual Agent

Thank you for your interest in contributing to Visual Agent! We welcome community contributions and aim to make the process as smooth as possible.

## Local Development Setup

### Front-End
1.  Navigate to `front_end/`.
2.  Run `npm install`.
3.  Run `npm run dev`.

### Back-End
1.  Navigate to `back_end/`.
2.  Ensure you have Go 1.21+ installed.
3.  Run `go mod tidy`.
4.  Run `go run cmd/visual-agent/main.go serve`.

## The JSON Contract (CRITICAL)

Visual Agent is built on a "vibe coding with tight contracts" philosophy. The core of the system is the bidirectional JSON schema that defines how workflows are serialized and executed.

**Any changes to node types, configurations, or edge properties MUST be reflected in BOTH of the following files simultaneously:**

1.  `front_end/src/schema/graph.ts` (Zod Schema)
2.  `back_end/internal/graph/schema.go` (Go Structs)

Pull requests that modify the contract in only one place or introduce schema mismatches will not be merged. Automated CI checks are in place to enforce this.

## Pull Request Process

1.  Create a new branch for your feature or bug fix.
2.  Ensure all front-end Zod validations pass.
3.  Ensure all back-end Go tests pass (`go test ./...`).
4.  Update the documentation if necessary.
5.  Open a Pull Request with a clear description of the changes.

---
Copyright © 2026 Jacob D. Bourne
