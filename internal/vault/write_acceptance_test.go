package vault

import (
	"slices"
	"strings"
	"testing"
)

func TestAcceptance_WriteFile_ReadableViaRead(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = WriteFile(v, "S01.11.11", "review.md", "# Review\n\nGreat show.")
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Re-parse to pick up new file
	v2, _ := ParseVault(root)
	result, err := Read(v2, "S01.11.11", "review.md")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !strings.Contains(result.Content, "Great show.") {
		t.Errorf("Read should return written content, got:\n%s", result.Content)
	}
}

func TestAcceptance_WriteFile_AppearsInFileListing(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = WriteFile(v, "S01.11.11", "checklist.md", "- [ ] Buy tickets")
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	v2, _ := ParseVault(root)
	result, err := Read(v2, "S01.11.11", "")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !slices.Contains(result.Files, "checklist.md") {
		t.Errorf("Files = %v, expected checklist.md", result.Files)
	}
}
