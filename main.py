from crewai import Agent, Task, Crew, Process
from langchain_openai import ChatOpenAI
from tools import run_pulse_sandbox
import os

# Initialize LLM for the Agents (Uses OpenAI API Key from environment)
llm = ChatOpenAI(model="gpt-4o-mini", temperature=0.7)

print("🛡️ Initializing RepoGuard CrewAI Orchestrator...")

# 1. The Repo Hunter (Scout)
hunter = Agent(
    role='Repository Scout',
    goal='Find relevant GitHub repositories for analysis',
    backstory='You are an expert at finding high-quality open-source projects. You currently only output a mock repository for safety.',
    verbose=True,
    llm=llm
)

# 2. The Cloner Agent
cloner = Agent(
    role='Environment Prep Engineer',
    goal='Identify the best build and clean commands for a given repository.',
    backstory='You analyze infrastructure files to determine how to build or clean a repo.',
    verbose=True,
    llm=llm
)

# 3. The Builder (Risk Assessor)
builder = Agent(
    role='Build Safety Engineer',
    goal='Execute build commands safely using the Pulse Sandbox.',
    backstory='You focus on build safety and configuration integrity. You NEVER run a command without sending it through the Pulse Sandbox first.',
    verbose=True,
    tools=[run_pulse_sandbox], # Give it access to Pulse
    llm=llm
)

# 4. The Reporter (Auditor)
reporter = Agent(
    role='Security Reporter',
    goal='Compile findings from the Builder into a readable risk report.',
    backstory='You translate technical risks into actionable advice for developers.',
    verbose=True,
    llm=llm
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
