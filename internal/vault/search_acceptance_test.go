package vault

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAcceptance_Search_ResultFields(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, "Theatre", SearchOpts{})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	for _, r := range results {
		if r.Ref == "" {
			t.Error("SearchResult.Ref should not be empty")
		}
		if r.Name == "" {
			t.Error("SearchResult.Name should not be empty")
		}
		if r.Path == "" {
			t.Error("SearchResult.Path should not be empty")
		}
		validTypes := map[string]bool{"scope": true, "area": true, "category": true, "id": true}
		if !validTypes[r.Type] {
			t.Errorf("SearchResult.Type %q is not valid", r.Type)
		}
	}
}

func TestAcceptance_Search_JDRefReturnsExactlyOne(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	tests := []struct {
		query    string
		wantType string
	}{
		{"S01", "scope"},
		{"S01.10-19", "area"},
		{"S01.11", "category"},
		{"S01.11.11", "id"},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			results, err := Search(v, tt.query, SearchOpts{})
			if err != nil {
				t.Fatalf("Search: %v", err)
			}
			if len(results) != 1 {
				t.Fatalf("got %d results, want 1", len(results))
			}
			if results[0].Type != tt.wantType {
				t.Errorf("got type %q, want %q", results[0].Type, tt.wantType)
			}
		})
	}
}

func TestAcceptance_Search_ContentResultsHaveFilenameInMatchLine(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, "Shakespeare", SearchOpts{Content: true})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	for _, r := range results {
		if !strings.Contains(r.MatchLine, ".md:") {
			t.Errorf("content MatchLine should include filename, got %q", r.MatchLine)
		}
	}
}

func TestAcceptance_Search_NameResultsHaveNoMatchLine(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, "Entertainment", SearchOpts{})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	for _, r := range results {
		if r.MatchLine != "" {
			t.Errorf("name search result should not have MatchLine, got %q", r.MatchLine)
		}
	}
}

func TestAcceptance_Search_ScopeFilterExcludesOthers(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, "Shakespeare", SearchOpts{Content: true, Scope: "S01"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	for _, r := range results {
		if !strings.HasPrefix(r.Ref, "S01") {
			t.Errorf("scope filter S01 but got result from %s", r.Ref)
		}
	}
}
