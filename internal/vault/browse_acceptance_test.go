package vault

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAcceptance_Browse_FullOutput(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	got, err := Browse(v, "")
	if err != nil {
		t.Fatalf("Browse: %v", err)
	}

	// Acceptance criteria 1: All scope names present
	for _, name := range []string{"Me", "Due Draghi"} {
		if !strings.Contains(got, name) {
			t.Errorf("output missing scope name %q", name)
		}
	}

	// Acceptance criteria 2: No non-JD entries
	for _, bad := range []string{".obsidian", "Attachments", "README.md"} {
		if strings.Contains(got, bad) {
			t.Errorf("non-JD entry %q should not appear in output", bad)
		}
	}

	// Acceptance criteria 3: Consistent indentation (2 spaces per level)
	lines := strings.Split(got, "\n")
	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		indent := len(line) - len(trimmed)
		if indent%2 != 0 {
			t.Errorf("odd indentation (%d spaces) on line: %q", indent, line)
		}
	}
}

func TestAcceptance_Browse_FilteredOutputExcludesOthers(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	got, err := Browse(v, "S02")
	if err != nil {
		t.Fatalf("Browse: %v", err)
	}

	// Filtered output should only contain S02 content
	lines := strings.Split(got, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Every line with a JD prefix should be S02
		if strings.HasPrefix(trimmed, "S") && !strings.HasPrefix(trimmed, "S02") {
			t.Errorf("filtered output contains non-S02 line: %q", line)
		}
	}
}
