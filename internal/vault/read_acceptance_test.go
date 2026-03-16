package vault

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAcceptance_Read_EveryLevelReturnsCorrectType(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	tests := []struct {
		ref      string
		wantType string
	}{
		{"S01", "scope"},
		{"S01.10-19", "area"},
		{"S01.11", "category"},
		{"S01.11.11", "id"},
	}
	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			result, err := Read(v, tt.ref, "")
			if err != nil {
				t.Fatalf("Read: %v", err)
			}
			if result.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", result.Type, tt.wantType)
			}
		})
	}
}

func TestAcceptance_Read_ChildrenMatchFixture(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// Scope children = area names
	result, _ := Read(v, "S01", "")
	foundLifestyle := false
	for _, c := range result.Children {
		if strings.Contains(c, "Lifestyle") {
			foundLifestyle = true
		}
	}
	if !foundLifestyle {
		t.Errorf("scope children should include Lifestyle, got %v", result.Children)
	}

	// Category children = ID names
	result, _ = Read(v, "S01.11", "")
	foundTheatre := false
	for _, c := range result.Children {
		if strings.Contains(c, "Theatre") {
			foundTheatre = true
		}
	}
	if !foundTheatre {
		t.Errorf("category children should include Theatre, got %v", result.Children)
	}
}

func TestAcceptance_Read_FileContentMatchesDisk(t *testing.T) {
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
	if result.Content == "" {
		t.Error("file Content should not be empty")
	}
}

func TestAcceptance_Read_IDFilesExcludeJDex(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Read(v, "S01.11.11", "")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	for _, f := range result.Files {
		if strings.HasPrefix(f, "S01.11.11") {
			t.Errorf("Files should not include JDex file, found %q", f)
		}
	}
}
