// Package agents provides pure function agents
package agents

import (
	"github.com/aryawadhwa/dike/pkg/advisor"
	"github.com/aryawadhwa/dike/pkg/pipeline"
)

// Advisor generates LLM explanations for risky commands.
// Pure function - takes context, returns new context with explanation.
func Advisor() pipeline.Agent {
	return func(ctx pipeline.Context) (pipeline.Context, error) {
		// Generate explanation based on command and sandbox results
		_ = advisor.ExplainCommand(ctx.RawCommand)

		// For now, we don't store explanation in context - could add if needed
		// This demonstrates the pattern of enriching context with metadata

		return ctx, nil
	}
}

// AdvisorWithSandbox generates explanation including sandbox results
func AdvisorWithSandbox() pipeline.Agent {
	return func(ctx pipeline.Context) (pipeline.Context, error) {
		if !ctx.RequiresPreview() {
			return ctx, nil
		}

		// Enhanced prompt with sandbox results
		explanation := advisor.ExplainCommand(ctx.RawCommand)

		_ = explanation // Could store in context if we add Explanation field

		return ctx, nil
	}
}
