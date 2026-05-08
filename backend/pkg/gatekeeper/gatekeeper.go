package gatekeeper

import (
	"strings"

	"github.com/aryawadhwa/dike/pkg/policy"
	"mvdan.cc/sh/v3/syntax"
)

// Evaluate parses the command and checks it against the Capability model.
func Evaluate(cmdString string, pol *policy.Policy) (policy.Decision, policy.Capability) {
	parser := syntax.NewParser()
	f, err := parser.Parse(strings.NewReader(cmdString), "")
	if err != nil {
		return policy.DecisionDeny, "" // Deny malformed commands
	}

	cap, found := EvaluateCapabilities(f)
	
	// 1. Check for explicit command-level denials first (e.g. rm -rf /)
	if pol != nil {
		for _, rule := range pol.Deny {
			for _, badCmd := range rule.Commands {
				if strings.Contains(cmdString, badCmd) {
					return policy.DecisionDeny, cap
				}
			}
		}
	}

	if !found {
		return policy.DecisionAllow, "" // Default to allow for non-destructive capabilities
	}

	// 2. Check policy for capability-specific actions
	if pol != nil {
		for _, rule := range pol.Deny {
			if policy.Capability(rule.Capability) == cap {
				return policy.DecisionDeny, cap
			}
		}
		for _, rule := range pol.Allow {
			if policy.Capability(rule.Capability) == cap {
				return policy.DecisionAllow, cap
			}
		}
		// If capability found but not explicitly allowed/denied, use default_action
		if pol.DefaultAction != "" {
			return pol.DefaultAction, cap
		}

	}

	// Fallback for default setup
	switch cap {
	case policy.CapMassDelete, policy.CapSystemModify, policy.CapExecArbitrary:
		return policy.DecisionPreview, cap
	default:
		return policy.DecisionAllow, cap
	}
}

// EvaluateCapabilities traverses the AST to find the semantic intent
func EvaluateCapabilities(node syntax.Node) (policy.Capability, bool) {
	var cmd *syntax.CallExpr
	syntax.Walk(node, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.CallExpr); ok {
			if len(c.Args) > 0 {
				cmd = c
				return false
			}
		}
		return true
	})

	if cmd == nil || len(cmd.Args) == 0 {
		return "", false
	}

	baseCmd := wordToString(cmd.Args[0])
	subCmd := ""
	if len(cmd.Args) > 1 {
		secondArg := wordToString(cmd.Args[1])
		if !strings.HasPrefix(secondArg, "-") {
			subCmd = secondArg
		}
	}

	flags := extractFlags(cmd.Args)

	for cap, signatures := range DestructiveCapabilities {
		for _, sig := range signatures {
			if matchSignature(baseCmd, subCmd, flags, sig) {
				return cap, true
			}
		}
	}

	return "", false
}

func matchSignature(baseCmd, subCmd string, presentFlags map[string]bool, sig Signature) bool {
	if baseCmd != sig.Command {
		return false
	}
	if sig.Subcommand != "" && subCmd != sig.Subcommand {
		return false
	}
	
	if len(sig.DangerousFlags) == 0 && sig.Subcommand != "" {
		return true
	}

	for _, df := range sig.DangerousFlags {
		if presentFlags[df] {
			return true
		}
	}

	return false
}

func extractFlags(args []*syntax.Word) map[string]bool {
	flags := make(map[string]bool)
	for _, arg := range args {
		str := wordToString(arg)
		if !strings.HasPrefix(str, "-") {
			continue
		}
		
		cleaned := strings.TrimLeft(str, "-")
		if strings.HasPrefix(str, "--") {
			flags[cleaned] = true
		} else {
			for _, r := range cleaned {
				flags[string(r)] = true
			}
		}
	}
	return flags
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
