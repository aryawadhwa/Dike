package ghost

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

const containerName = "pulse-ghost"

// InitSandbox ensures any old ghost container is removed and starts a fresh one.
func InitSandbox(cwd string) error {
	// Cleanup any stale instances
	_ = Teardown()

	// Run a long-lived hardened alpine container with Zero-Network access and strict quotas
	cmd := exec.Command("docker", "run", "-d",
		"--name", containerName,
		"--network", "none", // Air-gap: No internet access in sandbox
		"--memory", "128m",   // RAM quota: Prevent memory leaks from crashing host
		"--cpus", "0.5",      // CPU quota: Prevent fork-bombs
		"alpine", "tail", "-f", "/dev/null",
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start ghost container: %w, stderr: %s", err, stderr.String())
	}

	return nil
}

// ExecPreview copies the current directory into the ghost sandbox and executes a command.
// It returns the combined output.
func ExecPreview(command string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get cwd: %w", err)
	}

	// Prepare an empty /workspace in the container
	_ = exec.Command("docker", "exec", containerName, "rm", "-rf", "/workspace").Run()
	_ = exec.Command("docker", "exec", containerName, "mkdir", "-p", "/workspace").Run()

	// Copy the current host directory into the container's /workspace
	// Note: docker cp <src>/. <dest> copies contents of src into dest.
	cpCmd := exec.Command("docker", "cp", cwd+"/.", containerName+":/workspace/")
	if err := cpCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to copy host directory to ghost sandbox: %w", err)
	}

	// We run the command via sh -c to handle pipelines and basic shell semantics
	cmd := exec.Command("docker", "exec", "-w", "/workspace", containerName, "sh", "-c", command)

	// Combine stdout and stderr for the preview
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err = cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// The command ran but failed
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
