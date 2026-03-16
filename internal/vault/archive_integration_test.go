package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArchiveIntegration_ID(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Archive(v, "S01.11.11")
	if err != nil {
		t.Fatalf("Archive: %v", err)
	}

	// Original should not exist
	origPath := filepath.Join(root, "S01 Me", "S01.10-19 Lifestyle", "S01.11 Entertainment", "S01.11.11 Theatre, 2025 Season")
	if _, err := os.Stat(origPath); !os.IsNotExist(err) {
		t.Error("original path should no longer exist")
	}

	// New path should exist
	if _, err := os.Stat(result.NewPath); os.IsNotExist(err) {
		t.Fatalf("archived path does not exist: %s", result.NewPath)
	}

	// Should be renamed to [Archived] Name
	baseName := filepath.Base(result.NewPath)
	if !strings.HasPrefix(baseName, "[Archived]") {
		t.Errorf("archived folder should start with [Archived], got %q", baseName)
	}
	if !strings.Contains(baseName, "Theatre, 2025 Season") {
		t.Errorf("archived folder should contain original name, got %q", baseName)
	}
}

func TestArchiveIntegration_Category(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Archive(v, "S01.11")
	if err != nil {
		t.Fatalf("Archive: %v", err)
	}

	// Original should not exist
	origPath := filepath.Join(root, "S01 Me", "S01.10-19 Lifestyle", "S01.11 Entertainment")
	if _, err := os.Stat(origPath); !os.IsNotExist(err) {
		t.Error("original category path should no longer exist")
	}

	// New path should exist and keep ID
	if _, err := os.Stat(result.NewPath); os.IsNotExist(err) {
		t.Fatalf("archived path does not exist: %s", result.NewPath)
	}

	baseName := filepath.Base(result.NewPath)
	if !strings.HasPrefix(baseName, "S01.11") {
		t.Errorf("archived category should keep its ID, got %q", baseName)
	}
}

func TestArchiveIntegration_CreatesArchiveFolder(t *testing.T) {
	root := copyFixtureVault(t)

	// Ensure no .09 archive folder exists for S01.11
	archivePath := filepath.Join(root, "S01 Me", "S01.10-19 Lifestyle", "S01.11 Entertainment", "S01.11.09 Archive for S01.11")
	os.RemoveAll(archivePath) // remove if it exists

	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = Archive(v, "S01.11.11")
	if err != nil {
		t.Fatalf("Archive: %v", err)
	}

	// Archive folder should now exist
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Error("archive folder should have been created")
	}
}

func TestArchiveIntegration_IDNotFound(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = Archive(v, "S01.11.99")
	if err == nil {
		t.Fatal("expected error for ID not found")
	}
}

func TestArchiveIntegration_CategoryNotFound(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = Archive(v, "S99.99")
	if err == nil {
		t.Fatal("expected error for category not found")
	}
}
