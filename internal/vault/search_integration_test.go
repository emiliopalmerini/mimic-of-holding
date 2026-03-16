package vault

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSearchIntegration_JDRef(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, "S01.11.11", SearchOpts{})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	r := results[0]
	if r.Type != "id" || r.Name != "Theatre, 2025 Season" {
		t.Errorf("got %+v", r)
	}
}

func TestSearchIntegration_Name(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, "Episodes", SearchOpts{})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Type != "category" || results[0].Ref != "S02.11" {
		t.Errorf("got %+v", results[0])
	}
}

func TestSearchIntegration_ContentMatch(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, "Italian opera", SearchOpts{Content: true})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one content match")
	}
	// MatchLine should include filename
	for _, r := range results {
		if !strings.Contains(r.MatchLine, ":") {
			t.Errorf("MatchLine should include 'filename: line', got %q", r.MatchLine)
		}
	}
}

func TestSearchIntegration_ContentNoMatch(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, "xyzzy_nothing_matches_this", SearchOpts{Content: true})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
}

func TestSearchIntegration_ContentScopeFilter(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, "Italian", SearchOpts{Content: true, Scope: "S02"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	// The Italian opera text is in S01, so filtering by S02 should find nothing
	if len(results) != 0 {
		t.Errorf("expected 0 results with S02 scope filter, got %d", len(results))
	}
}

func TestSearchIntegration_ContentMaxLines(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// Search for something in a single file — should get at most 3 lines per file
	results, err := Search(v, "Season", SearchOpts{Content: true})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	// Count results per ref+file
	counts := make(map[string]int)
	for _, r := range results {
		key := r.Ref + "|" + strings.SplitN(r.MatchLine, ":", 2)[0]
		counts[key]++
	}
	for key, count := range counts {
		if count > 3 {
			t.Errorf("too many lines for %s: %d (max 3)", key, count)
		}
	}
}
