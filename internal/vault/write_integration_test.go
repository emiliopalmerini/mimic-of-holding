package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileIntegration_NewFile(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	path, err := WriteFile(v, "S01.11.11", "new-note.md", "# New Note\n\nSome content.")
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}
	if string(data) != "# New Note\n\nSome content." {
		t.Errorf("content mismatch: %q", string(data))
	}
}

func TestWriteFileIntegration_Overwrite(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// notes.md already exists in S01.11.11
	_, err = WriteFile(v, "S01.11.11", "notes.md", "Updated content.")
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	path := filepath.Join(root, "S01 Me", "S01.10-19 Lifestyle", "S01.11 Entertainment", "S01.11.11 Theatre, 2025 Season", "notes.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading overwritten file: %v", err)
	}
	if string(data) != "Updated content." {
		t.Errorf("content mismatch: %q", string(data))
	}
}

func TestWriteFileIntegration_EmptyContent(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	path, err := WriteFile(v, "S01.11.11", "empty.md", "")
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("expected empty file, got %d bytes", info.Size())
	}
}
