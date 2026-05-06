package repl

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aryawadhwa/dike/pkg/audit"
	"github.com/aryawadhwa/dike/pkg/advisor"
	"github.com/aryawadhwa/dike/pkg/diff"
	"github.com/aryawadhwa/dike/pkg/gatekeeper"
	"github.com/aryawadhwa/dike/pkg/ghost"
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

		decision := gatekeeper.Evaluate(input, pol)
		switch decision {
		case policy.DecisionDeny:
			fmt.Printf("❌ Gatekeeper DENIED this command.\n")
			_ = audit.LogDecision(input, "DENY", "REJECTED", "")
		case policy.DecisionPreview:
			fmt.Printf("⚠️  Gatekeeper required PREVIEW. Executing in Ghost Sandbox...\n")
			output, err := ghost.ExecPreview(input)
			if err != nil {
				fmt.Printf("Ghost Engine error: %v\n", err)
				_ = audit.LogDecision(input, "PREVIEW", "ERROR", err.Error())
			} else {
				fmt.Printf("--- Ghost Output ---\n%s\n--------------------\n", output)
				
				// Compute Diff
				diffSummary, err := diff.ComputeDiff("pulse-ghost")
				var diffStr string
				if err != nil {
					diffStr = fmt.Sprintf("Error computing diff: %v", err)
					fmt.Println(diffStr)
				} else {
					diffStr = diffSummary.String()
					fmt.Printf("\n--- Diff Preview ---\n%s\n--------------------\n", diffStr)
				}

				// Prompt User
				for {
					fmt.Print("\nApply to host? [y/N/e (explain)]: ")
					ans, _ := line.Prompt("")
					ans = strings.ToLower(strings.TrimSpace(ans))
					
					if ans == "e" || ans == "explain" {
						fmt.Println("\n" + advisor.ExplainCommand(input))
						// Loop continues to ask again
					} else if ans == "y" || ans == "yes" {
						fmt.Println("Executing on host...")
						executeHostCommand(input)
						_ = audit.LogDecision(input, "PREVIEW", "APPLIED", diffStr)
						break
					} else {
						fmt.Println("Discarding sandbox changes.")
						_ = audit.LogDecision(input, "PREVIEW", "REJECTED", diffStr)
						break
					}
				}
			}
		default:
			// DecisionAllow
			executeHostCommand(input)
			_ = audit.LogDecision(input, "ALLOW", "APPLIED", "")
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
