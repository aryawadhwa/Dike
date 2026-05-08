package repl

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/aryawadhwa/dike/pkg/agents"
	"github.com/aryawadhwa/dike/pkg/ghost"
	"github.com/aryawadhwa/dike/pkg/pipeline"
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

	fmt.Print(`
    ____  __  __ __    _____ ______
   / __ \/ / / / /    / ___// ____/
  / /_/ / / / / /     \__ \/ __/   
 / ____/ /_/ / /___  ___/ / /___   
/_/    \____/_____/ /____/_____/   
                                   
      [ SECURE SHELL v2.0 ]
`)
	fmt.Println("Welcome to Pulse (Dike) - Type 'exit' or 'quit' to leave.")

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

		// Build pipeline once - shared between REPL and headless modes
		pulsePipeline := buildPipeline(pol)

		// Execute Railway-Oriented Pipeline with immutable context
		ctx := pipeline.NewContext(input, pol)
		finalCtx, err := pipeline.Execute(ctx, pulsePipeline...)

		if err != nil {
			if pipeline.IsDenyError(err) {
				fmt.Printf("❌ %v\n", err)
				// Audit the denial
				_, _ = pipeline.Execute(finalCtx, agents.Auditor())
				continue
			}
			fmt.Printf("Pipeline error: %v\n", err)
			continue
		}

		// Handle decision
		switch finalCtx.Decision {
		case pipeline.DecisionAllow:
			executeHostCommand(input)
			_, _ = pipeline.Execute(finalCtx, agents.Auditor())

		case pipeline.DecisionPreview:
			// Show sandbox results
			fmt.Printf("\n📋 Sandbox Preview:\n")
			if len(finalCtx.Sandbox.FileChanges) > 0 {
				fmt.Println("File changes detected:")
				for _, change := range finalCtx.Sandbox.FileChanges {
					fmt.Printf("  %s: %s\n", change.Type, change.Path)
				}
			} else {
				fmt.Println("No file changes detected.")
			}

			// Human-in-the-loop
			for {
				fmt.Print("\nApply to host? [y/N/e (explain)]: ")
				ans, _ := line.Prompt("")
				ans = strings.ToLower(strings.TrimSpace(ans))

				if ans == "e" || ans == "explain" {
					_, _ = pipeline.Execute(finalCtx, agents.Advisor())
				} else if ans == "y" || ans == "yes" {
					fmt.Println("Executing on host...")
					executeHostCommand(input)
					// Audit the applied decision
					appliedCtx := finalCtx.WithDecision(pipeline.DecisionAllow, "user approved")
					_, _ = pipeline.Execute(appliedCtx, agents.Auditor())
					break
				} else {
					fmt.Println("Discarding sandbox changes.")
					rejectedCtx := finalCtx.WithDecision(pipeline.DecisionDeny, "user rejected")
					_, _ = pipeline.Execute(rejectedCtx, agents.Auditor())
					break
				}
			}
		}
	}
}

// buildPipeline creates the shared agent pipeline used by both REPL and headless modes.
// This eliminates the duplication between interactive and non-interactive execution.
func buildPipeline(pol *policy.Policy) []pipeline.Agent {
	return []pipeline.Agent{
		agents.Gatekeeper(),  // Evaluate against policy
		agents.Ghost(),       // Sandbox if preview required
	}
}

func executeHostCommand(cmdString string) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", cmdString)
	} else {
		cmd = exec.Command("bash", "-c", cmdString)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("Command error: %v\n", err)
	}
}
