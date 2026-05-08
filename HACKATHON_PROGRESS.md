# Pulse Hackathon: Progress Summary

This document summarizes all the features, integrations, and architectural components that have been successfully built for the **Pulse (Project Dike)** hackathon project.

We have implemented a **Multi-Model AI Orchestration Framework** combining OpenClaw-style session-based agents with a Railway-Oriented Pipeline for secure command execution.

---

## 🧠 1. Python Layer: Multi-Model Orchestrator (OpenClaw-Style)

A custom multi-agent orchestration layer built without LangChain/CrewAI, using pure ReAct loops and native OpenAI function calling.

### Agents & Models

| Agent | Model | Cost | Responsibility |
|-------|-------|------|----------------|
| **Hunter** | GPT-4o-mini | Free tier | Discovers target repositories |
| **Cloner** | GPT-4o-mini | Free tier | Analyzes Dockerfile/Makefile for build commands |
| **Gatekeeper** | GPT-4o-mini + Tools | Free tier | Evaluates safety using Pulse sandbox tool |
| **Advisor** | LLaMA-3 (Ollama) | 100% Free | Deep technical/business risk explanations |
| **Coder** | CodeLLaMA (Ollama) | 100% Free | Infrastructure code analysis |
| **Reporter** | GPT-4o-mini | Free tier | Executive summaries with action items |

### Key Patterns
- **ReAct Loops:** `Reason → Act (tool call) → Observe → Repeat`
- **Session-Based Routing:** Each agent has isolated conversation history
- **Model Specialization:** Different models for different cognitive tasks
- **Hybrid Cloud/Local:** GPT-4o-mini for speed, Ollama for deep analysis (zero API cost)

---

## 🖥️ 2. Go Layer: Railway-Oriented Pipeline (ROP)

A secure execution engine using immutable contexts and pure function agents.

### Architecture: Immutable State

**Before (Vulnerable):**
```go
ctx.IsSafe = true  // Any agent could mutate
ctx.Decision = "ALLOW"  // Unclear who set this
```

**After (Secure):**
```go
ctx = ctx.WithDecision(pipeline.DecisionAllow, "gatekeeper evaluated")
// ctx.IsSafe() is deterministic - no mutations possible
```

### Components

- **`pkg/pipeline`** (NEW): Railway-Oriented Pipeline core
  - `Context`: Immutable state container with `.With*()` methods
  - `Agent`: Pure functions `func(Context) (Context, error)`
  - `Execute()`: Railway-oriented error handling (success flows, errors branch)
  - `AgentStep`: Complete audit trail of every transformation

- **`pkg/agents`** (NEW): Refactored pure function agents
  - `Gatekeeper()`: Policy evaluation → DENY/PREVIEW/ALLOW
  - `Ghost()`: Docker sandbox execution (only if PREVIEW)
  - `Auditor()`: Immutable SQLite logging
  - `Advisor()`: LLM explanations

- **`pkg/gatekeeper`**: Shell AST parsing + policy evaluation
- **`pkg/ghost`**: Docker Alpine sandbox with filesystem sync
- **`pkg/diff`**: Host/container filesystem comparison
- **`pkg/audit`**: Append-only SQLite logging
- **`pkg/web`**: HTTP API for dashboard

### Security Guarantees
1. **No Hidden Mutations:** Context is immutable, every change creates new object
2. **Audit Trail:** Every agent step recorded with input/output snapshots
3. **Deterministic Safety:** `ctx.IsSafe()` purely derived from `ctx.Decision`
4. **Testable:** Agents are pure functions - test without Docker/SQLite

---

## 🎨 2. Frontend: The SOC2 Dashboard & Presentation

The frontend encompasses the user interface, the visual audit dashboard, and the emotional hackathon narrative.

- **Command Audit Timeline (`static/index.html`)**: A beautiful, single-page SOC2 compliance dashboard.
  - Features a clean, modern dark-mode UI with system fonts.
  - Automatically polls the backend API every 3 seconds to update the live "Command Audit Timeline".
  - Displays color-coded risk badges (`CRITICAL`, `HIGH`) and dynamic aggregate statistics (Total Commands, Applied vs Rejected ratios).
- **The Narrative ("Cutting the Root")**: The pitch explicitly focuses on local Indian businesses (Kirana tech, hospitals) being wiped out by simple typos or disgruntled freelancers. It positions Pulse as a preventative "pre-flight check" rather than a reactionary backup tool.
- **The Demo Script (`demo_script.md`)**: A precisely timed 2-minute pitch script designed to hook the judges with real-world news (e.g., KiranaPro, Microsoft Bangalore outage) and transition smoothly into a dual-screen live demo of the Backend CLI and Frontend Dashboard.

---

## Current Status: 100% Ready for Demo

The project is fully executable end-to-end with **multi-model orchestration**.

**Quick Start:**
```bash
# 1. Set API key
export OPENAI_API_KEY='your-key'

# 2. Install Python deps
pip install -r requirements.txt

# 3. Run multi-model orchestrator (6 agents, 3 models)
python main.py
```

**Alternative Modes:**
```bash
# Interactive REPL (Go backend only)
cd backend && go run cmd/pulse/main.go

# Web Dashboard
cd backend && go run cmd/pulse/main.go --web
# Dashboard at http://localhost:8080
```

**With Local Models (100% Free):**
```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh
ollama pull llama3 codellama

# Run (will use local models for Advisor + Coder)
python main.py
```
