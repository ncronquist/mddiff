package e2e

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// DiffReport structure matching the JSON output.
type DiffReport struct {
	Items []struct {
		Type   string `json:"type"`
		Path   string `json:"path"`
		Reason string `json:"reason"`
	} `json:"items"`
	Summary struct {
		TotalMissing  int `json:"total_missing"`
		TotalModified int `json:"total_modified"`
	} `json:"summary"`
}

func TestEndToEnd(t *testing.T) {
	// 1. Build the binary
	binaryPath := filepath.Join(os.TempDir(), "mddiff-e2e")
	// We assume we are in the root or test/e2e directory.
	// We need to find the module root.
	// This test file will be in test/e2e/, so root is ../../
	wd, _ := os.Getwd()
	moduleRoot := filepath.Join(wd, "..", "..")

	// If running from root via `go test ./test/e2e/...`:
	if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
		moduleRoot = wd
	}

	// nolint:gosec // Test file, running 'go build' on known package
	buildCmd := exec.CommandContext(context.Background(), "go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = moduleRoot
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, string(out))
	}
	defer func() {
		_ = os.Remove(binaryPath)
	}()

	// 2. Setup Test Data
	tempDir, err := os.MkdirTemp("", "mddiff-e2e-data")
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

	// Case 1: MISSING (In Src, not Tgt)
	mustWriteFile(t, filepath.Join(srcDir, "missing.mkv"), "content")

	// Case 2: EXTRA (In Tgt, not Src)
	mustWriteFile(t, filepath.Join(tgtDir, "extra.mkv"), "content")

	// Case 3: MATCH (Same name, same size/ext)
	mustWriteFile(t, filepath.Join(srcDir, "match.mp4"), "content")
	mustWriteFile(t, filepath.Join(tgtDir, "match.mp4"), "content")

	// Case 4: MODIFIED (Extension change)
	mustWriteFile(t, filepath.Join(srcDir, "movie.mkv"), "content")
	mustWriteFile(t, filepath.Join(tgtDir, "movie.mp4"), "content") // Different ext

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

	// Check Summary
	if report.Summary.TotalMissing != 1 {
		t.Errorf("Expected 1 missing, got %d", report.Summary.TotalMissing)
	}
	if report.Summary.TotalModified != 1 {
		t.Errorf("Expected 1 modified, got %d", report.Summary.TotalModified)
	}

	// Check Items
	foundMissing := false
	foundModified := false
	foundExtra := false

	for _, item := range report.Items {
		switch item.Type {
		case "MISSING":
			if item.Path == "missing.mkv" {
				foundMissing = true
			}
		case "MODIFIED":
			if item.Path == "movie.mkv" && strings.Contains(item.Reason, "Extension changed") {
				foundModified = true
			}
		case "EXTRA":
			if item.Path == "extra.mkv" {
				foundExtra = true
			}
		}
	}

	if !foundMissing {
		t.Error("Did not find expected MISSING item: missing.mkv")
	}
	if !foundModified {
		t.Error("Did not find expected MODIFIED item: movie.mkv")
	}
	if !foundExtra {
		t.Error("Did not find expected EXTRA item: extra.mkv")
	}
}

func mustMkdir(t *testing.T, path string) {
	// G301: Expect directory permissions to be 0750 or less
	if err := os.MkdirAll(path, 0o750); err != nil {
		t.Fatal(err)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	// G306: Expect WriteFile permissions to be 0600 or less
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
