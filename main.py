from crewai import Agent, Task, Crew, Process
from tools import run_pulse_sandbox
import os
from dotenv import load_dotenv

load_dotenv()

print("🛡️ Initializing RepoGuard CrewAI Orchestrator...")

# 1. The Repo Hunter (Scout)
hunter = Agent(
    role='Repository Scout',
    goal='Find relevant GitHub repositories for analysis',
    backstory='You are an expert at finding high-quality open-source projects. You currently only output a mock repository for safety.',
    verbose=True,
    llm="openai/gpt-4o-mini"
)

# 2. The Cloner Agent
cloner = Agent(
    role='Environment Prep Engineer',
    goal='Identify the best build and clean commands for a given repository.',
    backstory='You analyze infrastructure files to determine how to build or clean a repo.',
    verbose=True,
    llm="openai/gpt-4o-mini"
)

# 3. The Builder (Risk Assessor)
builder = Agent(
    role='Build Safety Engineer',
    goal='Execute build commands safely using the Pulse Sandbox.',
    backstory='You focus on build safety and configuration integrity. You NEVER run a command without sending it through the Pulse Sandbox first.',
    verbose=True,
    tools=[run_pulse_sandbox], # Give it access to Pulse
    llm="openai/gpt-4o-mini"
)

# 4. The Reporter (Auditor)
reporter = Agent(
    role='Security Reporter',
    goal='Compile findings from the Builder into a readable risk report.',
    backstory='You translate technical risks into actionable advice for developers.',
    verbose=True,
    llm="openai/gpt-4o-mini"
)

# Define Tasks
scan_task = Task(
    description="Output the string 'github.com/example/bad-repo' as the target repository for this test run.",
    expected_output="A single repository URL string.",
    agent=hunter
)

clone_task = Task(
    description="For the repo found, identify a risky command that might be used to clean the environment, such as 'rm -rf node_modules' or 'git clean -fdx'.",
    expected_output="A single terminal command string.",
    agent=cloner,
    context=[scan_task]
)

analyze_task = Task(
    description="Take the command identified by the cloner and run it through the PulseSkill tool. Return the JSON output.",
    expected_output="A JSON object containing risk level and file diff.",
    agent=builder,
    context=[clone_task]
)

report_task = Task(
    description="Summarize the JSON risks found by the Builder into a final audit email draft.",
    expected_output="A string containing the email body.",
    agent=reporter,
    context=[analyze_task]
)

# Kickoff
crew = Crew(
    agents=[hunter, cloner, builder, reporter],
    tasks=[scan_task, clone_task, analyze_task, report_task],
    process=Process.sequential,
    memory=True,  # Enables the "Sensing" memory layer
    cache=True,   # Improves performance
    verbose=True
)

if __name__ == "__main__":
    if "OPENAI_API_KEY" not in os.environ:
        print("⚠️ WARNING: OPENAI_API_KEY not found in environment. CrewAI requires this to function.")
    
    print("\n🚀 Starting CrewAI Run...")
    result = crew.kickoff()
    
    print("\n==============================================")
    print("📝 FINAL AUDIT REPORT")
    print("==============================================")
    print(result)
