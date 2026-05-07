package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: a clean fixture should produce zero issues.
func TestAcceptance_Lint_CleanFixtureNoIssues(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	r, err := Lint(v, "")
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if len(r.Issues) != 0 {
		t.Fatalf("expected 0 issues, got %d:\n%s", len(r.Issues), FormatLintReport(r))
	}
}

func TestAcceptance_Lint_DetectsMissingJDex(t *testing.T) {
	root := copyFixtureVault(t)
	jdex := filepath.Join(root, "S01 Me", "S01.10-19 Lifestyle", "S01.11 Entertainment", "S01.11.11 Theatre, 2025 Season", "S01.11.11 Theatre, 2025 Season.md")
	if err := os.Remove(jdex); err != nil {
		t.Fatalf("remove JDex: %v", err)
	}
	v, _ := ParseVault(root)
	r, err := Lint(v, "S01")
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if !hasIssue(r.Issues, "missing-jdex", "S01.11.11 Theatre, 2025 Season") {
		t.Errorf("expected missing-jdex finding, got:\n%s", FormatLintReport(r))
	}
}

func TestAcceptance_Lint_DetectsBrokenFrontmatter(t *testing.T) {
	root := copyFixtureVault(t)
	jdex := filepath.Join(root, "S01 Me", "S01.10-19 Lifestyle", "S01.11 Entertainment", "S01.11.11 Theatre, 2025 Season", "S01.11.11 Theatre, 2025 Season.md")
	if err := os.WriteFile(jdex, []byte("# Theatre\n\nNo frontmatter at all.\n"), 0o644); err != nil {
		t.Fatalf("write JDex: %v", err)
	}
	v, _ := ParseVault(root)
	r, _ := Lint(v, "S01")
	if !hasIssueCategory(r.Issues, "broken-frontmatter") {
		t.Errorf("expected broken-frontmatter finding, got:\n%s", FormatLintReport(r))
	}
}

func TestAcceptance_Lint_DetectsNameMismatch(t *testing.T) {
	root := copyFixtureVault(t)
	folder := filepath.Join(root, "S01 Me", "S01.10-19 Lifestyle", "S01.11 Entertainment", "S01.11.11 Theatre, 2025 Season")
	bogus := filepath.Join(folder, "S01.11.99 Wrong.md")
	if err := os.WriteFile(bogus, []byte("---\naliases:\n  - x\nlocation: Obsidian\ntags:\n  - jdex\n  - index\n---\n"), 0o644); err != nil {
		t.Fatalf("write bogus: %v", err)
	}
	v, _ := ParseVault(root)
	r, _ := Lint(v, "S01")
	if !hasIssueCategory(r.Issues, "name-mismatch") {
		t.Errorf("expected name-mismatch finding, got:\n%s", FormatLintReport(r))
	}
}

func TestAcceptance_Lint_DetectsNumberingGap(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	if _, err := Create(v, "S01.12", "Pasta", "", nil); err != nil {
		t.Fatalf("Create .11: %v", err)
	}
	v, _ = ParseVault(root)
	if _, err := Create(v, "S01.12", "Pizza", "", nil); err != nil {
		t.Fatalf("Create .12: %v", err)
	}
	v, _ = ParseVault(root)
	if _, err := Create(v, "S01.12", "Sushi", "", nil); err != nil {
		t.Fatalf("Create .13: %v", err)
	}
	// Remove the .12 to create a gap
	gap := filepath.Join(root, "S01 Me", "S01.10-19 Lifestyle", "S01.12 Food", "S01.12.12 Pizza")
	if err := os.RemoveAll(gap); err != nil {
		t.Fatalf("remove .12: %v", err)
	}
	v2, _ := ParseVault(root)
	r, _ := Lint(v2, "S01")
	if !hasIssueCategory(r.Issues, "numbering-gap") {
		t.Errorf("expected numbering-gap, got:\n%s", FormatLintReport(r))
	}
}

func TestAcceptance_Lint_DetectsBrokenLink(t *testing.T) {
	root := copyFixtureVault(t)
	notes := filepath.Join(root, "S01 Me", "S01.10-19 Lifestyle", "S01.11 Entertainment", "S01.11.11 Theatre, 2025 Season", "notes.md")
	data, _ := os.ReadFile(notes)
	if err := os.WriteFile(notes, append(data, []byte("\n\nSee [[Definitely Not A Page]] too.\n")...), 0o644); err != nil {
		t.Fatalf("append link: %v", err)
	}
	v, _ := ParseVault(root)
	r, _ := Lint(v, "S01")
	if !hasIssueCategory(r.Issues, "broken-link") {
		t.Errorf("expected broken-link, got:\n%s", FormatLintReport(r))
	}
}

func TestAcceptance_Lint_DetectsOrphanID(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	// Create an ID that is not referenced by anything.
	if _, err := Create(v, "S01.12", "Lonely", "", nil); err != nil {
		t.Fatalf("Create: %v", err)
	}
	v2, _ := ParseVault(root)
	r, _ := Lint(v2, "S01")
	found := false
	for _, it := range r.Issues {
		if it.Category == "orphan-id" && strings.Contains(it.Ref, "Lonely") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected orphan-id for the Lonely ID, got:\n%s", FormatLintReport(r))
	}
}

func TestAcceptance_Lint_ScopeFilterRestrictsFindings(t *testing.T) {
	root := copyFixtureVault(t)
	// Damage S02
	jdex := filepath.Join(root, "S02 Due Draghi", "S02.10-19 Due Draghi al Microfono", "S02.11 Episodes", "S02.11.17 Season 7 Episode 1", "S02.11.17 Season 7 Episode 1.md")
	if err := os.Remove(jdex); err != nil {
		t.Fatalf("remove S02 JDex: %v", err)
	}
	v, _ := ParseVault(root)
	r, _ := Lint(v, "S01")
	for _, it := range r.Issues {
		if strings.HasPrefix(it.Ref, "S02") {
			t.Errorf("S02 finding leaked into S01-filtered lint: %+v", it)
		}
	}
}

func TestAcceptance_Lint_ReportFormat(t *testing.T) {
	r := &LintResult{
		Scope: "S01",
		Issues: []LintIssue{
			{Category: "missing-jdex", Ref: "S01.11.99 Foo", Detail: "folder has no JDex matching its name"},
			{Category: "missing-jdex", Ref: "S01.21.14 AI Patterns", Detail: "folder has no JDex matching its name"},
			{Category: "broken-frontmatter", Ref: "S01.11.11 Theatre/S01.11.11 Theatre.md", Detail: "missing aliases"},
		},
	}
	out := FormatLintReport(r)
	for _, want := range []string{
		"# Lint Report — S01",
		"## missing-jdex (2)",
		"## broken-frontmatter (1)",
		"`S01.11.99 Foo`",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("report missing %q:\n%s", want, out)
		}
	}
}

func TestAcceptance_Lint_EmptyReport(t *testing.T) {
	r := &LintResult{Scope: "S01"}
	out := FormatLintReport(r)
	if !strings.Contains(out, "No issues found.") {
		t.Errorf("expected clean message, got:\n%s", out)
	}
}

// --- helpers ---

func hasIssue(issues []LintIssue, category, refContains string) bool {
	for _, it := range issues {
		if it.Category == category && strings.Contains(it.Ref, refContains) {
			return true
		}
	}
	return false
}

func hasIssueCategory(issues []LintIssue, category string) bool {
	for _, it := range issues {
		if it.Category == category {
			return true
		}
	}
	return false
}
