package policy

import (
	"os"
	"testing"
)

func TestLoadPolicy(t *testing.T) {
	yamlContent := `
rules:
  - name: "Test Rule"
    action: DENY
    match:
      commands: ["rm"]
      args: ["-rf"]
`
	tmpfile, err := os.CreateTemp("", "safety.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(yamlContent)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	pol, err := LoadPolicy(tmpfile.Name())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(pol.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(pol.Rules))
	}

	if pol.Rules[0].Name != "Test Rule" {
		t.Errorf("Expected name 'Test Rule', got %s", pol.Rules[0].Name)
	}
	if pol.Rules[0].Action != DecisionDeny {
		t.Errorf("Expected action DENY, got %s", pol.Rules[0].Action)
	}
}
