package vault

import (
	"testing"
)

func TestCreate_EmptyName(t *testing.T) {
	_, err := Create(searchFixture, "S01.11", "", "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestCreate_InvalidCategoryRef_Scope(t *testing.T) {
	_, err := Create(searchFixture, "S01", "Cinema", "")
	if err == nil {
		t.Fatal("expected error for scope ref")
	}
}

func TestCreate_InvalidCategoryRef_Garbage(t *testing.T) {
	_, err := Create(searchFixture, "xyz", "Cinema", "")
	if err == nil {
		t.Fatal("expected error for invalid ref")
	}
}

func TestCreate_InvalidCategoryRef_ID(t *testing.T) {
	_, err := Create(searchFixture, "S01.11.11", "Cinema", "")
	if err == nil {
		t.Fatal("expected error for ID ref")
	}
}

func TestCreate_CategoryNotFound(t *testing.T) {
	_, err := Create(searchFixture, "S99.99", "Cinema", "")
	if err == nil {
		t.Fatal("expected error for category not found")
	}
}
