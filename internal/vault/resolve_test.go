package vault

import (
	"testing"
)

func TestResolve_EmptyRef(t *testing.T) {
	_, err := Resolve(searchFixture, "", "")
	if err == nil {
		t.Fatal("expected error for empty ref")
	}
}

func TestResolve_Scope(t *testing.T) {
	path, err := Resolve(searchFixture, "S01", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/tmp/S01 Me" {
		t.Errorf("got %q, want /tmp/S01 Me", path)
	}
}

func TestResolve_ScopeNotFound(t *testing.T) {
	_, err := Resolve(searchFixture, "S99", "")
	if err == nil {
		t.Fatal("expected error for scope not found")
	}
}

func TestResolve_Area(t *testing.T) {
	path, err := Resolve(searchFixture, "S01.10-19", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/tmp/S01 Me/S01.10-19 Lifestyle" {
		t.Errorf("got %q, want /tmp/S01 Me/S01.10-19 Lifestyle", path)
	}
}

func TestResolve_AreaNotFound(t *testing.T) {
	_, err := Resolve(searchFixture, "S01.90-99", "")
	if err == nil {
		t.Fatal("expected error for area not found")
	}
}

func TestResolve_Category(t *testing.T) {
	path, err := Resolve(searchFixture, "S01.11", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/tmp/S01 Me/S01.10-19 Lifestyle/S01.11 Entertainment" {
		t.Errorf("got %q, want /tmp/S01 Me/S01.10-19 Lifestyle/S01.11 Entertainment", path)
	}
}

func TestResolve_CategoryNotFound(t *testing.T) {
	_, err := Resolve(searchFixture, "S01.99", "")
	if err == nil {
		t.Fatal("expected error for category not found")
	}
}

func TestResolve_ID(t *testing.T) {
	path, err := Resolve(searchFixture, "S01.11.11", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "/tmp/S01 Me/S01.10-19 Lifestyle/S01.11 Entertainment/S01.11.11 Theatre, 2025 Season"
	if path != want {
		t.Errorf("got %q, want %q", path, want)
	}
}

func TestResolve_IDNotFound(t *testing.T) {
	_, err := Resolve(searchFixture, "S01.11.99", "")
	if err == nil {
		t.Fatal("expected error for ID not found")
	}
}

func TestResolve_IDWithFile(t *testing.T) {
	path, err := Resolve(searchFixture, "S01.11.11", "notes.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "/tmp/S01 Me/S01.10-19 Lifestyle/S01.11 Entertainment/S01.11.11 Theatre, 2025 Season/notes.md"
	if path != want {
		t.Errorf("got %q, want %q", path, want)
	}
}

func TestResolve_FileWithNonIDRef(t *testing.T) {
	_, err := Resolve(searchFixture, "S01", "notes.md")
	if err == nil {
		t.Fatal("expected error for file arg with scope ref")
	}
	_, err = Resolve(searchFixture, "S01.10-19", "notes.md")
	if err == nil {
		t.Fatal("expected error for file arg with area ref")
	}
	_, err = Resolve(searchFixture, "S01.11", "notes.md")
	if err == nil {
		t.Fatal("expected error for file arg with category ref")
	}
}

func TestResolve_InvalidRef(t *testing.T) {
	_, err := Resolve(searchFixture, "garbage", "")
	if err == nil {
		t.Fatal("expected error for invalid ref")
	}
}
