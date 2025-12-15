package e2e

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
)

func TestDirectoryHiding(t *testing.T) {
	// 1. Build the binary
	binaryPath := filepath.Join(os.TempDir(), "mddiff-e2e-dir")
	wd, _ := os.Getwd()
	moduleRoot := filepath.Join(wd, "..", "..")
	if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
		moduleRoot = wd
	}

	// nolint:gosec // Test file, creating binary in temp dir
	buildCmd := exec.CommandContext(context.Background(), "go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = moduleRoot
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, string(out))
	}
	defer func() {
		_ = os.Remove(binaryPath)
	}()

	// 2. Setup Test Data
	tempDir, err := os.MkdirTemp("", "mddiff-dir-test-data")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	srcDir := filepath.Join(tempDir, "src")
	tgtDir := filepath.Join(tempDir, "tgt")

	mustMkdir(t, srcDir)
	mustMkdir(t, tgtDir)

	// Case 1: MISSING Empty Directory
	mustMkdir(t, filepath.Join(srcDir, "missing_empty"))

	// Case 2: MISSING Non-Empty Directory
	mustMkdir(t, filepath.Join(srcDir, "missing_non_empty"))
	mustWriteFile(t, filepath.Join(srcDir, "missing_non_empty", "file.txt"), "content")

	// Case 3: EXTRA Empty Directory
	mustMkdir(t, filepath.Join(tgtDir, "extra_empty"))

	// Case 4: EXTRA Non-Empty Directory
	mustMkdir(t, filepath.Join(tgtDir, "extra_non_empty"))
	mustWriteFile(t, filepath.Join(tgtDir, "extra_non_empty", "file.txt"), "content")

	// 3. Run Binary with JSON output
	// nolint:gosec // Test file, running binary on test data
	cmd := exec.CommandContext(context.Background(), binaryPath, srcDir, tgtDir, "--format", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, string(output))
	}

	// 4. Verify Output
	var report DiffReport
	if err := json.Unmarshal(output, &report); err != nil {
		t.Fatalf("Failed to decode JSON: %v\nOutput: %s", err, string(output))
	}

	var missingPaths []string
	var extraPaths []string

	for _, item := range report.Items {
		switch item.Type {
		case "MISSING":
			missingPaths = append(missingPaths, item.Path)
		case "EXTRA":
			extraPaths = append(extraPaths, item.Path)
		}
	}

	// Verify MISSING
	if !contains(missingPaths, "missing_empty") {
		t.Error("Expected 'missing_empty' to be reported")
	}
	if contains(missingPaths, "missing_non_empty") {
		t.Error("Expected 'missing_non_empty' (directory) NOT to be reported")
	}
	if !contains(missingPaths, "missing_non_empty/file.txt") {
		t.Error("Expected 'missing_non_empty/file.txt' to be reported")
	}

	// Verify EXTRA
	if !contains(extraPaths, "extra_empty") {
		t.Error("Expected 'extra_empty' to be reported")
	}
	if contains(extraPaths, "extra_non_empty") {
		t.Error("Expected 'extra_non_empty' (directory) NOT to be reported")
	}
	if !contains(extraPaths, "extra_non_empty/file.txt") {
		t.Error("Expected 'extra_non_empty/file.txt' to be reported")
	}
}

func contains(slice []string, val string) bool {
	return slices.Contains(slice, val)
}
