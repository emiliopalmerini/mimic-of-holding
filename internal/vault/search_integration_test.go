package vault

import (
	"path/filepath"
	"testing"
)

func TestSearchIntegration_JDRef(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, "S01.11.11")
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

	results, err := Search(v, "Episodes")
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

	results, err := Search(v, "?Italian opera")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one content match")
	}
	found := false
	for _, r := range results {
		if r.MatchLine != "" {
			found = true
		}
	}
	if !found {
		t.Error("content search results should have MatchLine set")
	}
}

func TestSearchIntegration_ContentNoMatch(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, "?xyzzy_nothing_matches_this")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
}
