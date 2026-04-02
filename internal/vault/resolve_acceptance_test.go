package vault

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAcceptance_Resolve_IDReturnsAbsolutePath(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	path, err := Resolve(v, "S01.11.11", "")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %q", path)
	}
	if !strings.Contains(path, "S01.11.11 Theatre, 2025 Season") {
		t.Errorf("path should contain ID folder name, got %q", path)
	}
}

func TestAcceptance_Resolve_IDWithFileReturnsFilePath(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	path, err := Resolve(v, "S01.11.11", "notes.md")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if !strings.HasSuffix(path, "notes.md") {
		t.Errorf("path should end with notes.md, got %q", path)
	}
}

func TestAcceptance_Resolve_NonExistentFileStillReturnsPath(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	path, err := Resolve(v, "S01.11.11", "does-not-exist.md")
	if err != nil {
		t.Fatalf("Resolve should not error for non-existent file: %v", err)
	}
	if !strings.HasSuffix(path, "does-not-exist.md") {
		t.Errorf("path should end with does-not-exist.md, got %q", path)
	}
}

func TestAcceptance_Resolve_CategoryReturnsPath(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	path, err := Resolve(v, "S01.11", "")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if !strings.Contains(path, "S01.11 Entertainment") {
		t.Errorf("path should contain category folder name, got %q", path)
	}
}

func TestAcceptance_Resolve_ScopeReturnsPath(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	path, err := Resolve(v, "S01", "")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if !strings.Contains(path, "S01 Me") {
		t.Errorf("path should contain scope folder name, got %q", path)
	}
}
