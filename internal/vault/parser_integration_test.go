package vault

import (
	"path/filepath"
	"runtime"
	"testing"
)

func testdataDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test file path")
	}
	return filepath.Join(filepath.Dir(filename), "..", "..", "testdata")
}

func TestParseVault_FullFixture(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v.Root != root {
		t.Errorf("Root = %q, want %q", v.Root, root)
	}

	// Fixture has S01 and S02
	if got := len(v.Scopes); got != 2 {
		t.Fatalf("got %d scopes, want 2", got)
	}

	// Scopes should be sorted by number
	if v.Scopes[0].Number != 1 || v.Scopes[1].Number != 2 {
		t.Errorf("scopes not sorted: got numbers %d, %d", v.Scopes[0].Number, v.Scopes[1].Number)
	}

	// S01 has 3 areas: 00-09, 10-19, 20-29
	s01 := v.Scopes[0]
	if got := len(s01.Areas); got != 3 {
		t.Fatalf("S01 got %d areas, want 3", got)
	}

	// S01.10-19 Lifestyle has 3 categories: S01.10, S01.11, S01.12
	lifestyle := s01.Areas[1]
	if got := len(lifestyle.Categories); got != 3 {
		t.Fatalf("S01.10-19 got %d categories, want 3", got)
	}

	// S01.11 Entertainment has 3 IDs: S01.11.01 (system), S01.11.03 (templates), S01.11.11 (regular)
	entertainment := lifestyle.Categories[1]
	if got := len(entertainment.IDs); got != 3 {
		t.Fatalf("S01.11 got %d IDs, want 3", got)
	}

	// S01.12 Food is empty
	food := lifestyle.Categories[2]
	if got := len(food.IDs); got != 0 {
		t.Errorf("S01.12 got %d IDs, want 0", got)
	}

	// S01.20-29 Learning is an empty area
	learning := s01.Areas[2]
	if got := len(learning.Categories); got != 0 {
		t.Errorf("S01.20-29 got %d categories, want 0", got)
	}

	// S02 has 1 area
	s02 := v.Scopes[1]
	if got := len(s02.Areas); got != 1 {
		t.Fatalf("S02 got %d areas, want 1", got)
	}
}

func TestParseVault_NonExistentRoot(t *testing.T) {
	_, err := ParseVault("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for non-existent path")
	}
}

func TestParseVault_RootIsFile(t *testing.T) {
	file := filepath.Join(testdataDir(t), "vault_as_file.txt")
	_, err := ParseVault(file)
	if err == nil {
		t.Fatal("expected error when root is a file")
	}
}
