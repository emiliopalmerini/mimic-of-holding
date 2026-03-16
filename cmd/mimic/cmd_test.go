package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func testdataVault(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test file path")
	}
	return filepath.Join(filepath.Dir(filename), "..", "..", "testdata", "vault")
}

// executeCmd runs a root command with the given args and returns stdout, stderr, and error.
func executeCmd(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := newRootCmd()
	var outBuf, errBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return outBuf.String(), errBuf.String(), err
}

// --- Browse ---

func TestCmd_Browse(t *testing.T) {
	out, _, err := executeCmd(t, "browse", "--vault", testdataVault(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "S01 Me") {
		t.Errorf("expected S01 Me in output:\n%s", out)
	}
	if !strings.Contains(out, "S02 Due Draghi") {
		t.Errorf("expected S02 Due Draghi in output:\n%s", out)
	}
}

func TestCmd_BrowseFilter(t *testing.T) {
	out, _, err := executeCmd(t, "browse", "S01", "--vault", testdataVault(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "S01 Me") {
		t.Error("expected S01 in output")
	}
	if strings.Contains(out, "S02") {
		t.Error("S02 should not appear when filtering by S01")
	}
}

// --- Search ---

func TestCmd_SearchByName(t *testing.T) {
	out, _, err := executeCmd(t, "search", "Entertainment", "--vault", testdataVault(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Entertainment") {
		t.Errorf("expected Entertainment in output:\n%s", out)
	}
}

func TestCmd_SearchByRef(t *testing.T) {
	out, _, err := executeCmd(t, "search", "S01.11.11", "--vault", testdataVault(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Theatre, 2025 Season") {
		t.Errorf("expected Theatre in output:\n%s", out)
	}
}

func TestCmd_SearchByContent(t *testing.T) {
	out, _, err := executeCmd(t, "search", "?Italian", "--vault", testdataVault(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Italian") {
		t.Errorf("expected content match in output:\n%s", out)
	}
}

func TestCmd_SearchMissingArg(t *testing.T) {
	_, _, err := executeCmd(t, "search", "--vault", testdataVault(t))
	if err == nil {
		t.Fatal("expected error for missing query")
	}
}

// --- Read ---

func TestCmd_Read(t *testing.T) {
	out, _, err := executeCmd(t, "read", "S01.11.11", "--vault", testdataVault(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Theatre, 2025 Season") {
		t.Errorf("expected JDex content in output:\n%s", out)
	}
}

func TestCmd_ReadNonID(t *testing.T) {
	_, _, err := executeCmd(t, "read", "S01", "--vault", testdataVault(t))
	if err == nil {
		t.Fatal("expected error for non-ID ref")
	}
}

func TestCmd_ReadMissingArg(t *testing.T) {
	_, _, err := executeCmd(t, "read", "--vault", testdataVault(t))
	if err == nil {
		t.Fatal("expected error for missing ref")
	}
}

// --- Create ---

func TestCmd_Create(t *testing.T) {
	root := copyTestdataVault(t)
	out, _, err := executeCmd(t, "create", "S01.12", "Pasta", "--vault", root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "S01.12.11") {
		t.Errorf("expected new ref in output:\n%s", out)
	}
}

func TestCmd_CreateMissingArgs(t *testing.T) {
	_, _, err := executeCmd(t, "create", "S01.12", "--vault", testdataVault(t))
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

// --- Archive ---

func TestCmd_Archive(t *testing.T) {
	root := copyTestdataVault(t)
	out, _, err := executeCmd(t, "archive", "S01.11.11", "--vault", root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Archived") {
		t.Errorf("expected archived confirmation in output:\n%s", out)
	}
}

func TestCmd_ArchiveMissingArg(t *testing.T) {
	_, _, err := executeCmd(t, "archive", "--vault", testdataVault(t))
	if err == nil {
		t.Fatal("expected error for missing ref")
	}
}

// --- Inbox ---

func TestCmd_Inbox(t *testing.T) {
	out, _, err := executeCmd(t, "inbox", "--vault", testdataVault(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "new-show-idea.md") {
		t.Errorf("expected inbox item in output:\n%s", out)
	}
}

func TestCmd_InboxScopeFilter(t *testing.T) {
	out, _, err := executeCmd(t, "inbox", "S01", "--vault", testdataVault(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "episode-pitch.md") {
		t.Error("S02 item should not appear when filtering by S01")
	}
}

// copyTestdataVault creates a temp copy for write tests (reuses logic from domain tests).
func copyTestdataVault(t *testing.T) string {
	t.Helper()
	src := testdataVault(t)
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
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatalf("copying fixture: %v", err)
	}
	return dst
}
