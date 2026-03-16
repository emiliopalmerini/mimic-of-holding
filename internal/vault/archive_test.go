package vault

import (
	"testing"
)

func TestArchive_EmptyRef(t *testing.T) {
	_, err := Archive(searchFixture, "")
	if err == nil {
		t.Fatal("expected error for empty ref")
	}
}

func TestArchive_ScopeRef(t *testing.T) {
	_, err := Archive(searchFixture, "S01")
	if err == nil {
		t.Fatal("expected error for scope ref")
	}
}

func TestArchive_AreaRef(t *testing.T) {
	_, err := Archive(searchFixture, "S01.10-19")
	if err == nil {
		t.Fatal("expected error for area ref")
	}
}

func TestArchive_InvalidRef(t *testing.T) {
	_, err := Archive(searchFixture, "xyz")
	if err == nil {
		t.Fatal("expected error for invalid ref")
	}
}
