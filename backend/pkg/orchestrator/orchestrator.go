package orchestrator

import "fmt"

// Agent defines the standard interface for the PulseFlow Multi-Agent orchestrator.
type Agent interface {
	Execute(ctx *Context) (Result, error)
	Name() string
}

// Context holds the shared state as the command moves through the agent pipeline.
type Context struct {
	Command      string
	Policy       interface{} // The loaded SAFETY.yaml policy
	RiskLevel    string
	Decision     string      // "ALLOW", "PREVIEW", "DENY", "APPLIED", "REJECTED"
	DiffSummary  string
	GhostOutput  string
	IsSafe       bool
}

// Result is the outcome of a single agent's execution.
type Result struct {
	Message string
	Data    interface{}
}

// RunPipeline executes the sequence of agents.
// It acts as the central brain ("The Director") coordinating the multi-agent system.
func RunPipeline(agents []Agent, ctx *Context) error {
	for _, agent := range agents {
		result, err := agent.Execute(ctx)
		if err != nil {
			return fmt.Errorf("[%s] failed: %w", agent.Name(), err)
		}
		
		if result.Message != "" {
			fmt.Printf("--- [%s]: %s\n", agent.Name(), result.Message)
		}

		// If Gatekeeper decided it's safe or denied, we might short-circuit the pipeline
		if ctx.Decision == "ALLOW" && agent.Name() == "Gatekeeper Agent" {
			ctx.IsSafe = true
		}
	}
	return nil
}
