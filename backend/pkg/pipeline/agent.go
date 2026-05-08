// Package pipeline provides pure function agents for secure command processing
package pipeline

import (
	"fmt"
)

// Agent is a pure function that transforms Context.
// It takes an immutable Context and returns a new Context.
// If it returns an error, the pipeline short-circuits (railway siding).
type Agent func(Context) (Context, error)

// DenyError represents a security denial - pipeline halts
type DenyError struct {
	Reason string
	Rule   string
}

func (e *DenyError) Error() string {
	return fmt.Sprintf("DENIED: %s (rule: %s)", e.Reason, e.Rule)
}

// IsDenyError checks if an error is a security denial
func IsDenyError(err error) bool {
	_, ok := err.(*DenyError)
	return ok
}

// Deny creates a denial error - use this to block commands
func Deny(reason string) error {
	return &DenyError{Reason: reason}
}

// DenyWithRule creates a denial error with specific rule name
func DenyWithRule(reason, rule string) error {
	return &DenyError{Reason: reason, Rule: rule}
}

// PreviewError represents a request for sandbox preview (not fatal)
type PreviewError struct {
	Command string
	Reason  string
}

func (e *PreviewError) Error() string {
	return fmt.Sprintf("PREVIEW_REQUIRED: %s", e.Reason)
}

// IsPreviewError checks if error requests preview
func IsPreviewError(err error) bool {
	_, ok := err.(*PreviewError)
	return ok
}
