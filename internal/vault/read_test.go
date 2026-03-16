package vault

import (
	"testing"
)

func TestRead_EmptyRef(t *testing.T) {
	_, err := Read(searchFixture, "", "")
	if err == nil {
		t.Fatal("expected error for empty ref")
	}
}

func TestRead_ScopeRef(t *testing.T) {
	result, err := Read(searchFixture, "S01", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Type != "scope" {
		t.Errorf("Type = %q, want scope", result.Type)
	}
	if result.Name != "Me" {
		t.Errorf("Name = %q, want Me", result.Name)
	}
	if len(result.Children) == 0 {
		t.Error("expected Children to list areas")
	}
}

func TestRead_AreaRef(t *testing.T) {
	result, err := Read(searchFixture, "S01.10-19", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Type != "area" {
		t.Errorf("Type = %q, want area", result.Type)
	}
	if len(result.Children) == 0 {
		t.Error("expected Children to list categories")
	}
}

func TestRead_CategoryRef(t *testing.T) {
	result, err := Read(searchFixture, "S01.11", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Type != "category" {
		t.Errorf("Type = %q, want category", result.Type)
	}
	if len(result.Children) == 0 {
		t.Error("expected Children to list IDs")
	}
}

func TestRead_IDRef(t *testing.T) {
	result, err := Read(searchFixture, "S01.11.11", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Type != "id" {
		t.Errorf("Type = %q, want id", result.Type)
	}
}

func TestRead_IDNotFound(t *testing.T) {
	_, err := Read(searchFixture, "S01.11.99", "")
	if err == nil {
		t.Fatal("expected error for ID not found")
	}
}

func TestRead_FileWithNonIDRef(t *testing.T) {
	_, err := Read(searchFixture, "S01", "somefile.md")
	if err == nil {
		t.Fatal("expected error when file param used with non-ID ref")
	}
}

func TestRead_InvalidRef(t *testing.T) {
	_, err := Read(searchFixture, "xyz", "")
	if err == nil {
		t.Fatal("expected error for invalid ref")
	}
}

func TestRead_DeepIDSameAsRegular(t *testing.T) {
	result, err := ReadDeep(searchFixture, "S01.11.11", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Type != "id" {
		t.Errorf("Type = %q, want id", result.Type)
	}
	if len(result.DeepChildren) != 0 {
		t.Error("deep read of ID should not have DeepChildren")
	}
}
