package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Rename integration tests ---

func TestRenameIntegration_ID(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Rename(v, "S01.11.11", "Cinema, 2025 Season")
	if err != nil {
		t.Fatalf("Rename: %v", err)
	}

	if result.OldName != "Theatre, 2025 Season" {
		t.Errorf("OldName = %q", result.OldName)
	}
	if result.NewName != "Cinema, 2025 Season" {
		t.Errorf("NewName = %q", result.NewName)
	}

	// Old path should not exist
	if _, err := os.Stat(result.OldPath); !os.IsNotExist(err) {
		t.Error("old path should not exist")
	}

	// New path should exist
	if _, err := os.Stat(result.NewPath); os.IsNotExist(err) {
		t.Error("new path should exist")
	}

	// New folder name should contain new name
	if !strings.Contains(filepath.Base(result.NewPath), "Cinema, 2025 Season") {
		t.Errorf("new folder name = %q", filepath.Base(result.NewPath))
	}

	// JDex file should be renamed
	jdexPath := filepath.Join(result.NewPath, "S01.11.11 Cinema, 2025 Season.md")
	if _, err := os.Stat(jdexPath); os.IsNotExist(err) {
		t.Error("JDex file should be renamed")
	}

	// JDex frontmatter should be updated
	data, _ := os.ReadFile(jdexPath)
	content := string(data)
	if !strings.Contains(content, "S01.11.11 Cinema, 2025 Season") {
		t.Errorf("JDex should contain new alias, got:\n%s", content)
	}
}

func TestRenameIntegration_Category(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Rename(v, "S01.12", "Cuisine")
	if err != nil {
		t.Fatalf("Rename: %v", err)
	}

	if !strings.Contains(filepath.Base(result.NewPath), "Cuisine") {
		t.Errorf("new path should contain Cuisine, got %q", filepath.Base(result.NewPath))
	}

	// Re-parse and verify
	v2, _ := ParseVault(root)
	results, _ := Search(v2, "Cuisine", SearchOpts{})
	if len(results) == 0 {
		t.Error("renamed category should be findable")
	}
}

func TestRenameIntegration_WikiLinksUpdated(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Rename(v, "S01.11.11", "Cinema, 2025 Season")
	if err != nil {
		t.Fatalf("Rename: %v", err)
	}

	if result.LinksUpdated == 0 {
		t.Error("expected wiki links to be updated")
	}

	// Check the S02 file that linked to the old name
	s02File := filepath.Join(root, "S02 Due Draghi", "S02.10-19 Due Draghi al Microfono",
		"S02.11 Episodes", "S02.11.17 Season 7 Episode 1", "S02.11.17 Season 7 Episode 1.md")
	data, _ := os.ReadFile(s02File)
	content := string(data)
	if strings.Contains(content, "Theatre, 2025 Season]]") {
		t.Error("old wiki link should be updated")
	}
	if !strings.Contains(content, "Cinema, 2025 Season") {
		t.Error("new wiki link should be present")
	}
}

// --- Move integration tests ---

func TestMoveIntegration_IDToCategory(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Move(v, "S01.11.11", "S01.12")
	if err != nil {
		t.Fatalf("Move: %v", err)
	}

	if result.OldRef != "S01.11.11" {
		t.Errorf("OldRef = %q", result.OldRef)
	}
	// Should get next available in S01.12 (which is empty → .11)
	if result.NewRef != "S01.12.11" {
		t.Errorf("NewRef = %q, want S01.12.11", result.NewRef)
	}

	// Old path gone, new path exists
	if _, err := os.Stat(result.OldPath); !os.IsNotExist(err) {
		t.Error("old path should not exist")
	}
	if _, err := os.Stat(result.NewPath); os.IsNotExist(err) {
		t.Error("new path should exist")
	}

	// Re-parse and verify findable at new location
	v2, _ := ParseVault(root)
	results, _ := Search(v2, result.NewRef, SearchOpts{})
	if len(results) != 1 {
		t.Errorf("moved item should be findable at %s", result.NewRef)
	}
}

func TestMoveIntegration_CategoryToArea(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Move(v, "S01.12", "S01.20-29")
	if err != nil {
		t.Fatalf("Move: %v", err)
	}

	// S01.20-29 Learning is empty → should keep number 12 if valid in range, else next available
	// 12 is in range 10-19 not 20-29, so needs reassignment → 21 (first available after 20)
	if !strings.HasPrefix(result.NewRef, "S01.2") {
		t.Errorf("NewRef = %q, expected S01.2x", result.NewRef)
	}

	if _, err := os.Stat(result.NewPath); os.IsNotExist(err) {
		t.Error("new path should exist")
	}
}

// --- MoveFile integration tests ---

func TestMoveFileIntegration(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	newPath, err := MoveFile(v, "S01.11.11", "notes.md", "S01.11.01")
	if err != nil {
		t.Fatalf("MoveFile: %v", err)
	}

	// File should exist at new location
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Error("file should exist at new location")
	}

	// File should be gone from old location
	oldPath := filepath.Join(root, "S01 Me", "S01.10-19 Lifestyle",
		"S01.11 Entertainment", "S01.11.11 Theatre, 2025 Season", "notes.md")
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Error("file should be gone from old location")
	}
}

func TestMoveFileIntegration_FileNotFound(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = MoveFile(v, "S01.11.11", "nonexistent.md", "S01.11.01")
	if err == nil {
		t.Fatal("expected error for file not found")
	}
}
