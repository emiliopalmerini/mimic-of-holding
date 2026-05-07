package vault

import (
	"path/filepath"
	"strings"
	"testing"
)

// Integration: append several entries, then tail returns them in chronological order.
func TestLogIntegration_AppendThenTail(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	steps := []struct {
		op, target, secondary, details string
	}{
		{"create", "S01.21.14 AI Patterns", "", ""},
		{"frontmatter", "S01.21.14/notes.md", "", "Field: tags"},
		{"archive", "S01.11.15 Theatre", "S01.11.09/[Archived] Theatre", "4 files moved"},
		{"ingest", "Karpathy LLM Wiki", "", "Touched: S01.21.11, S01.21.14"},
	}
	for _, s := range steps {
		if err := Log(v, "S01", s.op, s.target, s.secondary, s.details); err != nil {
			t.Fatalf("Log %q: %v", s.target, err)
		}
	}

	entries, err := LogTail(v, "S01", 4)
	if err != nil {
		t.Fatalf("LogTail: %v", err)
	}
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}

	for i, want := range []string{"create | S01.21.14", "frontmatter |", "archive |", "ingest |"} {
		if !strings.Contains(entries[i], want) {
			t.Errorf("entry %d does not contain %q:\n%s", i, want, entries[i])
		}
	}
}

// Integration: re-parsing the vault after Log shows the new IDs in the JD tree.
func TestLogIntegration_NewLogIDIsReparsed(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	if err := Log(v, "S02", "create", "X", "", ""); err != nil {
		t.Fatalf("Log: %v", err)
	}

	v2, err := ParseVault(root)
	if err != nil {
		t.Fatalf("re-ParseVault: %v", err)
	}

	results, err := Search(v2, "S02.07", SearchOpts{})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("S02.07 Log ID not found after Log call")
	}
	if !strings.Contains(results[0].Path, filepath.Join("S02 Due Draghi", "S02.00-09 Management for S02", "S02.07 Log for S02.00-09")) {
		t.Errorf("unexpected path: %s", results[0].Path)
	}
}
