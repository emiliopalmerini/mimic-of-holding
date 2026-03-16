package vault

import (
	"testing"
)

// --- Rename unit tests ---

func TestRename_EmptyRef(t *testing.T) {
	_, err := Rename(searchFixture, "", "New Name")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRename_EmptyName(t *testing.T) {
	_, err := Rename(searchFixture, "S01.11.11", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRename_InvalidRef(t *testing.T) {
	_, err := Rename(searchFixture, "xyz", "New Name")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRename_NotFound(t *testing.T) {
	_, err := Rename(searchFixture, "S99.99.99", "New Name")
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- Move unit tests ---

func TestMove_EmptyRef(t *testing.T) {
	_, err := Move(searchFixture, "", "S01.12")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMove_EmptyTarget(t *testing.T) {
	_, err := Move(searchFixture, "S01.11.11", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMove_InvalidRef(t *testing.T) {
	_, err := Move(searchFixture, "xyz", "S01.12")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMove_InvalidTarget(t *testing.T) {
	_, err := Move(searchFixture, "S01.11.11", "xyz")
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- MoveFile unit tests ---

func TestMoveFile_EmptyRef(t *testing.T) {
	_, err := MoveFile(searchFixture, "", "test.md", "S01.11.11")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMoveFile_EmptyFilename(t *testing.T) {
	_, err := MoveFile(searchFixture, "S01.11.11", "", "S01.11.01")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMoveFile_EmptyTarget(t *testing.T) {
	_, err := MoveFile(searchFixture, "S01.11.11", "test.md", "")
	if err == nil {
		t.Fatal("expected error")
	}
}
