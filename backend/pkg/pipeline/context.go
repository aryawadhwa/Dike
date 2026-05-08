// Package pipeline implements Railway-Oriented Programming for secure command execution.
// All state is immutable - agents return new contexts, never mutate existing ones.
package pipeline

import (
	"time"

	"github.com/aryawadhwa/dike/pkg/policy"
)

// Decision represents the security decision for a command
type Decision string

const (
	DecisionAllow   Decision = "ALLOW"
	DecisionPreview Decision = "PREVIEW"
	DecisionDeny    Decision = "DENY"
)

// FileChange represents a single filesystem modification
type FileChange struct {
	Path string `json:"path"`
	Type string `json:"type"` // "created", "modified", "deleted"
}

// SandboxResult contains the output from ghost sandbox execution
type SandboxResult struct {
	ExitCode   int
	Stdout     string
	Stderr     string
	FileChanges []FileChange
}

// Context is strictly immutable. Agents return new Contexts, never modify existing ones.
// This enables auditability - at any point we can serialize the full state.
type Context struct {
	// Input (set at creation, never changes)
	RawCommand string
	Policy     *policy.Policy
	SessionID  string
	Timestamp  time.Time

	// Parsed state (populated by parser agent)
	ParsedCommand string

	// Decision state (populated by gatekeeper)
	Decision   Decision
	DenyReason string

	// Sandbox results (populated if Decision == Preview)
	Sandbox SandboxResult

	// Audit trail (append-only)
	AgentTrail []AgentStep
}

// AgentStep records an agent's execution for auditability
type AgentStep struct {
	AgentName string
	Input     Context
	Output    Context
	Error     error
	Duration  time.Duration
}

// NewContext creates a fresh context for a command execution
func NewContext(command string, pol *policy.Policy) Context {
	return Context{
		RawCommand: command,
		Policy:     pol,
		Timestamp:  time.Now(),
		SessionID:  generateSessionID(),
		Decision:   DecisionAllow, // default - will be overwritten by gatekeeper
	}
}

// WithDecision returns a new Context with updated decision (immutable)
func (c Context) WithDecision(d Decision, reason string) Context {
	c.Decision = d
	c.DenyReason = reason
	return c
}

// WithParsedCommand returns a new Context with parsed command
func (c Context) WithParsedCommand(parsed string) Context {
	c.ParsedCommand = parsed
	return c
}

// WithSandboxResult returns a new Context with sandbox output
func (c Context) WithSandboxResult(result SandboxResult) Context {
	c.Sandbox = result
	return c
}

// WithAgentStep appends an audit step and returns new Context
func (c Context) WithAgentStep(step AgentStep) Context {
	c.AgentTrail = append(c.AgentTrail, step)
	return c
}

// IsSafe returns true if the command is allowed to run on host
func (c Context) IsSafe() bool {
	return c.Decision == DecisionAllow
}

// RequiresPreview returns true if command needs sandbox preview
func (c Context) RequiresPreview() bool {
	return c.Decision == DecisionPreview
}

// IsDenied returns true if command is blocked
func (c Context) IsDenied() bool {
	return c.Decision == DecisionDeny
}

// generateSessionID creates a unique session identifier
func generateSessionID() string {
	return time.Now().Format("20060102-150405-") + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
