package gatekeeper

import "github.com/aryawadhwa/dike/pkg/policy"

// Signature defines how a command invokes a capability
type Signature struct {
	// Command is the base binary (e.g., "git", "rm", "find")
	Command string
	
	// Subcommand is the first argument (e.g., "clean" for git clean)
	// Leave empty if not applicable (e.g., for "rm")
	Subcommand string
	
	// DangerousFlags are flags that ACTIVATE the dangerous capability
	// If empty, the base command is ALWAYS considered dangerous (e.g., chmod)
	DangerousFlags []string
}

// DestructiveCapabilities maps capabilities to the command signatures that trigger them
var DestructiveCapabilities = map[policy.Capability][]Signature{
	policy.CapMassDelete: {
		{Command: "rm", DangerousFlags: []string{"r", "f", "rf", "fr"}},
		{Command: "git", Subcommand: "clean"}, // git clean is inherently a mass-delete
		{Command: "git", Subcommand: "reset", DangerousFlags: []string{"hard"}},
		{Command: "find", DangerousFlags: []string{"delete", "exec rm"}},
		{Command: "truncate"}, 
	},
	policy.CapSystemModify: {
		{Command: "chmod", DangerousFlags: []string{"R"}}, // Recursive permission changes
		{Command: "chown", DangerousFlags: []string{"R"}},
		{Command: "mkfs"},
		{Command: "dd", DangerousFlags: []string{"if=/dev", "of=/dev"}},
	},
	policy.CapExecArbitrary: {
		{Command: "sh", DangerousFlags: []string{"c"}},
		{Command: "bash", DangerousFlags: []string{"c"}},
		{Command: "zsh", DangerousFlags: []string{"c"}},
		{Command: "python", DangerousFlags: []string{"c"}},
		{Command: "python3", DangerousFlags: []string{"c"}},
		{Command: "node", DangerousFlags: []string{"e"}},
	},
}
