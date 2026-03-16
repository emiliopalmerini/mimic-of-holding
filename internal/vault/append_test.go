package vault

import (
	"testing"
)

func TestAppendFile_EmptyRef(t *testing.T) {
	_, err := AppendFile(searchFixture, "", "test.md", "content")
	if err == nil {
		t.Fatal("expected error for empty ref")
	}
}

func TestAppendFile_NonIDRef(t *testing.T) {
	_, err := AppendFile(searchFixture, "S01", "test.md", "content")
	if err == nil {
		t.Fatal("expected error for non-ID ref")
	}
}

func TestAppendFile_EmptyFilename(t *testing.T) {
	_, err := AppendFile(searchFixture, "S01.11.11", "", "content")
	if err == nil {
		t.Fatal("expected error for empty filename")
	}
}

func TestAppendFile_IDNotFound(t *testing.T) {
	_, err := AppendFile(searchFixture, "S01.11.99", "test.md", "content")
	if err == nil {
		t.Fatal("expected error for ID not found")
	}
}
