package vault

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSearch_MetaInvalidQuery(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = Search(v, "no-colon", SearchOpts{Meta: true})
	if err == nil {
		t.Fatal("expected error for meta query without colon")
	}
}

func TestSearchIntegration_MetaByLocation(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, "location:Obsidian", SearchOpts{Meta: true})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for location:Obsidian")
	}
	for _, r := range results {
		if !strings.Contains(strings.ToLower(r.MatchLine), "location") {
			t.Errorf("MatchLine should contain key, got %q", r.MatchLine)
		}
	}
}

func TestSearchIntegration_MetaByTag(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, "tags:jdex", SearchOpts{Meta: true})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for tags:jdex")
	}
}

func TestSearchIntegration_MetaNoMatch(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, "location:Mars", SearchOpts{Meta: true})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearchIntegration_MetaWithScopeFilter(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, "tags:jdex", SearchOpts{Meta: true, Scope: "S01"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	for _, r := range results {
		if !strings.HasPrefix(r.Ref, "S01") {
			t.Errorf("scope filter S01 but got result from %s", r.Ref)
		}
	}
}

func TestAcceptance_MetaSearch_MatchLineFormat(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, "location:Obsidian", SearchOpts{Meta: true})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	for _, r := range results {
		if !strings.Contains(r.MatchLine, ":") {
			t.Errorf("MatchLine should be 'key: value' format, got %q", r.MatchLine)
		}
	}
}
