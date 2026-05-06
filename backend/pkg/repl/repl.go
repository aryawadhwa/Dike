package repl

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aryawadhwa/dike/pkg/ghost"
	"github.com/aryawadhwa/dike/pkg/orchestrator"
	"github.com/aryawadhwa/dike/pkg/policy"
	"github.com/peterh/liner"
)

// Start initializes and runs the basic Pulse REPL.
func Start() {


	pol, err := policy.LoadPolicy("configs/SAFETY.yaml")
	if err != nil {
		fmt.Printf("Warning: Failed to load policy: %v. Running without policy.\n", err)
	} else {
		fmt.Printf("Loaded policy with %d rules.\n", len(pol.Rules))
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Warning: Failed to get current working directory: %v\n", err)
	} else {
		fmt.Printf("Initializing Ghost Sandbox... ")
		if err := ghost.InitSandbox(cwd); err != nil {
			fmt.Printf("Failed: %v\n", err)
		} else {
			fmt.Println("Done.")
			defer ghost.Teardown()
		}
	}

	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)

	fmt.Println("Welcome to Pulse (Dike) - The Safe Command Shell")
	fmt.Println("Type 'exit' or 'quit' to leave.")

	for {
		input, err := line.Prompt("pulse> ")
		if err != nil {
			if err == liner.ErrPromptAborted {
				fmt.Println("Aborted")
				break
			}
			fmt.Printf("Error reading line: %v\n", err)
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		line.AppendHistory(input)

		if input == "exit" || input == "quit" {
			break
		}

		// Construct the Multi-Agent Context
		ctx := &orchestrator.Context{
			Command: input,
			Policy:  pol,
		}

		// Stage 1: Intercept & Evaluate
		stage1 := []orchestrator.Agent{
			&orchestrator.GatekeeperAgent{},
		}
		orchestrator.RunPipeline(stage1, ctx)

		if ctx.Decision == string(policy.DecisionDeny) {
			ctx.Decision = "REJECTED"
			orchestrator.RunPipeline([]orchestrator.Agent{&orchestrator.AuditorAgent{}}, ctx)
			continue
		}

		if ctx.Decision == string(policy.DecisionAllow) {
			executeHostCommand(input)
			ctx.Decision = "APPLIED"
			orchestrator.RunPipeline([]orchestrator.Agent{&orchestrator.AuditorAgent{}}, ctx)
			continue
		}

		// Stage 2: Sandbox Preview
		stage2 := []orchestrator.Agent{
			&orchestrator.GhostAgent{},
			&orchestrator.DiffAgent{},
		}
		err = orchestrator.RunPipeline(stage2, ctx)
		if err != nil {
			fmt.Printf("Pipeline error: %v\n", err)
			ctx.Decision = "ERROR"
			orchestrator.RunPipeline([]orchestrator.Agent{&orchestrator.AuditorAgent{}}, ctx)
			continue
		}

		// Stage 3: Human-in-the-Loop & Advisor
		for {
			fmt.Print("\nApply to host? [y/N/e (explain)]: ")
			ans, _ := line.Prompt("")
			ans = strings.ToLower(strings.TrimSpace(ans))

			if ans == "e" || ans == "explain" {
				adv := &orchestrator.AdvisorAgent{PromptResponse: ans}
				orchestrator.RunPipeline([]orchestrator.Agent{adv}, ctx)
			} else if ans == "y" || ans == "yes" {
				fmt.Println("Executing on host...")
				executeHostCommand(input)
				ctx.Decision = "APPLIED"
				orchestrator.RunPipeline([]orchestrator.Agent{&orchestrator.AuditorAgent{}}, ctx)
				break
			} else {
				fmt.Println("Discarding sandbox changes.")
				ctx.Decision = "REJECTED"
				orchestrator.RunPipeline([]orchestrator.Agent{&orchestrator.AuditorAgent{}}, ctx)
				break
			}
		}
	}
}

func executeHostCommand(cmdString string) {
	// Naive execution using bash -c
	cmd := exec.Command("bash", "-c", cmdString)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("Command error: %v\n", err)
	}
}
