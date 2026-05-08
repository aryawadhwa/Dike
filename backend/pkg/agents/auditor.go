// Package agents provides pure function agents
package agents

import (
	"github.com/aryawadhwa/dike/pkg/audit"
	"github.com/aryawadhwa/dike/pkg/pipeline"
)

// Auditor logs the final decision to SQLite.
// This is a side-effect agent, but still pure - it doesn't mutate context.
// The audit log is append-only external state.
func Auditor() pipeline.Agent {
	return func(ctx pipeline.Context) (pipeline.Context, error) {
		// Determine final decision string
		decision := "ALLOWED"
		if ctx.IsDenied() {
			decision = "REJECTED"
		} else if ctx.RequiresPreview() {
			decision = "PREVIEWED"
		}

		// Get risk level
		riskLevel := "LOW"
		if ctx.IsDenied() {
			riskLevel = "CRITICAL"
		} else if ctx.RequiresPreview() {
			riskLevel = "HIGH"
		}

		// Build preview summary from sandbox results
		previewSummary := "No changes"
		if len(ctx.Sandbox.FileChanges) > 0 {
			previewSummary = formatFileChanges(ctx.Sandbox.FileChanges)
		}

		// Log to SQLite (side effect - append only)
		_ = audit.LogDecision(
			ctx.RawCommand,
			riskLevel,
			decision,
			previewSummary,
		)

		// Return unchanged context (auditor doesn't modify state)
		return ctx, nil
	}
}

// formatFileChanges creates readable summary
func formatFileChanges(changes []pipeline.FileChange) string {
	if len(changes) == 0 {
		return "No file changes"
	}

	var result string
	for _, c := range changes {
		result += c.Type + ": " + c.Path + "\n"
	}
	return result
}

// AuditorWithTrail logs full agent execution trail
func AuditorWithTrail() pipeline.Agent {
	return func(ctx pipeline.Context) (pipeline.Context, error) {
		// Log with full audit trail for debugging
		decision := "ALLOWED"
		if ctx.IsDenied() {
			decision = "REJECTED"
		} else if ctx.RequiresPreview() {
			decision = "PREVIEWED"
		}

		riskLevel := "LOW"
		if ctx.IsDenied() {
			riskLevel = "CRITICAL"
		} else if ctx.RequiresPreview() {
			riskLevel = "HIGH"
		}

		previewSummary := formatFileChanges(ctx.Sandbox.FileChanges)

		_ = audit.LogDecision(
			ctx.RawCommand,
			riskLevel,
			decision,
			previewSummary,
		)

		return ctx, nil
	}
}
