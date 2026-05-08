// Package agents provides pure function agents for the security pipeline
package agents

import (
	"github.com/aryawadhwa/dike/pkg/gatekeeper"
	"github.com/aryawadhwa/dike/pkg/ghost"
	"github.com/aryawadhwa/dike/pkg/pipeline"
	"github.com/aryawadhwa/dike/pkg/policy"
)

// Gatekeeper evaluates commands against policy and returns decisions.
func Gatekeeper() pipeline.Agent {
	return func(ctx pipeline.Context) (pipeline.Context, error) {
		if ctx.Policy == nil {
			return ctx.WithDecision(pipeline.DecisionAllow, "no policy loaded"), nil
		}

		decision, cap := gatekeeper.Evaluate(ctx.RawCommand, ctx.Policy)
		
		// Populate capability and strategy for downstream agents (Ghost)
		newCtx := ctx.WithCapability(cap)
		if decision == policy.DecisionPreview {
			newCtx = newCtx.WithSandboxStrategy(ghost.DeriveStrategy(cap))
		}

		switch decision {
		case policy.DecisionDeny:
			return newCtx.WithDecision(pipeline.DecisionDeny, "matches blocklist rule"),
				pipeline.DenyWithRule("command matches deny policy", string(cap))

		case policy.DecisionPreview:
			return newCtx.WithDecision(pipeline.DecisionPreview, "requires sandbox preview"),
				&pipeline.PreviewError{Command: ctx.RawCommand, Reason: "risky command pattern detected"}

		case policy.DecisionAllow:
			return newCtx.WithDecision(pipeline.DecisionAllow, "safe command"), nil

		default:
			return newCtx.WithDecision(pipeline.DecisionAllow, "unknown decision type"), nil
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
