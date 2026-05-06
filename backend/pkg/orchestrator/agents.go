package orchestrator

import (
	"fmt"
	"strings"

	"github.com/aryawadhwa/dike/pkg/advisor"
	"github.com/aryawadhwa/dike/pkg/audit"
	"github.com/aryawadhwa/dike/pkg/diff"
	"github.com/aryawadhwa/dike/pkg/gatekeeper"
	"github.com/aryawadhwa/dike/pkg/ghost"
	"github.com/aryawadhwa/dike/pkg/policy"
)

// GatekeeperAgent intercepts and evaluates risk
type GatekeeperAgent struct{}

func (g *GatekeeperAgent) Name() string { return "Gatekeeper Agent" }
func (g *GatekeeperAgent) Execute(ctx *Context) (Result, error) {
	pol := ctx.Policy.(*policy.Policy)
	decision := gatekeeper.Evaluate(ctx.Command, pol)
	ctx.Decision = string(decision)
	
	if decision == policy.DecisionDeny {
		fmt.Printf("❌ Gatekeeper DENIED this command.\n")
		return Result{Message: "Command denied."}, nil
	} else if decision == policy.DecisionAllow {
		ctx.IsSafe = true
		return Result{Message: "Command is safe."}, nil
	}
	
	fmt.Printf("⚠️  Gatekeeper required PREVIEW. Routing to Ghost Agent...\n")
	return Result{Message: "Routing to sandbox."}, nil
}

// GhostAgent simulates the blast radius
type GhostAgent struct{}

func (g *GhostAgent) Name() string { return "Ghost Sandbox Agent" }
func (g *GhostAgent) Execute(ctx *Context) (Result, error) {
	output, err := ghost.ExecPreview(ctx.Command)
	if err != nil {
		ctx.Decision = "ERROR"
		return Result{}, fmt.Errorf("ghost execution failed: %v", err)
	}
	ctx.GhostOutput = output
	fmt.Printf("--- Ghost Output ---\n%s\n--------------------\n", output)
	return Result{Message: "Sandbox execution complete."}, nil
}

// DiffAgent analyzes the damage
type DiffAgent struct{}

func (d *DiffAgent) Name() string { return "Diff Analyzer Agent" }
func (d *DiffAgent) Execute(ctx *Context) (Result, error) {
	diffSummary, err := diff.ComputeDiff("pulse-ghost")
	if err != nil {
		return Result{}, fmt.Errorf("diff computation failed: %v", err)
	}
	ctx.DiffSummary = diffSummary.String()
	fmt.Printf("\n--- Diff Preview ---\n%s\n--------------------\n", ctx.DiffSummary)
	return Result{Message: "Diff generated."}, nil
}

// AdvisorAgent provides LLM guidance
type AdvisorAgent struct {
	PromptResponse string // We inject the user's interactive response here
}

func (a *AdvisorAgent) Name() string { return "Advisor Agent" }
func (a *AdvisorAgent) Execute(ctx *Context) (Result, error) {
	// The REPL loop handles the interactive prompt. This agent is called if the user typed 'e'
	if strings.ToLower(a.PromptResponse) == "e" || strings.ToLower(a.PromptResponse) == "explain" {
		explanation := advisor.ExplainCommand(ctx.Command)
		fmt.Println("\n" + explanation)
	}
	return Result{Message: "Advisor consulted."}, nil
}

// AuditorAgent logs the final decision
type AuditorAgent struct{}

func (a *AuditorAgent) Name() string { return "Auditor Agent" }
func (a *AuditorAgent) Execute(ctx *Context) (Result, error) {
	_ = audit.LogDecision(ctx.Command, ctx.Decision, ctx.Decision, ctx.DiffSummary)
	return Result{Message: "Audit log recorded."}, nil
}
