package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- ListTemplates integration tests ---

func TestListTemplatesIntegration_CategoryWithTemplates(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	templates, err := ListTemplates(v, "S01.11")
	if err != nil {
		t.Fatalf("ListTemplates: %v", err)
	}

	// Should find "Event Review" from category level and "Default Note" from area level
	// "Event Review" at area level is shadowed by category level
	var names []string
	for _, tmpl := range templates {
		names = append(names, tmpl.Name)
	}

	found := false
	for _, n := range names {
		if n == "Event Review" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'Event Review' in templates, got %v", names)
	}

	foundDefault := false
	for _, n := range names {
		if n == "Default Note" {
			foundDefault = true
		}
	}
	if !foundDefault {
		t.Errorf("expected 'Default Note' from area level, got %v", names)
	}
}

func TestListTemplatesIntegration_CategoryLevel(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	templates, err := ListTemplates(v, "S01.11")
	if err != nil {
		t.Fatalf("ListTemplates: %v", err)
	}

	// "Event Review" from category should come before area-level templates
	// and should shadow the area-level "Event Review"
	for _, tmpl := range templates {
		if tmpl.Name == "Event Review" {
			if tmpl.Source != "category" {
				t.Errorf("Event Review should be from category level, got %q", tmpl.Source)
			}
			return
		}
	}
	t.Error("Event Review not found")
}

func TestListTemplatesIntegration_Shadowing(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	templates, err := ListTemplates(v, "S01.11")
	if err != nil {
		t.Fatalf("ListTemplates: %v", err)
	}

	// "Event Review" exists at both category and area level
	// Only the category-level one should appear
	count := 0
	for _, tmpl := range templates {
		if tmpl.Name == "Event Review" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 'Event Review' (shadowed), got %d", count)
	}
}

func TestListTemplatesIntegration_CategoryWithoutTemplates(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// S01.12 Food has no .03 templates ID
	templates, err := ListTemplates(v, "S01.12")
	if err != nil {
		t.Fatalf("ListTemplates: %v", err)
	}

	// Should still find area-level templates from S01.10.03
	if len(templates) == 0 {
		t.Error("expected area-level templates even when category has no .03")
	}
}

// --- Create with template integration tests ---

func TestCreateIntegration_WithTemplate(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Create(v, "S01.11", "Opera Night", "Event Review")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// JDex file should contain template content with substituted variables
	jdexPath := filepath.Join(result.Path, result.Ref+" "+result.Name+".md")
	data, err := os.ReadFile(jdexPath)
	if err != nil {
		t.Fatalf("reading JDex: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, result.Ref) {
		t.Error("template variable {{ref}} should be substituted")
	}
	if !strings.Contains(content, "Opera Night") {
		t.Error("template variable {{name}} should be substituted")
	}
	if strings.Contains(content, "{{name}}") {
		t.Error("{{name}} should be substituted")
	}
	if strings.Contains(content, "{{ref}}") {
		t.Error("{{ref}} should be substituted")
	}
	// Should contain template structure
	if !strings.Contains(content, "## Review") {
		t.Error("template structure should be present")
	}
}

func TestCreateIntegration_WithTemplateNotFound(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = Create(v, "S01.11", "Opera Night", "Nonexistent Template")
	if err == nil {
		t.Fatal("expected error for template not found")
	}

	// Folder should not be created
	entries, _ := os.ReadDir(filepath.Join(root, "S01 Me", "S01.10-19 Lifestyle", "S01.11 Entertainment"))
	for _, e := range entries {
		if strings.Contains(e.Name(), "Opera Night") {
			t.Error("folder should not be created when template not found")
		}
	}
}

func TestCreateIntegration_WithoutTemplate(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Create(v, "S01.11", "Opera Night", "")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Should use default hardcoded JDex
	jdexPath := filepath.Join(result.Path, result.Ref+" "+result.Name+".md")
	data, _ := os.ReadFile(jdexPath)
	content := string(data)
	if !strings.Contains(content, "## Contents") {
		t.Error("default JDex should have ## Contents")
	}
	if !strings.Contains(content, "jdex") {
		t.Error("default JDex should have jdex tag")
	}
}

// --- WriteFile with template integration tests ---

func TestWriteFileIntegration_WithTemplateEmptyContent(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	path, err := WriteFile(v, "S01.11.11", "review.md", "", "Event Review")
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "## Review") {
		t.Error("template structure should be present")
	}
	if strings.Contains(content, "{{name}}") {
		t.Error("{{name}} should be substituted")
	}
	if !strings.Contains(content, "Theatre, 2025 Season") {
		t.Error("{{name}} should be substituted with ID name")
	}
}

func TestWriteFileIntegration_WithTemplateContentProvided(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	path, err := WriteFile(v, "S01.11.11", "review.md", "My custom content", "Event Review")
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if content != "My custom content" {
		t.Errorf("content should win over template, got %q", content)
	}
}

func TestWriteFileIntegration_WithTemplateNotFound(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = WriteFile(v, "S01.11.11", "review.md", "", "Nonexistent Template")
	if err == nil {
		t.Fatal("expected error for template not found")
	}
}

func TestWriteFileIntegration_NoContentNoTemplate(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// Empty content with no template creates an empty file (backward compatible)
	path, err := WriteFile(v, "S01.11.11", "review.md", "", "")
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	info, _ := os.Stat(path)
	if info.Size() != 0 {
		t.Errorf("expected empty file, got %d bytes", info.Size())
	}
}

// --- Acceptance tests ---

func TestCreateWithTemplateIntegration_Acceptance(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Create(v, "S01.11", "Opera Night", "Event Review")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Re-parse and read back
	v2, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault after create: %v", err)
	}

	readResult, err := Read(v2, result.Ref, "")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if !strings.Contains(readResult.Content, "## Review") {
		t.Error("template content should be readable after create")
	}
}

func TestWriteFileWithTemplateIntegration_Acceptance(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = WriteFile(v, "S01.11.11", "review.md", "", "Event Review")
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Read back
	readResult, err := Read(v, "S01.11.11", "review.md")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if !strings.Contains(readResult.Content, "## Review") {
		t.Error("template content should be readable after write")
	}
}
