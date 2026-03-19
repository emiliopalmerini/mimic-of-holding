package vault

import (
	"testing"
)

func TestEditFile_EmptyRef(t *testing.T) {
	_, err := EditFile(searchFixture, "", "test.md", "old", "new")
	if err == nil {
		t.Fatal("expected error for empty ref")
	}
}

func TestEditFile_NonIDRef(t *testing.T) {
	_, err := EditFile(searchFixture, "S01", "test.md", "old", "new")
	if err == nil {
		t.Fatal("expected error for non-ID ref")
	}
}

func TestEditFile_EmptyFilename(t *testing.T) {
	_, err := EditFile(searchFixture, "S01.11.11", "", "old", "new")
	if err == nil {
		t.Fatal("expected error for empty filename")
	}
}

func TestEditFile_EmptyOldString(t *testing.T) {
	_, err := EditFile(searchFixture, "S01.11.11", "test.md", "", "new")
	if err == nil {
		t.Fatal("expected error for empty old_string")
	}
}

func TestEditFile_NoOpWhenEqual(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = EditFile(v, "S01.11.11", "notes.md", "extra", "extra")
	if err != nil {
		t.Fatalf("expected no-op, got error: %v", err)
	}
}
