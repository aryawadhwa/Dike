package repl

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aryawadhwa/dike/pkg/agents"
	"github.com/aryawadhwa/dike/pkg/pipeline"
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

// HeadlessExecute runs a command non-interactively using the Railway-Oriented Pipeline.
// This is now a simple wrapper around the shared pipeline - no duplication with REPL.
func HeadlessExecute(cmdStr string, dir string) {
	// Default output structure
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

	// Load policy
	pol, err := policy.LoadPolicy("configs/SAFETY.yaml")
	if err != nil {
		out.Error = fmt.Sprintf("Failed to load policy: %v", err)
		printJSON(out)
		return
	}

	// Use shared pipeline - same as REPL mode
	pulsePipeline := buildPipeline(pol)

	// Execute Railway-Oriented Pipeline
	ctx := pipeline.NewContext(cmdStr, pol)
	finalCtx, err := pipeline.Execute(ctx, pulsePipeline...)

	// Handle pipeline outcomes
	if err != nil {
		if denyErr, ok := err.(*pipeline.DenyError); ok {
			out.RiskLevel = "CRITICAL"
			out.Explanation = denyErr.Reason
			printJSON(out)
			// Audit the denial
			_, _ = pipeline.Execute(finalCtx, agents.Auditor())
			return
		}
		out.Error = err.Error()
		printJSON(out)
		return
	}

	// Build output based on final context decision
	switch finalCtx.Decision {
	case pipeline.DecisionDeny:
		out.RiskLevel = "CRITICAL"
		out.Explanation = finalCtx.DenyReason

	case pipeline.DecisionPreview:
		out.RiskLevel = "HIGH"
		out.Stdout = finalCtx.Sandbox.Stdout
		
		// Populate diff summary from immutable context
		for _, change := range finalCtx.Sandbox.FileChanges {
			switch change.Type {
			case "created":
				out.DiffSummary.Created = append(out.DiffSummary.Created, change.Path)
			case "deleted":
				out.DiffSummary.Deleted = append(out.DiffSummary.Deleted, change.Path)
			case "modified":
				out.DiffSummary.Modified = append(out.DiffSummary.Modified, change.Path)
			}
		}
		out.Explanation = "Sandbox preview completed. Review file changes."

	case pipeline.DecisionAllow:
		out.RiskLevel = "LOW"
		out.Explanation = "Command is safe and allowed by policy."
	}

	// Audit the execution
	_, _ = pipeline.Execute(finalCtx, agents.Auditor())

	printJSON(out)
}

func printJSON(out HeadlessOutput) {
	b, _ := json.Marshal(out)
	fmt.Println(string(b))
}
