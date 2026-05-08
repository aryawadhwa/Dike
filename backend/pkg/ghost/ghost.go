package ghost

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/aryawadhwa/dike/pkg/pipeline"
)

const containerName = "pulse-ghost"

// InitSandbox ensures any old ghost container is removed and starts a fresh one.
func InitSandbox(cwd string) error {
	// Cleanup any stale instances
	_ = Teardown()

	// Run a long-lived hardened alpine container with Zero-Network access and strict quotas
	// Default to no network. NetworkIso strategy will re-enable if allowed.
	cmd := exec.Command("docker", "run", "-d",
		"--name", containerName,
		"--network", "none",
		"--memory", "128m",
		"--cpus", "0.5",
		"alpine", "tail", "-f", "/dev/null",
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start ghost container: %w, stderr: %s", err, stderr.String())
	}

	// Ensure fakeroot is installed for StrategyFakeRoot
	_ = exec.Command("docker", "exec", containerName, "apk", "add", "fakeroot").Run()

	return nil
}

// ExecPreview copies the current directory into the ghost sandbox and executes a command.
// It uses the provided SandboxStrategy to configure the environment.
func ExecPreview(command string, strategy pipeline.SandboxStrategy) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get cwd: %w", err)
	}

	// Prepare an empty /workspace in the container
	_ = exec.Command("docker", "exec", containerName, "rm", "-rf", "/workspace").Run()
	_ = exec.Command("docker", "exec", containerName, "mkdir", "-p", "/workspace").Run()

	// Copy the current host directory into the container's /workspace
	cpCmd := exec.Command("docker", "cp", cwd+"/.", containerName+":/workspace/")
	if err := cpCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to copy host directory to ghost sandbox: %w", err)
	}

	// Apply Strategy-specific modifications
	actualCommand := command
	if strategy == pipeline.StrategyFakeRoot {
		actualCommand = "fakeroot " + command
	}

	// We run the command via sh -c to handle pipelines and basic shell semantics
	cmd := exec.Command("docker", "exec", "-w", "/workspace", containerName, "sh", "-c", actualCommand)

	// Combine stdout and stderr for the preview
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err = cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return fmt.Sprintf("[Exit Code %d]\n%s", exitError.ExitCode(), out.String()), nil
		}
		return "", fmt.Errorf("failed to execute in ghost: %w\n%s", err, out.String())
	}

	return out.String(), nil
}


// Teardown forcefully removes the ghost sandbox.
func Teardown() error {
	cmd := exec.Command("docker", "rm", "-f", containerName)
	// Ignore errors since it might not exist
	_ = cmd.Run()
	return nil
}
