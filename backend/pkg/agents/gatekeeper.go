// Package agents provides pure function agents for the security pipeline
package agents

import (
	"github.com/aryawadhwa/dike/pkg/gatekeeper"
	"github.com/aryawadhwa/dike/pkg/pipeline"
	"github.com/aryawadhwa/dike/pkg/policy"
)

// Gatekeeper evaluates commands against policy and returns decisions.
// This is a pure function - it takes immutable Context, returns new Context.
func Gatekeeper() pipeline.Agent {
	return func(ctx pipeline.Context) (pipeline.Context, error) {
		if ctx.Policy == nil {
			// No policy loaded, allow by default but log warning
			return ctx.WithDecision(pipeline.DecisionAllow, "no policy loaded"), nil
		}

		decision := gatekeeper.Evaluate(ctx.RawCommand, ctx.Policy)

		switch decision {
		case policy.DecisionDeny:
			return ctx.WithDecision(pipeline.DecisionDeny, "matches blocklist rule"),
				pipeline.DenyWithRule("command matches deny policy", "blocklist")

		case policy.DecisionPreview:
			// Preview is not an error - it's a state transition
			// Return special error that pipeline converts to context state
			return ctx.WithDecision(pipeline.DecisionPreview, "requires sandbox preview"),
				&pipeline.PreviewError{Command: ctx.RawCommand, Reason: "risky command pattern detected"}

		case policy.DecisionAllow:
			return ctx.WithDecision(pipeline.DecisionAllow, "safe command"), nil

		default:
			return ctx.WithDecision(pipeline.DecisionAllow, "unknown decision type"), nil
		}
	}
}

// GatekeeperWithCustomRules allows runtime rule injection for testing
func GatekeeperWithCustomRules(rules []policy.Rule) pipeline.Agent {
	return func(ctx pipeline.Context) (pipeline.Context, error) {
		customPolicy := &policy.Policy{Rules: rules}
		decision := gatekeeper.Evaluate(ctx.RawCommand, customPolicy)

		switch decision {
		case policy.DecisionDeny:
			return ctx.WithDecision(pipeline.DecisionDeny, "custom rule match"),
				pipeline.DenyWithRule("custom rule triggered", "custom")
		case policy.DecisionPreview:
			return ctx.WithDecision(pipeline.DecisionPreview, "custom preview rule"),
				&pipeline.PreviewError{Command: ctx.RawCommand, Reason: "custom preview rule"}
		default:
			return ctx.WithDecision(pipeline.DecisionAllow, "passed custom rules"), nil
		}
	}
}
