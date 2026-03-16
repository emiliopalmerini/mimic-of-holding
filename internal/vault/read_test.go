package vault

import (
	"testing"
)

func TestRead_EmptyRef(t *testing.T) {
	_, err := Read(searchFixture, "")
	if err == nil {
		t.Fatal("expected error for empty ref")
	}
}

func TestRead_ScopeRef(t *testing.T) {
	_, err := Read(searchFixture, "S01")
	if err == nil {
		t.Fatal("expected error for scope ref")
	}
}

func TestRead_AreaRef(t *testing.T) {
	_, err := Read(searchFixture, "S01.10-19")
	if err == nil {
		t.Fatal("expected error for area ref")
	}
}

func TestRead_CategoryRef(t *testing.T) {
	_, err := Read(searchFixture, "S01.11")
	if err == nil {
		t.Fatal("expected error for category ref")
	}
}

func TestRead_IDNotFound(t *testing.T) {
	_, err := Read(searchFixture, "S01.11.99")
	if err == nil {
		t.Fatal("expected error for ID not found")
	}
}
