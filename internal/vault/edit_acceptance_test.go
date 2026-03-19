package vault

import (
	"strings"
	"testing"
)

func TestAcceptance_EditFile_ReadableViaRead(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = EditFile(v, "S01.11.11", "notes.md", "extra notes", "detailed notes")
	if err != nil {
		t.Fatalf("EditFile: %v", err)
	}

	// Re-parse to pick up changes
	v2, _ := ParseVault(root)
	result, err := Read(v2, "S01.11.11", "notes.md")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !strings.Contains(result.Content, "detailed notes") {
		t.Errorf("Read should return edited content, got:\n%s", result.Content)
	}
	if strings.Contains(result.Content, "extra notes") {
		t.Errorf("Read should not contain old text, got:\n%s", result.Content)
	}
}
