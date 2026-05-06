package diff

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Summary represents the differences found.
type Summary struct {
	Added   []string
	Deleted []string
	// Modified is omitted for this simple hackathon version
}

// ComputeDiff compares the host's current directory with the container's /workspace.
// For simplicity, it just compares the presence of files.
func ComputeDiff(containerName string) (*Summary, error) {
	// Get host files
	hostFiles := make(map[string]bool)
	hostCmd := exec.Command("find", ".", "-type", "f", "-not", "-path", "*/.git/*")
	var hostOut bytes.Buffer
	hostCmd.Stdout = &hostOut
	if err := hostCmd.Run(); err == nil {
		for _, f := range strings.Split(hostOut.String(), "\n") {
			f = strings.TrimSpace(f)
			if f != "" {
				// Normalize path (remove leading ./)
				f = strings.TrimPrefix(f, "./")
				hostFiles[f] = true
			}
		}
	}

	// Get container files
	containerFiles := make(map[string]bool)
	contCmd := exec.Command("docker", "exec", containerName, "find", "/workspace", "-type", "f", "-not", "-path", "*/.git/*")
	var contOut bytes.Buffer
	contCmd.Stdout = &contOut
	if err := contCmd.Run(); err == nil {
		for _, f := range strings.Split(contOut.String(), "\n") {
			f = strings.TrimSpace(f)
			if f != "" {
				// Normalize path (remove leading /workspace/)
				f = strings.TrimPrefix(f, "/workspace/")
				containerFiles[f] = true
			}
		}
	}

	summary := &Summary{}

	// Check for deleted files (exist in host, not in container)
	for hf := range hostFiles {
		if !containerFiles[hf] {
			summary.Deleted = append(summary.Deleted, hf)
		}
	}

	// Check for added files (exist in container, not in host)
	for cf := range containerFiles {
		if !hostFiles[cf] {
			summary.Added = append(summary.Added, cf)
		}
	}

	return summary, nil
}

// String provides a clean textual representation of the diff.
func (s *Summary) String() string {
	if len(s.Added) == 0 && len(s.Deleted) == 0 {
		return "No file changes detected."
	}

	var sb strings.Builder
	if len(s.Added) > 0 {
		sb.WriteString("Files created:\n")
		for _, f := range s.Added {
			sb.WriteString(fmt.Sprintf("  + %s\n", f))
		}
	}
	if len(s.Deleted) > 0 {
		sb.WriteString("Files deleted:\n")
		for _, f := range s.Deleted {
			sb.WriteString(fmt.Sprintf("  - %s\n", f))
		}
	}
	return strings.TrimSpace(sb.String())
}
