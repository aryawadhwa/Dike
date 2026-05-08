package ghost

import (
	"github.com/aryawadhwa/dike/pkg/pipeline"
	"github.com/aryawadhwa/dike/pkg/policy"
)

// DeriveStrategy maps security capabilities to the optimal sandbox execution mode
func DeriveStrategy(cap policy.Capability) pipeline.SandboxStrategy {
	switch cap {
	case policy.CapMassDelete:
		// Snapshot strategy uses OverlayFS to capture deletions and modifications
		return pipeline.StrategySnapshot
		
	case policy.CapSystemModify:
		// FakeRoot strategy handles administrative commands by simulating root privileges
		return pipeline.StrategyFakeRoot
		
	case policy.CapExecArbitrary:
		// NetworkIso strategy allows outbound traffic for downloads but restricts exfiltration
		return pipeline.StrategyNetworkIso
		
	default:
		// Fallback for general preview
		return pipeline.StrategySnapshot
	}
}
