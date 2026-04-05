# Visual Agent

Visual Agent is a node-based IDE for building, validating, and executing generative AI workflows. It provides a visual canvas for designing multi-agent systems, then compiles those graphs into Google ADK (Agent Development Kit) agents that can run locally or on Vertex AI.

![Visual Agent canvas overview](documents/images/sample_agent.png)

## Architecture

Visual Agent follows a "vibe coding with tight contracts" philosophy:

1.  **Front-End:** A React application using React Flow for the visual canvas and Zustand for state management.
2.  **Shared JSON Contract:** Zod in TypeScript and matching Go structs define how workflows are serialized and executed.
3.  **Back-End:** A Go-based compiler that translates the Graph JSON into optimized Google ADK agents.
4.  **Execution:** Agents are executed via the ADK, targeting a local in-memory runtime by default and Vertex AI when configured.

## Prerequisites

- **Node.js:** `^20.19.0` or `>=22.12.0` (for the front-end)
- **Go:** `1.25+` (for the back-end)
- **Model-backed execution:** either a `GOOGLE_API_KEY` for the Gemini Developer API, or Vertex AI access with [Application Default Credentials (ADC)](https://cloud.google.com/docs/authentication/provide-credentials-adc), `GOOGLE_CLOUD_PROJECT`, and an optional `GOOGLE_CLOUD_LOCATION`.

## Quickstart

### 1. Start the Back-End API
```bash
cd back_end
go run cmd/visual-agent/main.go serve
```
The API listens on `http://127.0.0.1:8080` by default. Override it with `VISUAL_AGENT_SERVER_ADDR` if needed.

### 2. Start the Front-End IDE
```bash
cd front_end
npm install
npm run dev
```
Open `http://localhost:5173` in your browser.

### 3. Configure model access

For the Gemini Developer API:
```bash
export GOOGLE_API_KEY=your_api_key
```

For Vertex AI:
```bash
export VISUAL_AGENT_RUNTIME_TYPE=vertex
export GOOGLE_CLOUD_PROJECT=your-project-id
export GOOGLE_CLOUD_LOCATION=us-central1
```

## Development

- Front-end checks: `cd front_end && npm run lint && npm run typecheck && npm run build`
- Back-end checks: `cd back_end && go test ./... && go vet ./... && golangci-lint run ./...`
- Public screenshots and longer-form project docs live in [`documents/`](documents/README.md).

## License & Community

This project is licensed under the **MIT License**.

We love our community! If you find this project useful, please consider giving us a ⭐ on GitHub!

---
Copyright © 2026 Jacob D. Bourne
