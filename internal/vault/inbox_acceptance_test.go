package vault

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAcceptance_Inbox_ResultFields(t *testing.T) {
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
		if item.InboxRef == "" {
			t.Error("InboxRef should not be empty")
		}
		if item.InboxName == "" {
			t.Error("InboxName should not be empty")
		}
		if item.File == "" {
			t.Error("File should not be empty")
		}
	}
}

func TestAcceptance_Inbox_RefEndsWith01(t *testing.T) {
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
		if !strings.HasSuffix(item.InboxRef, ".01") {
			t.Errorf("InboxRef %q should end with .01", item.InboxRef)
		}
	}
}

func TestAcceptance_Inbox_NoFilterReturnsMultipleScopes(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	items, err := Inbox(v, "")
	if err != nil {
		t.Fatalf("Inbox: %v", err)
	}

	scopes := make(map[string]bool)
	for _, item := range items {
		scopes[item.InboxRef[:3]] = true
	}
	if len(scopes) < 2 {
		t.Errorf("expected items from multiple scopes, got %v", scopes)
	}
}
