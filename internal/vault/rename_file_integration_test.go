package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenameFileIntegration_RegularFile(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := RenameFile(v, "S01.11.11", "notes.md", "theatre-notes.md")
	if err != nil {
		t.Fatalf("RenameFile: %v", err)
	}

	// Old file should be gone
	if _, err := os.Stat(result.OldPath); !os.IsNotExist(err) {
		t.Error("old file should not exist")
	}

	// New file should exist
	if _, err := os.Stat(result.NewPath); os.IsNotExist(err) {
		t.Error("new file should exist")
	}

	if result.IsJDex {
		t.Error("should not be marked as JDex")
	}
	if result.FolderRenamed {
		t.Error("folder should not be renamed for regular file")
	}
}

func TestRenameFileIntegration_WikiLinksUpdated(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := RenameFile(v, "S01.11.11", "notes.md", "theatre-notes.md")
	if err != nil {
		t.Fatalf("RenameFile: %v", err)
	}

	if result.LinksUpdated == 0 {
		t.Error("expected wiki links to be updated")
	}

	// Check that the S02 file wikilink was updated
	s02File := filepath.Join(root, "S02 Due Draghi", "S02.10-19 Due Draghi al Microfono",
		"S02.11 Episodes", "S02.11.17 Season 7 Episode 1", "S02.11.17 Season 7 Episode 1.md")
	data, _ := os.ReadFile(s02File)
	content := string(data)
	if strings.Contains(content, "[[notes]]") {
		t.Error("old wikilink [[notes]] should be updated")
	}
	if !strings.Contains(content, "[[theatre-notes]]") {
		t.Error("new wikilink [[theatre-notes]] should be present")
	}
}

func TestRenameFileIntegration_HeadingUpdated(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := RenameFile(v, "S01.11.11", "notes.md", "theatre-notes.md")
	if err != nil {
		t.Fatalf("RenameFile: %v", err)
	}

	if !result.HeadingUpdated {
		t.Error("heading should have been updated")
	}

	data, _ := os.ReadFile(result.NewPath)
	content := string(data)
	if strings.Contains(content, "# notes\n") {
		t.Error("old heading should be replaced")
	}
	if !strings.Contains(content, "# theatre-notes\n") {
		t.Error("new heading should be present")
	}
}

func TestRenameFileIntegration_HeadingNotMatchingLeftAlone(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// new-show-idea.md in S01.11.01 has no H1 matching its stem
	// Create a file with a different H1
	idPath := filepath.Join(root, "S01 Me", "S01.10-19 Lifestyle",
		"S01.11 Entertainment", "S01.11.01 Inbox for S01.11")
	testFile := filepath.Join(idPath, "draft.md")
	os.WriteFile(testFile, []byte("# My Custom Title\n\nSome content.\n"), 0o644)

	// Re-parse to pick up the new file
	v, _ = ParseVault(root)

	result, err := RenameFile(v, "S01.11.01", "draft.md", "final.md")
	if err != nil {
		t.Fatalf("RenameFile: %v", err)
	}

	if result.HeadingUpdated {
		t.Error("heading should not be updated when it doesn't match the stem")
	}

	data, _ := os.ReadFile(result.NewPath)
	if !strings.Contains(string(data), "# My Custom Title") {
		t.Error("original heading should be preserved")
	}
}

func TestRenameFileIntegration_JDexFile(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// Rename the JDex file — this should rename the folder too
	result, err := RenameFile(v, "S01.11.11", "S01.11.11 Theatre, 2025 Season.md", "S01.11.11 Cinema, 2025 Season.md")
	if err != nil {
		t.Fatalf("RenameFile: %v", err)
	}

	if !result.IsJDex {
		t.Error("should be marked as JDex")
	}
	if !result.FolderRenamed {
		t.Error("folder should be renamed for JDex file")
	}

	// New folder should exist with new name
	newFolderPath := filepath.Dir(result.NewPath)
	if !strings.Contains(filepath.Base(newFolderPath), "Cinema, 2025 Season") {
		t.Errorf("folder should be renamed, got %q", filepath.Base(newFolderPath))
	}

	// JDex file should exist at new path
	if _, err := os.Stat(result.NewPath); os.IsNotExist(err) {
		t.Error("JDex file should exist at new path")
	}

	// Frontmatter should be updated
	data, _ := os.ReadFile(result.NewPath)
	content := string(data)
	if !strings.Contains(content, "S01.11.11 Cinema, 2025 Season") {
		t.Error("JDex frontmatter should be updated")
	}
}

func TestRenameFileIntegration_FileNotFound(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = RenameFile(v, "S01.11.11", "nonexistent.md", "new.md")
	if err == nil {
		t.Fatal("expected error for file not found")
	}
}

func TestRenameFileIntegration_TargetExists(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// notes.md and S01.11.11 Theatre, 2025 Season.md both exist in S01.11.11
	_, err = RenameFile(v, "S01.11.11", "notes.md", "S01.11.11 Theatre, 2025 Season.md")
	if err == nil {
		t.Fatal("expected error when target already exists")
	}
}

func TestRenameFileIntegration_PipedWikiLinks(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// The S02 file has [[Theatre, 2025 Season|the theatre season]]
	// Renaming the JDex should preserve the display text
	result, err := RenameFile(v, "S01.11.11", "S01.11.11 Theatre, 2025 Season.md", "S01.11.11 Cinema, 2025 Season.md")
	if err != nil {
		t.Fatalf("RenameFile: %v", err)
	}

	if result.LinksUpdated == 0 {
		t.Error("expected wiki links to be updated")
	}

	s02File := filepath.Join(root, "S02 Due Draghi", "S02.10-19 Due Draghi al Microfono",
		"S02.11 Episodes", "S02.11.17 Season 7 Episode 1", "S02.11.17 Season 7 Episode 1.md")
	data, _ := os.ReadFile(s02File)
	content := string(data)
	if !strings.Contains(content, "Cinema, 2025 Season|the theatre season]]") {
		t.Errorf("piped wikilink display should be preserved, got:\n%s", content)
	}
}

func TestRenameFileIntegration_AutoAppendMdExtension(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// Omit .md extension — should be auto-appended
	result, err := RenameFile(v, "S01.11.11", "notes", "theatre-notes")
	if err != nil {
		t.Fatalf("RenameFile: %v", err)
	}

	if !strings.HasSuffix(result.NewPath, "theatre-notes.md") {
		t.Errorf("new path should end with .md, got %q", result.NewPath)
	}
}

func TestRenameFileIntegration_AcceptanceReparse(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = RenameFile(v, "S01.11.11", "notes.md", "theatre-notes.md")
	if err != nil {
		t.Fatalf("RenameFile: %v", err)
	}

	// Re-parse and read back
	v2, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault after rename: %v", err)
	}

	result, err := Read(v2, "S01.11.11", "")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	found := false
	for _, f := range result.Files {
		if strings.Contains(f, "theatre-notes.md") {
			found = true
		}
	}
	if !found {
		t.Error("renamed file should appear in file list")
	}
}

func TestRenameFileIntegration_JDexAcceptanceBrowse(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = RenameFile(v, "S01.11.11", "S01.11.11 Theatre, 2025 Season.md", "S01.11.11 Cinema, 2025 Season.md")
	if err != nil {
		t.Fatalf("RenameFile: %v", err)
	}

	// Re-parse and browse
	v2, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault after rename: %v", err)
	}

	results, err := Search(v2, "Cinema", SearchOpts{})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Error("renamed JDex ID should be findable by new name")
	}
}
