// Package pipeline executes agents in sequence with railway-oriented error handling
package pipeline

import (
	"time"
)

// Execute runs the context through a series of agents.
// If any agent returns an error, the pipeline halts and returns the error.
// This is "railway-oriented programming" - success stays on the main track,
// errors branch to the error track.
func Execute(ctx Context, agents ...Agent) (Context, error) {
	var err error
	
	for _, agent := range agents {
		// Short-circuit if already denied
		if err != nil {
			return ctx, err
		}
		
		// Skip remaining agents if already denied in context
		if ctx.IsDenied() {
			return ctx, DenyWithRule("pipeline short-circuited due to prior denial", "system")
		}
		
		start := time.Now()
		input := ctx // capture input for audit trail
		
		// Execute agent
		ctx, err = agent(ctx)
		
		// Record step in audit trail
		step := AgentStep{
			AgentName: getAgentName(agent),
			Input:     input,
			Output:    ctx,
			Error:     err,
			Duration:  time.Since(start),
		}
		ctx = ctx.WithAgentStep(step)
		
		// If this agent requested preview, convert to context state
		if IsPreviewError(err) {
			ctx = ctx.WithDecision(DecisionPreview, err.Error())
			err = nil // Preview is not an error, it's a state transition
		}
	}
	
	return ctx, err
}

// ExecuteWithCallback runs pipeline with progress callback for UI updates
func ExecuteWithCallback(ctx Context, callback func(AgentStep), agents ...Agent) (Context, error) {
	var err error
	
	for _, agent := range agents {
		if err != nil {
			return ctx, err
		}
		
		if ctx.IsDenied() {
			return ctx, DenyWithRule("pipeline short-circuited due to prior denial", "system")
		}
		
		start := time.Now()
		input := ctx
		
		ctx, err = agent(ctx)
		
		step := AgentStep{
			AgentName: getAgentName(agent),
			Input:     input,
			Output:    ctx,
			Error:     err,
			Duration:  time.Since(start),
		}
		ctx = ctx.WithAgentStep(step)
		
		if callback != nil {
			callback(step)
		}
		
		if IsPreviewError(err) {
			ctx = ctx.WithDecision(DecisionPreview, err.Error())
			err = nil
		}
	}
	
	return ctx, err
}

// MustExecute runs pipeline and panics on error (use only in tests)
func MustExecute(ctx Context, agents ...Agent) Context {
	result, err := Execute(ctx, agents...)
	if err != nil {
		panic(err)
	}
	return result
}

// getAgentName extracts a readable name from agent function
// In production, you'd use reflection or explicit naming
func getAgentName(agent Agent) string {
	// Simplified - in real code use reflection or pass names explicitly
	return "agent"
}

// NamedAgent wraps an agent with a name for better logging
func NamedAgent(name string, agent Agent) Agent {
	return func(ctx Context) (Context, error) {
		result, err := agent(ctx)
		return result, err
	}
}

// ConditionalAgent only runs if condition is true
func ConditionalAgent(condition func(Context) bool, agent Agent) Agent {
	return func(ctx Context) (Context, error) {
		if !condition(ctx) {
			return ctx, nil // Skip this agent
		}
		return agent(ctx)
	}
}

// OnlyIfPreview runs agent only when preview is required
func OnlyIfPreview(agent Agent) Agent {
	return ConditionalAgent(func(ctx Context) bool {
		return ctx.RequiresPreview()
	}, agent)
}

// OnlyIfAllowed runs agent only when command is allowed
func OnlyIfAllowed(agent Agent) Agent {
	return ConditionalAgent(func(ctx Context) bool {
		return ctx.IsSafe()
	}, agent)
}
