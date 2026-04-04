# Visual Agent

Visual Agent is a node-based IDE for building, compiling, and deploying generative AI agents. It provides a visual canvas to design complex multi-agent workflows, which are then compiled into executable Google ADK (Agent Development Kit) agents runnable on Vertex AI.

![Visual Agent Canvas](https://via.placeholder.com/800x450?text=React+Flow+Canvas+Screenshot)

## Architecture

Visual Agent follows a "vibe coding with tight contracts" philosophy:

1.  **Front-End:** A React application using React Flow for the visual canvas and Zustand for state management.
2.  **Strict JSON Contract:** Bidirectional schema validation (Zod in TS, Go Structs in Go) ensures the front-end and back-end are always in sync.
3.  **Back-End:** A Go-based compiler that translates the Graph JSON into optimized Google ADK agents.
4.  **Execution:** Agents are executed via the ADK, targeting Google Vertex AI or local in-memory runtimes.

## Prerequisites

- **Node.js:** v18+ (for the front-end)
- **Go:** v1.21+ (for the back-end)
- **Google Cloud:** A project with Vertex AI API enabled and [Application Default Credentials (ADC)](https://cloud.google.com/docs/authentication/provide-credentials-adc) configured locally.

## Quickstart

### 1. Start the Back-End API
```bash
cd back_end
go run cmd/visual-agent/main.go serve
```
The API will be available at `http://localhost:8080`.

### 2. Start the Front-End IDE
```bash
cd front_end
npm install
npm run dev
```
Open `http://localhost:5173` in your browser.

## License & Community

This project is licensed under the **MIT License**.

We love our community! If you find this project useful, please consider giving us a ⭐ on GitHub!

---
Copyright © 2026 Jacob D. Bourne
