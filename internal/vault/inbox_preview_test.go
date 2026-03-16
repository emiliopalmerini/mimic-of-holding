package vault

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestInboxPreviewIntegration_Populated(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	items, err := Inbox(v, "S01")
	if err != nil {
		t.Fatalf("Inbox: %v", err)
	}

	foundPreview := false
	for _, item := range items {
		if item.Preview != "" {
			foundPreview = true
		}
	}
	if !foundPreview {
		t.Error("expected at least one item with a preview")
	}
}

func TestInboxPreviewIntegration_SkipsFrontmatter(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	items, err := Inbox(v, "S01")
	if err != nil {
		t.Fatalf("Inbox: %v", err)
	}

	for _, item := range items {
		if item.File == "with-frontmatter.md" {
			if strings.Contains(item.Preview, "---") {
				t.Errorf("preview should not contain frontmatter delimiters, got %q", item.Preview)
			}
			if !strings.Contains(item.Preview, "Draft Show Idea") {
				t.Errorf("preview should contain body content, got %q", item.Preview)
			}
			return
		}
	}
	t.Error("with-frontmatter.md not found in inbox")
}

func TestAcceptance_InboxPreview_NonEmptyForContentFiles(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	items, err := Inbox(v, "")
	if err != nil {
		t.Fatalf("Inbox: %v", err)
	}

	for _, item := range items {
		if item.File == "new-show-idea.md" && item.Preview == "" {
			t.Error("preview should not be empty for a file with content")
		}
	}
}
