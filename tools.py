import subprocess
import json
from langchain.tools import tool

@tool("PulseSkill")
def run_pulse_sandbox(command: str, directory: str = ".") -> str:
    """
    Executes a potentially risky shell command inside the Pulse Docker Sandbox.
    Returns a JSON string containing the risk_level, diff_summary, stdout, and explanation.
    Always use this tool before executing any build or clean commands.
    """
    try:
        # Run Pulse with JSON output flag
        result = subprocess.run(
            ["./backend/pulse", "--json", "--cmd", command, "--dir", directory], 
            capture_output=True, 
            text=True,
            check=True
        )
        return result.stdout
    except subprocess.CalledProcessError as e:
        return json.dumps({"status": "error", "message": str(e), "output": e.stdout})
    except FileNotFoundError:
        return json.dumps({"status": "error", "message": "Pulse binary not found. Please build it first with 'go build -o pulse cmd/pulse/main.go'"})
