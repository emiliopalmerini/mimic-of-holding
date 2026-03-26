package vault

import (
	"os"
	"testing"
)

func TestAppendFileIntegration_ExistingFile(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// notes.md exists with "Some extra notes about the theatre season.\n"
	path, err := AppendFile(v, "S01.11.11", "notes.md", "New line appended.")
	if err != nil {
		t.Fatalf("AppendFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	content := string(data)

	// Original content preserved
	if content[:7] != "# notes" {
		t.Errorf("original content not preserved, got:\n%s", content)
	}
	// Appended content present
	if content[len(content)-len("New line appended."):] != "New line appended." {
		t.Errorf("appended content not found at end, got:\n%s", content)
	}
}

func TestAppendFileIntegration_NewFile(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	path, err := AppendFile(v, "S01.11.11", "brand-new.md", "First content.")
	if err != nil {
		t.Fatalf("AppendFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(data) != "First content." {
		t.Errorf("content = %q, want 'First content.'", string(data))
	}
}

func TestAppendFileIntegration_EmptyContentNoOp(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// Read original content via Read
	orig, err := Read(v, "S01.11.11", "notes.md")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	_, err = AppendFile(v, "S01.11.11", "notes.md", "")
	if err != nil {
		t.Fatalf("AppendFile: %v", err)
	}

	after, _ := Read(v, "S01.11.11", "notes.md")
	if orig.Content != after.Content {
		t.Error("empty content should not modify file")
	}
}

func TestAppendFileIntegration_AddsNewlineIfMissing(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// Write a file without trailing newline
	_, _ = WriteFile(v, "S01.11.11", "no-newline.md", "Line one", "")

	_, err = AppendFile(v, "S01.11.11", "no-newline.md", "Line two")
	if err != nil {
		t.Fatalf("AppendFile: %v", err)
	}

	result, _ := Read(v, "S01.11.11", "no-newline.md")
	if result.Content != "Line one\nLine two" {
		t.Errorf("expected newline separator, got:\n%q", result.Content)
	}
}
