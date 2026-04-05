# Contributing to Visual Agent

Thank you for your interest in contributing to Visual Agent! We welcome community contributions and aim to make the process as smooth as possible.

## Prerequisites

- Node.js `^20.19.0` or `>=22.12.0`
- Go `1.25+`
- For model-backed execution, either `GOOGLE_API_KEY` or Vertex AI credentials

## Local Development Setup

### Front-End
1.  Navigate to `front_end/`.
2.  Run `npm install`.
3.  Run `npm run dev`.
4.  Run `npm run lint`.
5.  Run `npm run typecheck`.
6.  Run `npm run build`.

### Back-End
1.  Navigate to `back_end/`.
2.  Ensure you have Go 1.25+ installed.
3.  Run `go mod download`.
4.  Run `go test ./...`.
5.  Run `go vet ./...`.
6.  Run `golangci-lint run ./...` if you have it installed locally.
7.  Run `go run cmd/visual-agent/main.go serve`.

## The JSON Contract (CRITICAL)

The core of the system is the shared JSON schema that defines how workflows are serialized and executed.

**Any changes to node types, configurations, or edge properties MUST be reflected in BOTH of the following files simultaneously:**

1.  `front_end/src/schema/graph.ts` (Zod Schema)
2.  `back_end/internal/graph/schema.go` (Go Structs)

There is not currently a generated schema-diff check in CI, so contract changes should update both files in the same PR and be exercised through the UI and back-end validation paths.

## Pull Request Process

1.  Create a new branch for your feature or bug fix.
2.  Run the front-end checks: `npm run lint`, `npm run typecheck`, and `npm run build`.
3.  Run the back-end checks: `go test ./...` and `go vet ./...`.
4.  Update `README.md`, `documents/`, or screenshots when the user-facing workflow changes.
5.  Open a Pull Request with a clear description of the changes.

---
Copyright © 2026 Jacob D. Bourne
