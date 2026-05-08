package main

import (
	"fmt"
	"github.com/aryawadhwa/dike/pkg/gatekeeper"
)

func main() {
	commands := []string{
		"tar -xzf archive.tar.gz",
		"grep -f patterns.txt output.log",
		"ls -d */",
		"git clean -fdx",
		"rm -rf /",
		"sudo apt update",
	}

	for _, cmd := range commands {
		decision, cap := gatekeeper.Evaluate(cmd, nil)
		fmt.Printf("Command: %s\nDecision: %s\nCapability: %s\n\n", cmd, decision, cap)
	}

}
