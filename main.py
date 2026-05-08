"""
Multi-Model ReAct Orchestrator for RepoGuard
OpenClaw-style session-based multi-agent with hybrid models:
- GPT-4o mini (cloud, essentially free) for general tasks
- Local Ollama models (completely free) for specialized analysis
"""

import json
import os
import requests
import subprocess
from dataclasses import dataclass, field
from typing import List, Dict, Optional
from openai import OpenAI
from dotenv import load_dotenv

load_dotenv()

# Model routing configuration
MODEL_ROUTER = {
    "hunter": {"type": "openai", "model": "gpt-4o-mini"},
    "cloner": {"type": "openai", "model": "gpt-4o-mini"},
    "gatekeeper": {"type": "openai", "model": "gpt-4o-mini"},
    "advisor_deep": {"type": "ollama", "model": "llama3"},
    "coder": {"type": "ollama", "model": "codellama"},
    "reporter": {"type": "openai", "model": "gpt-4o-mini"},
}

SYSTEM_PROMPTS = {
    "hunter": "You are a Repository Scout. Find and identify relevant GitHub repositories for security analysis. Be concise and return only valid repository URLs.",
    "cloner": "You are an Environment Prep Engineer. Analyze repository structure (Dockerfile, Makefile, package.json) and identify build/clean commands. Return only the command string.",
    "gatekeeper": "You are a Build Safety Engineer. You NEVER run a command without sandboxing it first. Use the pulse_sandbox tool for every command execution.",
    "advisor_deep": "You are a Senior Security Advisor. Provide detailed technical explanations of blast radius, business impact, and self-healing recommendations.",
    "coder": "You are a Code Infrastructure Analyst. Analyze Dockerfiles, shell scripts, and build configurations for hidden risks.",
    "reporter": "You are a Security Auditor. Compile findings into actionable, executive-friendly reports emphasizing isolation features.",
}

PULSE_TOOL = {
    "type": "function",
    "function": {
        "name": "pulse_sandbox",
        "description": "Execute a potentially risky shell command inside the Pulse Docker Sandbox. Returns risk level and filesystem diff.",
        "parameters": {
            "type": "object",
            "properties": {
                "command": {"type": "string", "description": "The shell command to execute"},
                "directory": {"type": "string", "description": "Working directory (default: .)"}
            },
            "required": ["command"]
        }
    }
}


@dataclass
class Session:
    """Represents an agent session with its own model and state."""
    name: str
    model_config: Dict
    messages: List[Dict] = field(default_factory=list)


