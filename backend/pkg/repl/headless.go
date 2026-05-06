package repl

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aryawadhwa/dike/pkg/advisor"
	"github.com/aryawadhwa/dike/pkg/diff"
	"github.com/aryawadhwa/dike/pkg/gatekeeper"
	"github.com/aryawadhwa/dike/pkg/ghost"
	"github.com/aryawadhwa/dike/pkg/policy"
)

// HeadlessOutput is the JSON structure returned by the PulseSkill
type HeadlessOutput struct {
	RiskLevel   string   `json:"risk_level"`
	DiffSummary DiffJSON `json:"diff_summary"`
	Stdout      string   `json:"stdout"`
	Explanation string   `json:"explanation"`
	Error       string   `json:"error,omitempty"`
}

type DiffJSON struct {
	Created  []string `json:"created"`
	Deleted  []string `json:"deleted"`
	Modified []string `json:"modified"`
}

// HeadlessExecute runs a command non-interactively and outputs JSON
func HeadlessExecute(cmdStr string, dir string) {
	out := HeadlessOutput{
		RiskLevel: "UNKNOWN",
		DiffSummary: DiffJSON{
			Created:  []string{},
			Deleted:  []string{},
			Modified: []string{},
		},
	}

	if dir == "" {
		cwd, _ := os.Getwd()
		dir = cwd
	}

	pol, err := policy.LoadPolicy("configs/SAFETY.yaml")
	if err != nil {
		out.Error = fmt.Sprintf("Failed to load policy: %v", err)
		printJSON(out)
		return
	}

	if err := ghost.InitSandbox(dir); err != nil {
		out.Error = fmt.Sprintf("Failed to init sandbox: %v", err)
		printJSON(out)
		return
	}
	defer ghost.Teardown()

	decision := gatekeeper.Evaluate(cmdStr, pol)
	
	switch decision {
	case policy.DecisionDeny:
		out.RiskLevel = "CRITICAL"
		out.Explanation = advisor.ExplainCommand(cmdStr)
	case policy.DecisionPreview:
		out.RiskLevel = "HIGH"
		
		stdout, err := ghost.ExecPreview(cmdStr)
		out.Stdout = stdout
		if err != nil {
			out.Error = err.Error()
		}

		diffSummary, err := diff.ComputeDiff("pulse-ghost")
		if err == nil {
			out.DiffSummary.Created = diffSummary.Added
			out.DiffSummary.Deleted = diffSummary.Deleted
		}

		out.Explanation = advisor.ExplainCommand(cmdStr)
	case policy.DecisionAllow:
		out.RiskLevel = "LOW"
		out.Explanation = "Command is safe and allowed by policy."
	}

	printJSON(out)
}

func printJSON(out HeadlessOutput) {
	b, _ := json.Marshal(out)
	fmt.Println(string(b))
}
