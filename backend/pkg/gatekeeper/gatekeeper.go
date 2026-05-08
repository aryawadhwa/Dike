package gatekeeper

import (
	"strings"

	"github.com/aryawadhwa/dike/pkg/policy"
	"mvdan.cc/sh/v3/syntax"
)

// Evaluate parses the command and checks it against the loaded policy.
func Evaluate(cmdString string, pol *policy.Policy) policy.Decision {
	if pol == nil || len(pol.Rules) == 0 {
		return policy.DecisionAllow // Default to allow if no policy
	}

	parser := syntax.NewParser()
	f, err := parser.Parse(strings.NewReader(cmdString), "")
	if err != nil {
		// Deny malformed commands to prevent bypasses
		return policy.DecisionDeny
	}

	decision := policy.DecisionAllow

	syntax.Walk(f, func(node syntax.Node) bool {
		call, ok := node.(*syntax.CallExpr)
		if !ok {
			return true // continue walking
		}

		if len(call.Args) == 0 {
			return true
		}

		// Extract the command name (first argument)
		cmdName := wordToString(call.Args[0])

		// Extract the rest as args
		var args []string
		for i := 1; i < len(call.Args); i++ {
			args = append(args, wordToString(call.Args[i]))
		}

		// Handle common wrappers like sudo, xargs
		var commandWrappers = map[string]bool{
			"sudo":  true,
			"xargs": true,
			"time":  true,
			"watch": true,
			"env":   true,
		}

		// Basic unwrap and heuristic check
		isSudo := cmdName == "sudo"
		if commandWrappers[cmdName] && len(args) > 0 {
			// Find the first arg that doesn't start with '-' (naive approach for Phase 1)
			for i, arg := range args {
				if !strings.HasPrefix(arg, "-") {
					cmdName = arg
					args = args[i+1:]
					break
				}
			}
		}

		// Heuristic: Sudo commands should always be previewed if not denied
		if isSudo && decision != policy.DecisionDeny {
			decision = policy.DecisionPreview
		}

		// Check against rules
		for _, rule := range pol.Rules {
			if ruleMatches(rule, cmdName, args) {
				// We escalate the decision. DENY overrides PREVIEW overrides ALLOW
				if rule.Action == policy.DecisionDeny {
					decision = policy.DecisionDeny
					return false // No need to check further, it's denied
				}
				if rule.Action == policy.DecisionPreview && decision != policy.DecisionDeny {
					decision = policy.DecisionPreview
				}
			}
		}

		return true
	})

	return decision
}

// wordToString extracts a static string representation of a shell word.
func wordToString(word *syntax.Word) string {
	var sb strings.Builder
	for _, part := range word.Parts {
		switch p := part.(type) {
		case *syntax.Lit:
			sb.WriteString(p.Value)
		case *syntax.SglQuoted:
			sb.WriteString(p.Value)
		case *syntax.DblQuoted:
			for _, dp := range p.Parts {
				if dplit, ok := dp.(*syntax.Lit); ok {
					sb.WriteString(dplit.Value)
				}
			}
		}
	}
	return sb.String()
}

// ruleMatches checks if a parsed command and its arguments match a given policy rule.
func ruleMatches(rule policy.Rule, cmd string, args []string) bool {
	cmdMatch := false
	if len(rule.Match.Commands) == 0 {
		cmdMatch = true
	} else {
		for _, c := range rule.Match.Commands {
			if c == cmd {
				cmdMatch = true
				break
			}
		}
	}

	if !cmdMatch {
		return false
	}

	// If no args specified in rule, and cmd matched, it's a match
	if len(rule.Match.Args) == 0 {
		return true
	}

	// If args are specified in rule, the command MUST have at least one of those args
	for _, ruleArg := range rule.Match.Args {
		for _, arg := range args {
			// For Phase 1, we do simple exact string or prefix matching based on args
			if arg == ruleArg || strings.HasPrefix(arg, ruleArg) {
				return true
			}
		}
	}

	return false
}
