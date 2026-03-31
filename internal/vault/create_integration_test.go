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
	result, err := Create(v, "S01.11", "Cinema", "", nil)
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
	result, err := Create(v, "S01.12", "Pasta Recipes", "", nil)
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
	result, err := Create(v, "S01.10", "Notes", "", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if result.Ref != "S01.10.11" {
		t.Errorf("Ref = %q, want S01.10.11", result.Ref)
	}
}

// --- Auto-create hierarchy tests ---

func TestCreateIntegration_AutoCreateCategory(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// S01.13 doesn't exist but area S01.10-19 does
	result, err := Create(v, "S01.13", "Fitness", "", nil, CreateOpts{CategoryName: "Fitness"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if result.Ref != "S01.13.11" {
		t.Errorf("Ref = %q, want S01.13.11", result.Ref)
	}

	// Verify category folder was created
	v2, err := ParseVault(root)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	cat, err := findCategory(v2, 1, 13)
	if err != nil {
		t.Fatalf("category should exist after auto-create: %v", err)
	}
	if cat.Name != "Fitness" {
		t.Errorf("category name = %q, want Fitness", cat.Name)
	}
}

func TestCreateIntegration_AutoCreateAreaAndCategory(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// S01.30-39 area doesn't exist, nor does S01.31 category
	result, err := Create(v, "S01.31", "Running", "", nil, CreateOpts{
		CategoryName: "Sports",
		AreaName:     "Health & Fitness",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if result.Ref != "S01.31.11" {
		t.Errorf("Ref = %q, want S01.31.11", result.Ref)
	}

	// Verify area and category exist after re-parse
	v2, err := ParseVault(root)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}

	area := findAreaForCategory(v2, 1, 31)
	if area == nil {
		t.Fatal("area S01.30-39 should exist after auto-create")
	}
	if area.Name != "Health & Fitness" {
		t.Errorf("area name = %q, want Health & Fitness", area.Name)
	}
	if area.RangeStart != 30 || area.RangeEnd != 39 {
		t.Errorf("area range = %d-%d, want 30-39", area.RangeStart, area.RangeEnd)
	}

	cat, err := findCategory(v2, 1, 31)
	if err != nil {
		t.Fatalf("category should exist: %v", err)
	}
	if cat.Name != "Sports" {
		t.Errorf("category name = %q, want Sports", cat.Name)
	}
}

func TestCreateIntegration_AutoCreateCategoryMissingName(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// S01.13 doesn't exist, no CategoryName provided
	_, err = Create(v, "S01.13", "Yoga", "", nil, CreateOpts{})
	if err == nil {
		t.Fatal("expected error when category doesn't exist and CategoryName is empty")
	}
}

func TestCreateIntegration_AutoCreateAreaMissingName(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// S01.30-39 area doesn't exist, AreaName not provided
	_, err = Create(v, "S01.31", "Running", "", nil, CreateOpts{CategoryName: "Sports"})
	if err == nil {
		t.Fatal("expected error when area doesn't exist and AreaName is empty")
	}
}

func TestCreateIntegration_AutoCreateScopeNotFound(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// S99 scope doesn't exist — should still error
	_, err = Create(v, "S99.11", "Test", "", nil, CreateOpts{
		CategoryName: "Test Cat",
		AreaName:     "Test Area",
	})
	if err == nil {
		t.Fatal("expected error when scope doesn't exist")
	}
}

func TestCreateIntegration_ExistingCategoryOptsIgnored(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// S01.11 exists — opts should be ignored, normal behavior
	result, err := Create(v, "S01.11", "Cinema", "", nil, CreateOpts{
		CategoryName: "Should Be Ignored",
		AreaName:     "Should Be Ignored",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if result.Ref != "S01.11.12" {
		t.Errorf("Ref = %q, want S01.11.12", result.Ref)
	}
}

func TestCreateIntegration_JDexContent(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Create(v, "S01.12", "Pasta Recipes", "", nil)
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
