package vault

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestReadIntegration_Scope(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Read(v, "S01", "")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if result.Type != "scope" {
		t.Errorf("Type = %q, want scope", result.Type)
	}
	// S01 has 3 areas
	if len(result.Children) != 3 {
		t.Errorf("Children = %v, want 3 areas", result.Children)
	}
}

func TestReadIntegration_Area(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Read(v, "S01.10-19", "")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if result.Type != "area" {
		t.Errorf("Type = %q, want area", result.Type)
	}
	// S01.10-19 has 3 categories: S01.10, S01.11, S01.12
	if len(result.Children) != 3 {
		t.Errorf("Children = %v, want 3 categories", result.Children)
	}
}

func TestReadIntegration_Category(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Read(v, "S01.11", "")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if result.Type != "category" {
		t.Errorf("Type = %q, want category", result.Type)
	}
	// S01.11 has 2 IDs
	if len(result.Children) != 2 {
		t.Errorf("Children = %v, want 2 IDs", result.Children)
	}
}

func TestReadIntegration_IDWithJDex(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Read(v, "S01.11.11", "")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if result.Type != "id" {
		t.Errorf("Type = %q, want id", result.Type)
	}
	if result.Content == "" {
		t.Error("Content (JDex) should not be empty")
	}
	if !slices.Contains(result.Files, "notes.md") {
		t.Errorf("Files = %v, expected notes.md", result.Files)
	}
}

func TestReadIntegration_IDNoJDex(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Read(v, "S01.11.01", "")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if result.Content != "" {
		t.Errorf("Content should be empty for ID without JDex, got %q", result.Content)
	}
}

func TestReadIntegration_File(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Read(v, "S01.11.11", "notes.md")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if result.Type != "file" {
		t.Errorf("Type = %q, want file", result.Type)
	}
	if !strings.Contains(result.Content, "theatre season") {
		t.Errorf("Content should contain file text, got:\n%s", result.Content)
	}
}

func TestReadIntegration_FileNotFound(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = Read(v, "S01.11.11", "nonexistent.md")
	if err == nil {
		t.Fatal("expected error for file not found")
	}
}

func TestReadDeepIntegration_Area(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := ReadDeep(v, "S01.10-19", "")
	if err != nil {
		t.Fatalf("ReadDeep: %v", err)
	}

	if result.Type != "area" {
		t.Errorf("Type = %q, want area", result.Type)
	}
	if len(result.DeepChildren) == 0 {
		t.Fatal("deep read area should have DeepChildren")
	}
	// Should contain categories, which contain IDs
	foundID := false
	for _, cat := range result.DeepChildren {
		if cat.Type != "category" {
			t.Errorf("area deep child should be category, got %q", cat.Type)
		}
		for _, id := range cat.DeepChildren {
			if id.Type == "id" {
				foundID = true
			}
		}
	}
	if !foundID {
		t.Error("deep read area should contain IDs in nested categories")
	}
}

func TestReadDeepIntegration_Category(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := ReadDeep(v, "S01.11", "")
	if err != nil {
		t.Fatalf("ReadDeep: %v", err)
	}

	if len(result.DeepChildren) == 0 {
		t.Fatal("deep read category should have DeepChildren")
	}
	// Check that IDs have content
	foundContent := false
	for _, id := range result.DeepChildren {
		if id.Content != "" {
			foundContent = true
		}
	}
	if !foundContent {
		t.Error("deep read category should include JDex content from IDs")
	}
}

func TestReadIntegration_ScopeNotFound(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = Read(v, "S99", "")
	if err == nil {
		t.Fatal("expected error for scope not found")
	}
}
