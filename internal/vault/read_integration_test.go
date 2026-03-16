package vault

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestReadIntegration_WithJDex(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Read(v, "S01.11.11")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if result.Ref != "S01.11.11" {
		t.Errorf("Ref = %q, want S01.11.11", result.Ref)
	}
	if result.Name != "Theatre, 2025 Season" {
		t.Errorf("Name = %q, want Theatre, 2025 Season", result.Name)
	}
	if result.JDex == "" {
		t.Error("JDex should not be empty")
	}
	// Extra file notes.md should appear in Files
	if !slices.Contains(result.Files, "notes.md") {
		t.Errorf("Files = %v, expected notes.md", result.Files)
	}
}

func TestReadIntegration_NoJDex(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Read(v, "S01.11.01")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if result.JDex != "" {
		t.Errorf("JDex should be empty for ID without JDex file, got %q", result.JDex)
	}
}

func TestReadIntegration_EmptyFolder(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// S01.10.01 Inbox for S01.10-19 is an empty folder
	result, err := Read(v, "S01.10.01")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if result.JDex != "" {
		t.Errorf("JDex should be empty, got %q", result.JDex)
	}
	if len(result.Files) != 0 {
		t.Errorf("Files should be empty, got %v", result.Files)
	}
}
