package vault

import (
	"testing"
)

func TestWriteFile_EmptyRef(t *testing.T) {
	_, err := WriteFile(searchFixture, "", "test.md", "content")
	if err == nil {
		t.Fatal("expected error for empty ref")
	}
}

func TestWriteFile_NonIDRef(t *testing.T) {
	_, err := WriteFile(searchFixture, "S01", "test.md", "content")
	if err == nil {
		t.Fatal("expected error for non-ID ref")
	}
}

func TestWriteFile_EmptyFilename(t *testing.T) {
	_, err := WriteFile(searchFixture, "S01.11.11", "", "content")
	if err == nil {
		t.Fatal("expected error for empty filename")
	}
}

func TestWriteFile_IDNotFound(t *testing.T) {
	_, err := WriteFile(searchFixture, "S01.11.99", "test.md", "content")
	if err == nil {
		t.Fatal("expected error for ID not found")
	}
}
