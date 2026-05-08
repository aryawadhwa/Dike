package policy

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Decision string
type Capability string

const (
	DecisionAllow   Decision = "ALLOW"
	DecisionPreview Decision = "PREVIEW"
	DecisionDeny    Decision = "DENY"

	CapMassDelete    Capability = "MASS_DELETE"
	CapSystemModify  Capability = "SYSTEM_MODIFY"
	CapNetworkExfil  Capability = "NETWORK_EXFIL"
	CapExecArbitrary Capability = "EXEC_ARBITRARY"
)

type Match struct {
	Capability Capability `yaml:"capability"`
	Commands   []string   `yaml:"commands,omitempty"`
	Args       []string   `yaml:"args,omitempty"`
	Context    string     `yaml:"context,omitempty"`
}

type Rule struct {
	Name       string   `yaml:"name,omitempty"`
	Capability string   `yaml:"capability,omitempty"`
	Commands   []string `yaml:"commands,omitempty"`
}

type Policy struct {
	DefaultAction Decision `yaml:"default_action"`
	Allow         []Rule   `yaml:"allow"`
	Deny          []Rule   `yaml:"deny"`
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
