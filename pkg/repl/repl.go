package repl

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aryawadhwa/dike/pkg/gatekeeper"
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

		// Phase 1: Gatekeeper evaluation
		decision := gatekeeper.Evaluate(input, pol)
		switch decision {
		case policy.DecisionDeny:
			fmt.Printf("❌ Gatekeeper DENIED this command.\n")
		case policy.DecisionPreview:
			fmt.Printf("⚠️  Gatekeeper PREVIEW required for this command. (Ghost Sandbox coming in Phase 2)\n")
		default:
			// DecisionAllow
			executeHostCommand(input)
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
