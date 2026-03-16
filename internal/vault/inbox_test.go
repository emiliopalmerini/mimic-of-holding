package vault

import (
	"testing"
)

func TestInbox_InvalidScopeFilter(t *testing.T) {
	_, err := Inbox(searchFixture, "xyz")
	if err == nil {
		t.Fatal("expected error for invalid scope filter")
	}
}

func TestInbox_ScopeNotFound(t *testing.T) {
	_, err := Inbox(searchFixture, "S99")
	if err == nil {
		t.Fatal("expected error when scope not found")
	}
}
