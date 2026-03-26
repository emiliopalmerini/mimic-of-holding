package vault

import (
	"os"
	"testing"
)

func TestEditFileIntegration_ReplaceText(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	path, err := EditFile(v, "S01.11.11", "notes.md", "extra notes", "detailed notes")
	if err != nil {
		t.Fatalf("EditFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading edited file: %v", err)
	}
	want := "# notes\n\nSome detailed notes about the theatre season.\n"
	if string(data) != want {
		t.Errorf("content = %q, want %q", string(data), want)
	}
}

func TestEditFileIntegration_OldStringNotFound(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// Read original content first
	result, err := Read(v, "S01.11.11", "notes.md")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	original := result.Content

	_, err = EditFile(v, "S01.11.11", "notes.md", "nonexistent text", "replacement")
	if err == nil {
		t.Fatal("expected error for old_string not found")
	}

	// Verify file unchanged
	result2, err := Read(v, "S01.11.11", "notes.md")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if result2.Content != original {
		t.Errorf("file should be unchanged, got %q", result2.Content)
	}
}

func TestEditFileIntegration_AmbiguousMatch(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// Write a file with duplicate text
	path, err := WriteFile(v, "S01.11.11", "dupes.md", "foo bar foo", "")
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err = EditFile(v, "S01.11.11", "dupes.md", "foo", "baz")
	if err == nil {
		t.Fatal("expected error for ambiguous match")
	}

	// Verify file unchanged
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(data) != "foo bar foo" {
		t.Errorf("file should be unchanged, got %q", string(data))
	}
}

func TestEditFileIntegration_EmptyNewString(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	path, err := EditFile(v, "S01.11.11", "notes.md", "extra ", "")
	if err != nil {
		t.Fatalf("EditFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading edited file: %v", err)
	}
	want := "# notes\n\nSome notes about the theatre season.\n"
	if string(data) != want {
		t.Errorf("content = %q, want %q", string(data), want)
	}
}

func TestEditFileIntegration_MultilineReplace(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = WriteFile(v, "S01.11.11", "multi.md", "line one\nline two\nline three\n", "")
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	path, err := EditFile(v, "S01.11.11", "multi.md", "line two\nline three", "line 2\nline 3\nline 4")
	if err != nil {
		t.Fatalf("EditFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading edited file: %v", err)
	}
	want := "line one\nline 2\nline 3\nline 4\n"
	if string(data) != want {
		t.Errorf("content = %q, want %q", string(data), want)
	}
}

func TestEditFileIntegration_FileNotFound(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = EditFile(v, "S01.11.11", "nonexistent.md", "old", "new")
	if err == nil {
		t.Fatal("expected error for file not found")
	}
}

func TestEditFileIntegration_PreservesRestOfFile(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = WriteFile(v, "S01.11.11", "preserve.md", "header\n\nbody content here\n\nfooter\n", "")
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	path, err := EditFile(v, "S01.11.11", "preserve.md", "body content here", "updated body")
	if err != nil {
		t.Fatalf("EditFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading edited file: %v", err)
	}
	want := "header\n\nupdated body\n\nfooter\n"
	if string(data) != want {
		t.Errorf("content = %q, want %q", string(data), want)
	}
}
