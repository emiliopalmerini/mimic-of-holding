package vault

import (
	"reflect"
	"testing"
)

func TestNormalizeLinkHead(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Foo", "Foo"},
		{"Foo|display", "Foo"},
		{"Foo#Heading", "Foo"},
		{"Foo^block", "Foo"},
		{"Foo|alias#Heading", "Foo"},
		{"folder/Foo", "Foo"},
		{"a/b/Foo.md", "Foo"},
		{"  Foo  ", "Foo"},
		{"#Heading", ""},
		{"", ""},
	}
	for _, c := range cases {
		got := normalizeLinkHead(c.in)
		if got != c.want {
			t.Errorf("normalizeLinkHead(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestExtractAliases_BlockStyle(t *testing.T) {
	lines := []string{
		"aliases:",
		"  - First",
		"  - Second",
		"location: Obsidian",
	}
	got := extractAliases(lines)
	want := []string{"First", "Second"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractAliases = %v, want %v", got, want)
	}
}

func TestExtractAliases_InlineStyle(t *testing.T) {
	lines := []string{
		`aliases: [First, "Second", 'Third']`,
		"location: Obsidian",
	}
	got := extractAliases(lines)
	want := []string{"First", "Second", "Third"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractAliases = %v, want %v", got, want)
	}
}

func TestSummarizeLintIssues_Empty(t *testing.T) {
	if got := SummarizeLintIssues(nil); got != "No issues found." {
		t.Errorf("SummarizeLintIssues(nil) = %q", got)
	}
}

func TestSummarizeLintIssues_GroupedCount(t *testing.T) {
	issues := []LintIssue{
		{Category: "missing-jdex"},
		{Category: "missing-jdex"},
		{Category: "broken-link"},
	}
	got := SummarizeLintIssues(issues)
	want := "3 issues across 2 categories: 2 missing-jdex, 1 broken-link."
	if got != want {
		t.Errorf("SummarizeLintIssues:\n got: %q\nwant: %q", got, want)
	}
}

func TestLint_RejectsInvalidScope(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	if _, err := Lint(v, "S99"); err == nil {
		t.Error("expected error for unknown scope")
	}
	if _, err := Lint(v, "garbage"); err == nil {
		t.Error("expected error for malformed scope")
	}
}
