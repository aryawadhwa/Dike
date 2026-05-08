// Package agents provides pure function agents
package agents

import (
	"github.com/aryawadhwa/dike/pkg/diff"
	"github.com/aryawadhwa/dike/pkg/ghost"
	"github.com/aryawadhwa/dike/pkg/pipeline"
)

// Ghost executes commands in Docker sandbox.
// Only runs if context indicates preview is required.
// Returns new context with sandbox results (immutable).
func Ghost() pipeline.Agent {
	return func(ctx pipeline.Context) (pipeline.Context, error) {
		// Skip if not in preview mode
		if !ctx.RequiresPreview() {
			return ctx, nil
		}

		// Initialize sandbox if needed
		if err := ghost.InitSandbox("."); err != nil {
			return ctx, err
		}

		// Execute in sandbox with specific strategy
		stdout, err := ghost.ExecPreview(ctx.RawCommand, ctx.SandboxStrategy)
		if err != nil {
			return ctx, err
		}


		// Compute filesystem diff
		diffSummary, err := diff.ComputeDiff("pulse-ghost")
		if err != nil {
			return ctx, err
		}

		// Build sandbox result
		result := pipeline.SandboxResult{
			ExitCode: 0, // Could parse from stdout if needed
			Stdout:   stdout,
		}

		// Convert diff.Summary to pipeline.FileChange
		for _, added := range diffSummary.Added {
			result.FileChanges = append(result.FileChanges, pipeline.FileChange{
				Path: added,
				Type: "created",
			})
		}
		for _, deleted := range diffSummary.Deleted {
			result.FileChanges = append(result.FileChanges, pipeline.FileChange{
				Path: deleted,
				Type: "deleted",
			})
		}

		// Return NEW context with sandbox results (immutable)
		return ctx.WithSandboxResult(result), nil
	}
}

// GhostWithDir executes in specific directory
func GhostWithDir(dir string) pipeline.Agent {
	return func(ctx pipeline.Context) (pipeline.Context, error) {
		if !ctx.RequiresPreview() {
			return ctx, nil
		}

		if err := ghost.InitSandbox(dir); err != nil {
			return ctx, err
		}

		stdout, err := ghost.ExecPreview(ctx.RawCommand, ctx.SandboxStrategy)
		if err != nil {
			return ctx, err
		}


		result := pipeline.SandboxResult{
			ExitCode: 0,
			Stdout:   stdout,
		}

		return ctx.WithSandboxResult(result), nil
	}
}
