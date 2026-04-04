# Visual Agent - Project Overview

## System Architecture

Visual Agent is a node-based IDE for building, compiling, and deploying generative AI agents.

- **Front-end:** React, Vite, React Flow, Zustand. Acts as the visual editor.
- **Back-end:** Go. Acts as the compiler and deployment orchestrator.
- **Contract:** A strictly validated, bidirectional JSON schema. The front-end produces it; the back-end consumes it to compile Google ADK Agents runnable on Vertex AI. The front-end must also be able to ingest this JSON to perfectly rebuild the visual canvas.

## v0 Architecture & Constraints

- **Infrastructure:** Google Cloud Platform, utilizing Application Default Credentials (ADC).
- **AI Models:** Gemini via Vertex AI.
- **Extensibility:** While v0 is tightly coupled to Google services, the system design (especially the Go interfaces and front-end node configurations) must remain modular to support alternative LLMs and runtimes in the future.

## Core Data Structures

The primary artifact is the Graph JSON. It must contain:

1. `version` & `name`
2. `nodes`: Array containing `id`, `type`, `position` (x,y), and a polymorphic `config` object dependent on the node `type` (`input_node`, `llm_node`, `output_node`, `toolbox`).
3. `edges`: Array containing `id`, `source`, `source_port`, `target`, `target_port`, `data_type`, and `edge_kind`.

*Crucial Rule:* Edges represent typed data flows. Connections are strictly validated based on `data_type` (represented visually by port colors).
