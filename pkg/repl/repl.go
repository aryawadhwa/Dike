package repl

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/peterh/liner"
)

// Start initializes and runs the basic Pulse REPL.
func Start() {
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

		// Phase 0: Just execute directly without gatekeeper
		executeHostCommand(input)
	}
}

func executeHostCommand(cmdString string) {
	// Naive execution using bash -c for Phase 0
	cmd := exec.Command("bash", "-c", cmdString)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("Command error: %v\n", err)
	}
}
