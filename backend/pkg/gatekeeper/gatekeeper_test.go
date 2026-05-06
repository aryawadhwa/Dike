package gatekeeper

import (
	"testing"

	"github.com/aryawadhwa/dike/pkg/policy"
)

func TestEvaluate(t *testing.T) {
	pol := &policy.Policy{
		Rules: []policy.Rule{
			{
				Name:   "Block rm -rf",
				Action: policy.DecisionDeny,
				Match: policy.Match{
					Commands: []string{"rm"},
					Args:     []string{"-rf", "/*"},
				},
			},
			{
				Name:   "Preview apt",
				Action: policy.DecisionPreview,
				Match: policy.Match{
					Commands: []string{"apt", "apt-get"},
					Args:     []string{"install"},
				},
			},
		},
	}

	tests := []struct {
		name     string
		cmd      string
		expected policy.Decision
	}{
		{"Safe command", "ls -la", policy.DecisionAllow},
		{"Safe rm", "rm file.txt", policy.DecisionAllow},
		{"Denied command", "rm -rf /", policy.DecisionDeny},
		{"Denied command with options", "rm -rf /*", policy.DecisionDeny},
		{"Preview command", "apt install curl", policy.DecisionPreview},
		{"Multiple commands safe", "echo hello && ls", policy.DecisionAllow},
		{"Multiple commands denied", "echo hello && rm -rf /", policy.DecisionDeny},
		{"Piped commands denied", "ls | xargs rm -rf", policy.DecisionDeny},
		{"Malformed command", "if [ ; then", policy.DecisionDeny}, // syntax error
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Evaluate(tt.cmd, pol)
			if result != tt.expected {
				t.Errorf("Evaluate(%q) = %v; expected %v", tt.cmd, result, tt.expected)
			}
		})
	}
}
