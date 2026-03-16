package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAcceptance_Archive_IDNoLongerInTree(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = Archive(v, "S01.11.11")
	if err != nil {
		t.Fatalf("Archive: %v", err)
	}

	v2, err := ParseVault(root)
	if err != nil {
		t.Fatalf("re-ParseVault: %v", err)
	}

	results, err := Search(v2, "S01.11.11", SearchOpts{})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Error("archived ID should no longer appear in vault tree")
	}
}

func TestAcceptance_Archive_IDRenamedWithoutJDPrefix(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Archive(v, "S01.11.11")
	if err != nil {
		t.Fatalf("Archive: %v", err)
	}

	baseName := filepath.Base(result.NewPath)
	if strings.HasPrefix(baseName, "S01") {
		t.Errorf("archived ID should lose JD prefix, got %q", baseName)
	}
}

func TestAcceptance_Archive_CategoryContentsPreserved(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Archive(v, "S01.11")
	if err != nil {
		t.Fatalf("Archive: %v", err)
	}

	// The category had IDs inside — check they still exist under new path
	entries, err := os.ReadDir(result.NewPath)
	if err != nil {
		t.Fatalf("reading archived category: %v", err)
	}

	foundID := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "S01.11.") {
			foundID = true
			break
		}
	}
	if !foundID {
		t.Error("archived category should still contain its ID folders")
	}
}
