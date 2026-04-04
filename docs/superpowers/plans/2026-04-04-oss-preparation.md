# OSS Preparation & CI Infrastructure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prepare the `visual_agent` repository for open-source release with comprehensive documentation, contributor guidelines, and automated CI checks.

**Architecture:**
- Root-level documentation (`README.md`, `CONTRIBUTING.md`, `LICENSE`).
- GitHub-specific configurations (`.github/` templates and workflows).
- Dual-stack CI pipeline (Go + Node.js).

**Tech Stack:**
- Markdown, GitHub Actions (YAML), MIT License.

---

### Task 1: Root README.md

**Files:**
- Create: `visual_agent/README.md`

- [ ] **Step 1: Write the comprehensive README.md.**

---

### Task 2: Contributing Guide

**Files:**
- Create: `visual_agent/CONTRIBUTING.md`

- [ ] **Step 1: Write the CONTRIBUTING.md with the "JSON Contract" mandate.**

---

### Task 3: GitHub Templates

**Files:**
- Create: `visual_agent/.github/ISSUE_TEMPLATE/bug_report.md`
- Create: `visual_agent/.github/ISSUE_TEMPLATE/feature_request.md`
- Create: `visual_agent/.github/PULL_REQUEST_TEMPLATE.md`

- [ ] **Step 1: Create Bug Report template.**
- [ ] **Step 2: Create Feature Request template.**
- [ ] **Step 3: Create PR template with contract sync checklist.**

---

### Task 4: GitHub Actions CI Pipeline

**Files:**
- Create: `visual_agent/.github/workflows/ci.yml`

- [ ] **Step 1: Implement the CI workflow with `backend-checks` and `frontend-checks` jobs.**

---

### Task 5: Security & Ignored Files Audit

**Files:**
- Modify: `visual_agent/.gitignore`

- [ ] **Step 1: Audit and update .gitignore to ensure no credentials or build artifacts are committed.**

---

### Task 6: License

**Files:**
- Create: `visual_agent/LICENSE`

- [ ] **Step 1: Create the MIT License file.**
