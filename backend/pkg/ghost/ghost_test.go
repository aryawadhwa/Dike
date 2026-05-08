package ghost

import (
	"os/exec"
	"strings"
	"testing"
)

// Note: These tests require Docker to be installed and running on the host machine.
// In a real CI environment, you would mock exec.Command, but for this project,
// integration tests verifying actual docker commands are more valuable.

func TestGhostLifecycle(t *testing.T) {
	requireDocker(t)
	// 1. Init
	err := InitSandbox("/")
	if err != nil {
		t.Fatalf("InitSandbox failed: %v", err)
	}

	// Ensure teardown runs even if the test panics
	defer Teardown()

	// 2. Exec simple command
	out, err := ExecPreview("echo 'hello from ghost'")
	if err != nil {
		t.Fatalf("ExecPreview failed: %v", err)
	}

	if !strings.Contains(out, "hello from ghost") {
		t.Errorf("Expected output to contain 'hello from ghost', got: %s", out)
	}

	// 3. Exec failing command
	out, err = ExecPreview("ls /path-that-does-not-exist")
	// It should NOT return a Go error for a non-zero exit code of the command itself,
	// but rather capture the exit code in the output string or handle it gracefully.
	// Based on our implementation, it returns a formatted string with the exit code.
	if err != nil {
		t.Fatalf("ExecPreview should not fail the Go wrapper on non-zero exit, but got err: %v", err)
	}

	if !strings.Contains(out, "Exit Code") {
		t.Errorf("Expected output to contain 'Exit Code', got: %s", out)
	}

	// 4. Teardown
	err = Teardown()
	if err != nil {
		t.Fatalf("Teardown failed: %v", err)
	}
}

func requireDocker(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("skipping ghost integration test: docker CLI not found in PATH")
	}

	if out, err := exec.Command("docker", "info").CombinedOutput(); err != nil {
		t.Skipf("skipping ghost integration test: docker daemon unavailable: %v (%s)", err, strings.TrimSpace(string(out)))
	}
}
