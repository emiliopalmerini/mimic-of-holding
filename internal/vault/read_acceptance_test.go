package vault

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAcceptance_Read_ResultFields(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Read(v, "S01.11.11")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if result.Ref == "" {
		t.Error("Ref should not be empty")
	}
	if result.Name == "" {
		t.Error("Name should not be empty")
	}
	if result.Path == "" {
		t.Error("Path should not be empty")
	}
	if !filepath.IsAbs(result.Path) {
		t.Errorf("Path should be absolute: %s", result.Path)
	}
}

func TestAcceptance_Read_JDexContainsExpectedContent(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Read(v, "S01.11.11")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if !strings.Contains(result.JDex, "Theatre, 2025 Season") {
		t.Errorf("JDex should contain ID name, got:\n%s", result.JDex)
	}
}

func TestAcceptance_Read_FilesExcludesJDex(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	result, err := Read(v, "S01.11.11")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	for _, f := range result.Files {
		if strings.HasPrefix(f, "S01.11.11") {
			t.Errorf("Files should not include JDex file, found %q", f)
		}
	}
}
