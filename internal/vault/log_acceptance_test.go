package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: a single Log call on a scope that already has a Management area
// produces a log.md file inside the .07 Log ID with a JD-correct hierarchy.
func TestAcceptance_Log_CreatesJDCorrectHierarchy(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	if err := Log(v, "S01", "create", "S01.21.14 AI Patterns", "", ""); err != nil {
		t.Fatalf("Log: %v", err)
	}

	logIDPath := filepath.Join(root, "S01 Me", "S01.00-09 Management for S01", "S01.07 Log for S01.00-09")
	jdexPath := filepath.Join(logIDPath, "S01.07 Log for S01.00-09.md")
	logFile := filepath.Join(logIDPath, "log.md")

	if _, err := os.Stat(logIDPath); os.IsNotExist(err) {
		t.Fatalf("log ID folder not created: %s", logIDPath)
	}
	if _, err := os.Stat(jdexPath); os.IsNotExist(err) {
		t.Errorf("JDex file not created: %s", jdexPath)
	}
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("log.md not created: %s", logFile)
	}
}

// Acceptance: scope with no Management area gets the whole chain created.
func TestAcceptance_Log_LazyCreatesManagementArea(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// S02 in fixture has no Management area
	mgmtArea := filepath.Join(root, "S02 Due Draghi", "S02.00-09 Management for S02")
	if _, err := os.Stat(mgmtArea); err == nil {
		t.Fatalf("precondition: S02 Management area should not exist in fixture")
	}

	if err := Log(v, "S02", "ingest", "Test source", "", ""); err != nil {
		t.Fatalf("Log: %v", err)
	}

	if _, err := os.Stat(mgmtArea); os.IsNotExist(err) {
		t.Errorf("Management area not created: %s", mgmtArea)
	}
	logFile := filepath.Join(mgmtArea, "S02.07 Log for S02.00-09", "log.md")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("log.md not created: %s", logFile)
	}
}

// Acceptance: JDex file has the required frontmatter fields.
func TestAcceptance_Log_JDexHasRequiredFrontmatter(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	if err := Log(v, "S01", "create", "X", "", ""); err != nil {
		t.Fatalf("Log: %v", err)
	}

	jdexPath := filepath.Join(root, "S01 Me", "S01.00-09 Management for S01", "S01.07 Log for S01.00-09", "S01.07 Log for S01.00-09.md")
	data, err := os.ReadFile(jdexPath)
	if err != nil {
		t.Fatalf("reading JDex: %v", err)
	}
	content := string(data)
	for _, want := range []string{
		"aliases:",
		"S01.07 Log for S01.00-09",
		"location: Obsidian",
		"- jdex",
		"- index",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("JDex missing %q:\n%s", want, content)
		}
	}
}

// Acceptance: log.md begins with H1 header and entries are H2 with the
// expected grammar.
func TestAcceptance_Log_EntryHeaderFormat(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	if err := Log(v, "S01", "archive", "S01.11.15 Theatre", "S01.11.09/[Archived] Theatre", "Item-level archive."); err != nil {
		t.Fatalf("Log: %v", err)
	}

	logFile := filepath.Join(root, "S01 Me", "S01.00-09 Management for S01", "S01.07 Log for S01.00-09", "log.md")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("reading log: %v", err)
	}
	content := string(data)

	if !strings.HasPrefix(content, "# Activity Log — S01\n") {
		t.Errorf("log should start with H1 'Activity Log — S01', got:\n%s", content)
	}
	if !strings.Contains(content, "archive | S01.11.15 Theatre → S01.11.09/[Archived] Theatre") {
		t.Errorf("expected header with secondary target, got:\n%s", content)
	}
	if !strings.Contains(content, "Item-level archive.") {
		t.Errorf("expected details body, got:\n%s", content)
	}
}

// Acceptance: LogTail returns the last n entries verbatim.
func TestAcceptance_LogTail_ReturnsLastN(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	for _, target := range []string{"A", "B", "C", "D"} {
		if err := Log(v, "S01", "create", target, "", ""); err != nil {
			t.Fatalf("Log %s: %v", target, err)
		}
	}

	entries, err := LogTail(v, "S01", 2)
	if err != nil {
		t.Fatalf("LogTail: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if !strings.Contains(entries[0], "create | C") {
		t.Errorf("first tail entry should be C, got: %s", entries[0])
	}
	if !strings.Contains(entries[1], "create | D") {
		t.Errorf("last tail entry should be D, got: %s", entries[1])
	}
}
