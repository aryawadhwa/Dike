# Pulse Hackathon: Progress Summary

This document summarizes all the features, integrations, and architectural components that have been successfully built for the **Pulse (Project Dike)** hackathon project up to this point. 

We have completely rebranded and implemented a **Multi-Agent DevSecOps Framework** designed to protect local Indian MSMEs from devastating human error on the command line.

---

## 🖥️ 1. Backend: The Multi-Agent Infrastructure

The backend is written in Go and orchestrates a decoupled team of AI agents that intercept, simulate, and audit commands before they touch the host OS.

- **Interceptor Agent (`pkg/repl`)**: A custom Go-based interactive REPL that intercepts user shell commands before execution.
- **Gatekeeper Agent (`pkg/gatekeeper`)**: Evaluates intercepted commands against YAML risk policies (`configs/SAFETY.yaml`). It autonomously classifies commands as `ALLOW`, `PREVIEW`, or `DENY`.
- **Ghost Sandbox Agent (`pkg/ghost`)**: For risky commands (e.g., `rm -rf`), this agent spins up an isolated Docker Alpine container. It dynamically clones the host directory into the container to safely simulate the command's "blast radius."
- **Diff Agent (`pkg/diff`)**: Scans and compares the host filesystem with the sandbox filesystem post-execution, generating a unified summary of exactly which files would be destroyed or altered.
- **Advisor Agent (`pkg/advisor`)**: An LLM-powered safety advisor. When a user requests an explanation (`e`), it invokes a local offline model (LLaMA-3) to instantly generate a 3-bullet explanation of the command's technical risk and potential business impact. 
- **Auditor Agent (`pkg/audit`)**: An immutable SQLite logging engine that records every intercepted command, the Gatekeeper's risk assessment, the final user decision, and the Diff Agent's preview summary.
- **HTTP API (`pkg/web`)**: A lightweight Go `net/http` web server that exposes `/api/audit` (raw JSON logs) and `/api/stats` (real-time metrics) for the frontend to consume.

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

The project is fully executable end-to-end. 

**To run the complete system (Backend & Frontend):**
```bash
# Navigate to backend
cd backend

# Launch the Pulse Backend CLI and the Frontend Web Dashboard simultaneously
go run cmd/pulse/main.go --web
```
*The web dashboard is available at `http://localhost:8080`.*
