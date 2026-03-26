package vault

import "testing"

func TestRenameFile_EmptyRef(t *testing.T) {
	_, err := RenameFile(searchFixture, "", "old.md", "new.md")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRenameFile_NonIDRef(t *testing.T) {
	_, err := RenameFile(searchFixture, "S01", "old.md", "new.md")
	if err == nil {
		t.Fatal("expected error for scope ref")
	}
	_, err = RenameFile(searchFixture, "S01.11", "old.md", "new.md")
	if err == nil {
		t.Fatal("expected error for category ref")
	}
	_, err = RenameFile(searchFixture, "S01.10-19", "old.md", "new.md")
	if err == nil {
		t.Fatal("expected error for area ref")
	}
}

func TestRenameFile_EmptyOldName(t *testing.T) {
	_, err := RenameFile(searchFixture, "S01.11.11", "", "new.md")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRenameFile_EmptyNewName(t *testing.T) {
	_, err := RenameFile(searchFixture, "S01.11.11", "old.md", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRenameFile_SameName(t *testing.T) {
	result, err := RenameFile(searchFixture, "S01.11.11", "notes.md", "notes.md")
	if err != nil {
		t.Fatalf("expected no-op, got error: %v", err)
	}
	if result.LinksUpdated != 0 {
		t.Errorf("expected 0 links updated for no-op, got %d", result.LinksUpdated)
	}
}
