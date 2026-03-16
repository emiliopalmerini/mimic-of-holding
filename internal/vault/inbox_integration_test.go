package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestInboxIntegration_WithItems(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	items, err := Inbox(v, "")
	if err != nil {
		t.Fatalf("Inbox: %v", err)
	}

	if len(items) == 0 {
		t.Fatal("expected inbox items")
	}

	// Should find new-show-idea.md in S01.11.01 and episode-pitch.md in S02.11.01
	foundS01 := false
	foundS02 := false
	for _, item := range items {
		if item.InboxRef == "S01.11.01" && item.File == "new-show-idea.md" {
			foundS01 = true
		}
		if item.InboxRef == "S02.11.01" && item.File == "episode-pitch.md" {
			foundS02 = true
		}
	}
	if !foundS01 {
		t.Error("missing new-show-idea.md from S01.11.01")
	}
	if !foundS02 {
		t.Error("missing episode-pitch.md from S02.11.01")
	}
}

func TestInboxIntegration_ScopeFilter(t *testing.T) {
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
		if item.InboxRef[:3] != "S01" {
			t.Errorf("scope filter S01 but got item from %s", item.InboxRef)
		}
	}

	// Should not include S02 items
	for _, item := range items {
		if item.File == "episode-pitch.md" {
			t.Error("S02 item should not appear when filtering by S01")
		}
	}
}

func TestInboxIntegration_EmptyInboxes(t *testing.T) {
	root := copyFixtureVault(t)

	// Remove all inbox content files
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// First get items to know what to remove, then remove them
	items, _ := Inbox(v, "")
	for _, item := range items {
		// Find the inbox and remove the file
		for _, s := range v.Scopes {
			for _, a := range s.Areas {
				for _, c := range a.Categories {
					for _, id := range c.IDs {
						ref := fmt.Sprintf("S%02d.%02d.%02d", id.ScopeNumber, id.CategoryNum, id.Number)
						if ref == item.InboxRef {
							_ = os.Remove(filepath.Join(id.Path, item.File))
						}
					}
				}
			}
		}
	}

	// Re-parse and check
	v2, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	items2, err := Inbox(v2, "")
	if err != nil {
		t.Fatalf("Inbox: %v", err)
	}
	if len(items2) != 0 {
		t.Errorf("expected 0 items, got %d", len(items2))
	}
}

func TestInboxIntegration_ExcludesJDex(t *testing.T) {
	root := copyFixtureVault(t)

	// Create a JDex file inside an inbox
	inboxPath := filepath.Join(root, "S01 Me", "S01.10-19 Lifestyle", "S01.11 Entertainment", "S01.11.01 Inbox for S01.11")
	jdexPath := filepath.Join(inboxPath, "S01.11.01 Inbox for S01.11.md")
	_ = os.WriteFile(jdexPath, []byte("# Inbox\n"), 0o644)

	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	items, err := Inbox(v, "S01")
	if err != nil {
		t.Fatalf("Inbox: %v", err)
	}

	for _, item := range items {
		if item.File == "S01.11.01 Inbox for S01.11.md" {
			t.Error("JDex file should be excluded from inbox items")
		}
	}
}
