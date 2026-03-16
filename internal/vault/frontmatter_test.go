package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Unit tests

func TestSetFrontmatter_InvalidRef(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = SetFrontmatter(v, "bad", "notes.md", "key", "value")
	if err == nil {
		t.Fatal("expected error for invalid ref")
	}
}

func TestSetFrontmatter_FileNotFound(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = SetFrontmatter(v, "S01.11.11", "nonexistent.md", "key", "value")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestAddToFrontmatterList_InvalidRef(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = AddToFrontmatterList(v, "bad", "notes.md", "tags", "new")
	if err == nil {
		t.Fatal("expected error for invalid ref")
	}
}

func TestRemoveFromFrontmatterList_InvalidRef(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = RemoveFromFrontmatterList(v, "bad", "notes.md", "tags", "old")
	if err == nil {
		t.Fatal("expected error for invalid ref")
	}
}

// Integration tests — use a temp copy to avoid mutating fixtures

func copyTestVault(t *testing.T) (string, *Vault) {
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
		return os.WriteFile(target, data, info.Mode())
	})
	if err != nil {
		t.Fatalf("copy vault: %v", err)
	}

	v, err := ParseVault(dst)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	return dst, v
}

func TestSetFrontmatterIntegration_SetExistingField(t *testing.T) {
	_, v := copyTestVault(t)

	// S01.11.11 JDex has location: Obsidian — change it
	path, err := SetFrontmatter(v, "S01.11.11", "S01.11.11 Theatre, 2025 Season.md", "location", "Notion")
	if err != nil {
		t.Fatalf("SetFrontmatter: %v", err)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "location: Notion") {
		t.Errorf("expected 'location: Notion' in file, got:\n%s", data)
	}
	// Should not contain old value
	if strings.Contains(string(data), "location: Obsidian") {
		t.Error("old value should be replaced")
	}
}

func TestSetFrontmatterIntegration_AddNewField(t *testing.T) {
	_, v := copyTestVault(t)

	path, err := SetFrontmatter(v, "S01.11.11", "S01.11.11 Theatre, 2025 Season.md", "status", "active")
	if err != nil {
		t.Fatalf("SetFrontmatter: %v", err)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "status: active") {
		t.Errorf("expected 'status: active' in file, got:\n%s", data)
	}
}

func TestSetFrontmatterIntegration_CreateFrontmatter(t *testing.T) {
	_, v := copyTestVault(t)

	// notes.md has no frontmatter
	path, err := SetFrontmatter(v, "S01.11.11", "notes.md", "status", "draft")
	if err != nil {
		t.Fatalf("SetFrontmatter: %v", err)
	}
	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.HasPrefix(content, "---\n") {
		t.Error("should start with frontmatter delimiter")
	}
	if !strings.Contains(content, "status: draft") {
		t.Errorf("expected 'status: draft' in file, got:\n%s", data)
	}
	// Original content should still be there
	if !strings.Contains(content, "Some extra notes") {
		t.Error("original content should be preserved")
	}
}

func TestAddToFrontmatterListIntegration_AddTag(t *testing.T) {
	_, v := copyTestVault(t)

	path, err := AddToFrontmatterList(v, "S01.11.11", "S01.11.11 Theatre, 2025 Season.md", "tags", "theatre")
	if err != nil {
		t.Fatalf("AddToFrontmatterList: %v", err)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "- theatre") {
		t.Errorf("expected '- theatre' in file, got:\n%s", data)
	}
	// Existing tags should still be there
	if !strings.Contains(string(data), "- jdex") {
		t.Error("existing tags should be preserved")
	}
}

func TestAddToFrontmatterListIntegration_Idempotent(t *testing.T) {
	_, v := copyTestVault(t)

	// jdex already exists
	path, err := AddToFrontmatterList(v, "S01.11.11", "S01.11.11 Theatre, 2025 Season.md", "tags", "jdex")
	if err != nil {
		t.Fatalf("AddToFrontmatterList: %v", err)
	}
	data, _ := os.ReadFile(path)
	count := strings.Count(string(data), "- jdex")
	if count != 1 {
		t.Errorf("expected exactly 1 occurrence of '- jdex', got %d", count)
	}
}

func TestRemoveFromFrontmatterListIntegration_RemoveTag(t *testing.T) {
	_, v := copyTestVault(t)

	path, err := RemoveFromFrontmatterList(v, "S01.11.11", "S01.11.11 Theatre, 2025 Season.md", "tags", "index")
	if err != nil {
		t.Fatalf("RemoveFromFrontmatterList: %v", err)
	}
	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "- index") {
		t.Error("'- index' should be removed")
	}
	// Other tags should remain
	if !strings.Contains(string(data), "- jdex") {
		t.Error("other tags should be preserved")
	}
}

func TestRemoveFromFrontmatterListIntegration_IdempotentNoOp(t *testing.T) {
	_, v := copyTestVault(t)

	// nonexistent tag — should be a no-op
	path, err := RemoveFromFrontmatterList(v, "S01.11.11", "S01.11.11 Theatre, 2025 Season.md", "tags", "nonexistent")
	if err != nil {
		t.Fatalf("RemoveFromFrontmatterList: %v", err)
	}
	data, _ := os.ReadFile(path)
	// File should be unchanged
	if !strings.Contains(string(data), "- jdex") {
		t.Error("file should be unchanged")
	}
}

// Acceptance tests

func TestAcceptance_Frontmatter_PreservesContent(t *testing.T) {
	_, v := copyTestVault(t)

	path, err := SetFrontmatter(v, "S01.11.11", "S01.11.11 Theatre, 2025 Season.md", "status", "active")
	if err != nil {
		t.Fatalf("SetFrontmatter: %v", err)
	}
	data, _ := os.ReadFile(path)
	content := string(data)

	// Body content should be preserved
	if !strings.Contains(content, "Shakespeare comedy") {
		t.Error("body content should be preserved")
	}
	// Frontmatter should be valid (starts and ends with ---)
	if !strings.HasPrefix(content, "---\n") {
		t.Error("should start with ---")
	}
	// Count --- occurrences — should be exactly 2
	count := strings.Count(content[:strings.Index(content, "# S01")+1], "---")
	if count != 2 {
		t.Errorf("expected 2 frontmatter delimiters before body, got %d", count)
	}
}

func TestAcceptance_AddToList_CreatesList(t *testing.T) {
	_, v := copyTestVault(t)

	// notes.md has no frontmatter — adding a list item should create frontmatter with a list
	path, err := AddToFrontmatterList(v, "S01.11.11", "notes.md", "tags", "note")
	if err != nil {
		t.Fatalf("AddToFrontmatterList: %v", err)
	}
	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, "tags:") {
		t.Error("should create tags key")
	}
	if !strings.Contains(content, "- note") {
		t.Error("should contain '- note'")
	}
}
