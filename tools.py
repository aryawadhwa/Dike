import subprocess
import json

def run_pulse_sandbox(command: str, directory: str = ".") -> str:
    """
    Executes a potentially risky shell command inside the Pulse Docker Sandbox.
    Returns a JSON string containing the risk_level, diff_summary, stdout, and explanation.
    Always use this tool before executing any build or clean commands.
    """
    try:
        result = subprocess.run(
            ["go", "run", "cmd/pulse/main.go", "--json", "--cmd", command, "--dir", directory],
            cwd="./backend",
            capture_output=True,
            text=True,
            check=True
        )
        return result.stdout
    except subprocess.CalledProcessError as e:
        return json.dumps({"status": "error", "message": str(e), "output": e.stdout})
