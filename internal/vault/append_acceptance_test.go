package vault

import (
	"strings"
	"testing"
)

func TestAcceptance_AppendFile_ReadableViaRead(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = AppendFile(v, "S01.11.11", "notes.md", "APPENDED_MARKER")
	if err != nil {
		t.Fatalf("AppendFile: %v", err)
	}

	result, err := Read(v, "S01.11.11", "notes.md")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !strings.Contains(result.Content, "APPENDED_MARKER") {
		t.Error("appended content not found via Read")
	}
}

func TestAcceptance_AppendFile_OriginalPreserved(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// Read original
	orig, _ := Read(v, "S01.11.11", "notes.md")

	_, err = AppendFile(v, "S01.11.11", "notes.md", "Extra stuff")
	if err != nil {
		t.Fatalf("AppendFile: %v", err)
	}

	result, _ := Read(v, "S01.11.11", "notes.md")
	if !strings.HasPrefix(result.Content, orig.Content[:10]) {
		t.Error("original content should be preserved at start of file")
	}
}
