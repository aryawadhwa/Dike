package policy

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Decision string

const (
	DecisionAllow   Decision = "ALLOW"
	DecisionPreview Decision = "PREVIEW"
	DecisionDeny    Decision = "DENY"
)

type Match struct {
	Commands []string `yaml:"commands"`
	Args     []string `yaml:"args"`
}

type Rule struct {
	Name   string   `yaml:"name"`
	Action Decision `yaml:"action"`
	Match  Match    `yaml:"match"`
}

type Policy struct {
	Rules []Rule `yaml:"rules"`
}

// LoadPolicy parses a YAML policy file from the given path
func LoadPolicy(path string) (*Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy file: %w", err)
	}

	var p Policy
	err = yaml.Unmarshal(data, &p)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal policy: %w", err)
	}

	return &p, nil
}