class MultiModelOrchestrator:
    """
    Hybrid multi-model orchestrator using:
    - GPT-4o mini (OpenAI, essentially free tier) for fast inference
    - Local Ollama models (completely free) for specialized deep analysis
    """
    
    def __init__(self):
        self.client = OpenAI(api_key=os.getenv("OPENAI_API_KEY"))
        self.sessions: Dict[str, Session] = {}
        self.ollama_available = self._check_ollama()
    
    def _check_ollama(self) -> bool:
        """Check if Ollama is running locally."""
        try:
            resp = requests.get("http://localhost:11434/api/tags", timeout=2)
            return resp.status_code == 200
        except:
            return False
    
    def _call_ollama(self, model: str, prompt: str, system: str) -> str:
        """Call local Ollama instance - completely free."""
        payload = {
            "model": model,
            "prompt": prompt,
            "system": system,
            "stream": False,
            "options": {"temperature": 0.3}
        }
        try:
            resp = requests.post("http://localhost:11434/api/generate", 
                                json=payload, timeout=60)
            return resp.json().get("response", "Error: No response from Ollama")
        except Exception as e:
            return f"Ollama error: {e}. Fallback to GPT-4o mini."
    
    def _call_openai_react(self, model: str, messages: List[Dict], 
                          tools: Optional[List[Dict]] = None, max_iter: int = 5) -> str:
        """
        ReAct loop: Reason → Act (tool call) → Observe → Repeat
        """
        for i in range(max_iter):
            try:
                response = self.client.chat.completions.create(
                    model=model,
                    messages=messages,
                    tools=tools if tools else None,
                    temperature=0.3
                )
            except Exception as e:
                return f"OpenAI API error: {e}"
            
            choice = response.choices[0]
            messages.append({
                "role": "assistant",
                "content": choice.message.content or "",
                "tool_calls": [tc.model_dump() for tc in choice.message.tool_calls] if choice.message.tool_calls else []
            })
            
            # No tool calls = agent is done
            if not choice.message.tool_calls:
                return choice.message.content or "No response"
            
            # Execute tools and feed results back
            for tool_call in choice.message.tool_calls:
                tool_name = tool_call.function.name
                tool_args = json.loads(tool_call.function.arguments)
                
                print(f"  🔧 Tool call: {tool_name}({tool_args})")
                
                if tool_name == "pulse_sandbox":
                    result = self._run_pulse_sandbox(tool_args.get("command"), 
                                                     tool_args.get("directory", "."))
                else:
                    result = f"Unknown tool: {tool_name}"
                
                messages.append({
                    "role": "tool",
                    "tool_call_id": tool_call.id,
                    "content": result
                })
                print(f"  📊 Tool result: {result[:200]}...")
        
        return "Max iterations reached"
    
    def _run_pulse_sandbox(self, command: str, directory: str = ".") -> str:
        """Execute command via Go Pulse backend."""
        try:
            result = subprocess.run(
                ["go", "run", "cmd/pulse/main.go", "--json", 
                 "--cmd", command, "--dir", directory],
                cwd="./backend",
                capture_output=True,
                text=True,
                timeout=30
            )
            return result.stdout if result.returncode == 0 else result.stderr
        except subprocess.TimeoutExpired:
            return json.dumps({"error": "Pulse sandbox timed out"})
        except Exception as e:
            return json.dumps({"error": str(e)})
    
    def spawn(self, agent_name: str, task: str, 
              use_tools: bool = False) -> str:
        """
        Spawn an agent with appropriate model based on role.
        
        Args:
            agent_name: hunter, cloner, gatekeeper, advisor_deep, coder, reporter
            task: The task description
            use_tools: Whether this agent can use tools
        """
        config = MODEL_ROUTER.get(agent_name, {"type": "openai", "model": "gpt-4o-mini"})
        system_prompt = SYSTEM_PROMPTS.get(agent_name, "Be helpful.")
        
        print(f"\n🤖 Spawning [{agent_name}] using {config['type']}/{config['model']}")
        print(f"   Task: {task[:80]}...")
        
        if config["type"] == "ollama":
            if not self.ollama_available:
                print(f"   ⚠️ Ollama not available, falling back to GPT-4o mini")
                config = {"type": "openai", "model": "gpt-4o-mini"}
            else:
                return self._call_ollama(config["model"], task, system_prompt)
        
        # OpenAI ReAct loop
        messages = [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": task}
        ]
        
        tools = [PULSE_TOOL] if use_tools else None
        return self._call_openai_react(config["model"], messages, tools)
    
    def run_pipeline(self, repo_url: str = "github.com/example/bad-repo"):
        """Execute the full multi-model orchestration pipeline."""
        
        print("\n" + "="*60)
        print("🛡️ RepoGuard Multi-Model Orchestrator")
        print("   GPT-4o mini (cloud) + Ollama local models (free)")
        print("="*60)
        
        results = {}
        
        # Stage 1: Hunter (GPT-4o mini - fast search)
        results["repo"] = self.spawn(
            "hunter",
            f"Identify the target repository: {repo_url}. Return just the URL."
        )
        print(f"✅ Hunter found: {results['repo'][:100]}")
        
        # Stage 2: Cloner (GPT-4o mini - command identification)
        results["command"] = self.spawn(
            "cloner",
            f"For repo {results['repo']}, identify a risky cleanup command like 'rm -rf node_modules' or 'git clean -fdx'. Return just the command."
        )
        print(f"✅ Cloner identified command: {results['command'][:100]}")
        
        # Stage 3: Gatekeeper (GPT-4o mini - with tool calling)
        results["risk"] = self.spawn(
            "gatekeeper",
            f"Execute this command through Pulse sandbox: {results['command']}",
            use_tools=True
        )
        print(f"✅ Gatekeeper analysis complete")
        
        # Stage 4: Deep Advisor (Ollama llama3 - free deep analysis)
        if "HIGH" in results["risk"].upper() or "CRITICAL" in results["risk"].upper():
            print("\n⚠️ High risk detected - spawning local deep advisor...")
            results["deep_analysis"] = self.spawn(
                "advisor_deep",
                f"Command: {results['command']}\nRisk Assessment: {results['risk']}\n\nExplain: 1) Technical blast radius, 2) Business impact, 3) Self-healing recommendations."
            )
            print(f"✅ Deep advisor analysis complete")
        
        # Stage 5: Coder (Ollama codellama - free code analysis)
        print("\n🔍 Running local code analysis...")
        results["code_analysis"] = self.spawn(
            "coder",
            f"Analyze this command for hidden infrastructure risks: {results['command']}. Check for: filesystem operations, network calls, privilege escalation."
        )
        print(f"✅ Code analysis complete")
        
        # Stage 6: Reporter (GPT-4o mini - cheap summarization)
        results["report"] = self.spawn(
            "reporter",
            f"Compile final audit report:\nRepository: {results['repo']}\nCommand: {results['command']}\nRisk: {results.get('risk', 'Unknown')}\nDeep Analysis: {results.get('deep_analysis', 'N/A')}\nCode Analysis: {results.get('code_analysis', 'N/A')}\n\nWrite a concise executive summary with action items."
        )
        
        return results


def main():
    # Check API key
    if not os.getenv("OPENAI_API_KEY"):
        print("⚠️ WARNING: OPENAI_API_KEY not found. Set it with: export OPENAI_API_KEY='your-key'")
        print("   Or create a .env file with OPENAI_API_KEY=your-key")
        return
    
    # Initialize orchestrator
    orchestrator = MultiModelOrchestrator()
    
    if orchestrator.ollama_available:
        print("✅ Ollama detected - local models available (completely free)")
    else:
        print("ℹ️ Ollama not running. Install with: curl -fsSL https://ollama.com/install.sh | sh")
        print("   Then: ollama pull llama3 codellama")
    
    # Run the pipeline
    results = orchestrator.run_pipeline()
    
    # Final output
    print("\n" + "="*60)
    print("� FINAL AUDIT REPORT")
    print("="*60)
    print(f"Models used: GPT-4o mini (cloud, ~free) + Ollama (local, free)")
    print(f"Repository: {results.get('repo', 'N/A')}")
    print(f"Command Analyzed: {results.get('command', 'N/A')}")
    print(f"\n--- RISK ASSESSMENT ---")
    print(results.get('risk', 'No risk data')[:500])
    if 'deep_analysis' in results:
        print(f"\n--- DEEP ANALYSIS (Local llama3) ---")
        print(results['deep_analysis'][:500])
    print(f"\n--- CODE ANALYSIS (Local codellama) ---")
    print(results.get('code_analysis', 'N/A')[:500])
    print(f"\n--- EXECUTIVE SUMMARY ---")
    print(results.get('report', 'N/A'))
    print("="*60)


if __name__ == "__main__":
    main()
