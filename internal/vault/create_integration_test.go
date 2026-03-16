package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// copyFixtureVault creates a temporary copy of the test fixture vault for write tests.
func copyFixtureVault(t *testing.T) string {
	t.Helper()
	src := filepath.Join(testdataDir(t), "vault")
	dst := t.TempDir()

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatalf("copying fixture: %v", err)
	}
	return dst
}

func TestCreateIntegration_WithExistingIDs(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// S01.11 has IDs .01 (system) and .11 (regular) → next should be .12
	result, err := Create(v, "S01.11", "Cinema")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if result.Ref != "S01.11.12" {
		t.Errorf("Ref = %q, want S01.11.12", result.Ref)
	}
	if result.Name != "Cinema" {
		t.Errorf("Name = %q, want Cinema", result.Name)
	}

	// Verify folder exists
	if _, err := os.Stat(result.Path); os.IsNotExist(err) {
		t.Fatalf("folder not created at %s", result.Path)
	}

	// Verify JDex file exists
	jdexPath := filepath.Join(result.Path, "S01.11.12 Cinema.md")
	if _, err := os.Stat(jdexPath); os.IsNotExist(err) {
		t.Fatalf("JDex file not created at %s", jdexPath)
	}
}

func TestCreateIntegration_EmptyCategory(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// S01.12 Food has no IDs → first should be .11
	result, err := Create(v, "S01.12", "Pasta Recipes")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if result.Ref != "S01.12.11" {
		t.Errorf("Ref = %q, want S01.12.11", result.Ref)
	}
}

func TestCreateIntegration_OnlySystemIDs(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// S01.10 Management has only .01 (system) → first regular should be .11
	result, err := Create(v, "S01.10", "Notes")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if result.Ref != "S01.10.11" {
		t.Errorf("Ref = %q, want S01.10.11", result.Ref)
	}
}

func TestCreateIntegration_JDexContent(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Create(v, "S01.12", "Pasta Recipes")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	jdexPath := filepath.Join(result.Path, "S01.12.11 Pasta Recipes.md")
	data, err := os.ReadFile(jdexPath)
	if err != nil {
		t.Fatalf("reading JDex: %v", err)
	}

	content := string(data)
	for _, want := range []string{
		"aliases:",
		"S01.12.11 Pasta Recipes",
		"location: Obsidian",
		"jdex",
		"index",
		"# S01.12.11 Pasta Recipes",
		"## Contents",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("JDex missing %q\n\ngot:\n%s", want, content)
		}
	}
}
