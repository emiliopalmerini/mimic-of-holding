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

func TestCmd_ReadScope(t *testing.T) {
	out, _, err := executeCmd(t, "read", "S01", "--vault", testdataVault(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Lifestyle") {
		t.Errorf("expected area name in scope read:\n%s", out)
	}
}

func TestCmd_ReadFile(t *testing.T) {
	out, _, err := executeCmd(t, "read", "S01.11.11", "notes.md", "--vault", testdataVault(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "theatre season") {
		t.Errorf("expected file content in output:\n%s", out)
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

// --- Log ---

func TestCmd_LogAppendThenTail(t *testing.T) {
	root := copyTestdataVault(t)
	if _, _, err := executeCmd(t, "log", "append", "S01", "ingest", "Karpathy LLM Wiki",
		"--details", "Touched: S01.21.11", "--vault", root); err != nil {
		t.Fatalf("log append: %v", err)
	}
	out, _, err := executeCmd(t, "log", "tail", "S01", "--vault", root)
	if err != nil {
		t.Fatalf("log tail: %v", err)
	}
	if !strings.Contains(out, "ingest | Karpathy LLM Wiki") {
		t.Errorf("expected entry in tail output:\n%s", out)
	}
	if !strings.Contains(out, "Touched: S01.21.11") {
		t.Errorf("expected details in tail output:\n%s", out)
	}
}

func TestCmd_LogAppendWithSecondary(t *testing.T) {
	root := copyTestdataVault(t)
	if _, _, err := executeCmd(t, "log", "append", "S01", "archive", "S01.11.15 Theatre",
		"--secondary", "S01.11.09/[Archived] Theatre", "--vault", root); err != nil {
		t.Fatalf("log append: %v", err)
	}
	out, _, err := executeCmd(t, "log", "tail", "S01", "-n", "1", "--vault", root)
	if err != nil {
		t.Fatalf("log tail: %v", err)
	}
	if !strings.Contains(out, "→ S01.11.09/[Archived] Theatre") {
		t.Errorf("expected secondary in tail output:\n%s", out)
	}
}

func TestCmd_LogAppendRejectsBadOp(t *testing.T) {
	root := copyTestdataVault(t)
	_, _, err := executeCmd(t, "log", "append", "S01", "DELETE", "X", "--vault", root)
	if err == nil {
		t.Fatal("expected error for non-canonical op")
	}
}

func TestCmd_LogAppendMissingArgs(t *testing.T) {
	root := copyTestdataVault(t)
	_, _, err := executeCmd(t, "log", "append", "S01", "create", "--vault", root)
	if err == nil {
		t.Fatal("expected error for missing target arg")
	}
}

func TestCmd_CreateAutoLogs(t *testing.T) {
	root := copyTestdataVault(t)
	if _, _, err := executeCmd(t, "create", "S01.12", "Pasta", "--vault", root); err != nil {
		t.Fatalf("create: %v", err)
	}
	out, _, err := executeCmd(t, "log", "tail", "S01", "-n", "1", "--vault", root)
	if err != nil {
		t.Fatalf("log tail: %v", err)
	}
	if !strings.Contains(out, "create | S01.12.11 Pasta") {
		t.Errorf("expected auto-log entry, got:\n%s", out)
	}
}

func TestCmd_ArchiveAutoLogs(t *testing.T) {
	root := copyTestdataVault(t)
	if _, _, err := executeCmd(t, "archive", "S01.11.11", "--vault", root); err != nil {
		t.Fatalf("archive: %v", err)
	}
	out, _, err := executeCmd(t, "log", "tail", "S01", "-n", "1", "--vault", root)
	if err != nil {
		t.Fatalf("log tail: %v", err)
	}
	if !strings.Contains(out, "archive | S01.11.11") {
		t.Errorf("expected auto-log archive entry, got:\n%s", out)
	}
	if !strings.Contains(out, "→ [Archived]") {
		t.Errorf("expected secondary target with [Archived] prefix, got:\n%s", out)
	}
}

func TestCmd_FrontmatterAutoLogs(t *testing.T) {
	root := copyTestdataVault(t)
	if _, _, err := executeCmd(t, "frontmatter", "set", "S01.11.11", "S01.11.11 Theatre, 2025 Season.md", "location", "Notion", "--vault", root); err != nil {
		t.Fatalf("frontmatter set: %v", err)
	}
	out, _, err := executeCmd(t, "log", "tail", "S01", "-n", "1", "--vault", root)
	if err != nil {
		t.Fatalf("log tail: %v", err)
	}
	if !strings.Contains(out, "frontmatter | S01.11.11/") {
		t.Errorf("expected frontmatter auto-log entry, got:\n%s", out)
	}
	if !strings.Contains(out, "set location") {
		t.Errorf("expected details to mention 'set location', got:\n%s", out)
	}
}

func TestCmd_FailedMutationDoesNotLog(t *testing.T) {
	root := copyTestdataVault(t)
	// Archive a non-existent ref; should fail and not log.
	if _, _, err := executeCmd(t, "archive", "S01.99.99", "--vault", root); err == nil {
		t.Fatal("expected error archiving non-existent ref")
	}
	out, _, err := executeCmd(t, "log", "tail", "S01", "--vault", root)
	if err != nil {
		t.Fatalf("log tail: %v", err)
	}
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected empty log after failed mutation, got:\n%s", out)
	}
}

func TestCmd_LogTailEmptyWhenNoLog(t *testing.T) {
	root := copyTestdataVault(t)
	out, _, err := executeCmd(t, "log", "tail", "S01", "--vault", root)
	if err != nil {
		t.Fatalf("log tail: %v", err)
	}
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected empty output, got:\n%s", out)
	}
}

// --- Lint ---

func TestCmd_LintCleanFixture(t *testing.T) {
	root := copyTestdataVault(t)
	out, _, err := executeCmd(t, "lint", "--vault", root)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}
	if !strings.Contains(out, "No issues found.") {
		t.Errorf("expected clean report, got:\n%s", out)
	}
}

func TestCmd_LintReportsIssuesNonZero(t *testing.T) {
	root := copyTestdataVault(t)
	jdex := root + "/S01 Me/S01.10-19 Lifestyle/S01.11 Entertainment/S01.11.11 Theatre, 2025 Season/S01.11.11 Theatre, 2025 Season.md"
	if err := os.Remove(jdex); err != nil {
		t.Fatalf("remove JDex: %v", err)
	}
	out, _, err := executeCmd(t, "lint", "S01", "--vault", root)
	if err == nil {
		t.Fatal("expected non-zero exit on findings")
	}
	if !strings.Contains(out, "missing-jdex") {
		t.Errorf("expected report to mention missing-jdex:\n%s", out)
	}
}

func TestCmd_LintScopeAutoLogs(t *testing.T) {
	root := copyTestdataVault(t)
	if _, _, err := executeCmd(t, "lint", "S01", "--vault", root); err != nil {
		t.Fatalf("lint: %v", err)
	}
	out, _, err := executeCmd(t, "log", "tail", "S01", "-n", "1", "--vault", root)
	if err != nil {
		t.Fatalf("log tail: %v", err)
	}
	if !strings.Contains(out, "lint | S01") {
		t.Errorf("expected lint auto-log entry, got:\n%s", out)
	}
}

func TestCmd_LintWholeVaultDoesNotAutoLog(t *testing.T) {
	root := copyTestdataVault(t)
	if _, _, err := executeCmd(t, "lint", "--vault", root); err != nil {
		t.Fatalf("lint: %v", err)
	}
	out, _, err := executeCmd(t, "log", "tail", "S01", "--vault", root)
	if err != nil {
		t.Fatalf("log tail: %v", err)
	}
	if strings.TrimSpace(out) != "" {
		t.Errorf("whole-vault lint should not auto-log; got:\n%s", out)
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
