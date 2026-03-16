package vault

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestAcceptance_Create_FolderNameMatchesPattern(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Create(v, "S01.11", "Cinema")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	folderName := filepath.Base(result.Path)
	pattern := regexp.MustCompile(`^S\d{2}\.\d{2}\.\d{2} .+$`)
	if !pattern.MatchString(folderName) {
		t.Errorf("folder name %q does not match JD ID pattern", folderName)
	}
}

func TestAcceptance_Create_JDexNamedAfterFolder(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Create(v, "S01.11", "Cinema")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	folderName := filepath.Base(result.Path)
	jdexPath := filepath.Join(result.Path, folderName+".md")
	if _, err := os.Stat(jdexPath); os.IsNotExist(err) {
		t.Errorf("JDex file should be named after folder: %s", jdexPath)
	}
}

func TestAcceptance_Create_ReparseIncludesNewID(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Create(v, "S01.12", "Sushi Guide")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Re-parse the vault
	v2, err := ParseVault(root)
	if err != nil {
		t.Fatalf("re-ParseVault: %v", err)
	}

	// Search for the new ID
	results, err := Search(v2, result.Ref)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for %s after reparse, got %d", result.Ref, len(results))
	}
	if results[0].Name != "Sushi Guide" {
		t.Errorf("Name = %q, want Sushi Guide", results[0].Name)
	}
}
