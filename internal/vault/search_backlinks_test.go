package vault

import (
	"path/filepath"
	"strings"
	"testing"
)

// Unit tests

func TestSearch_BacklinksInvalidRef(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = Search(v, "not-a-ref", SearchOpts{Backlinks: true})
	if err == nil {
		t.Fatal("expected error for invalid ref")
	}
}

func TestSearch_BacklinksRefNotFound(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = Search(v, "S99.99.99", SearchOpts{Backlinks: true})
	if err == nil {
		t.Fatal("expected error for ref not found")
	}
}

// Integration tests

func TestSearchIntegration_BacklinksFindsLinkingNotes(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// S01.11.11 is linked from S02.11.17
	results, err := Search(v, "S01.11.11", SearchOpts{Backlinks: true})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one backlink result")
	}

	found := false
	for _, r := range results {
		if strings.Contains(r.Ref, "S02.11.17") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected backlink from S02.11.17")
	}
}

func TestSearchIntegration_BacklinksExcludesSelf(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// S02.11.17 links to S01.11.11 but also S01.11.11 links to S02.11.17
	// When querying backlinks for S02.11.17, self-links should be excluded
	results, err := Search(v, "S02.11.17", SearchOpts{Backlinks: true})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	for _, r := range results {
		if r.Ref == "S02.11.17" {
			t.Error("self-links should be excluded from backlinks")
		}
	}
}

func TestSearchIntegration_BacklinksNoLinks(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// S01.11.01 is an inbox — nothing links to it
	results, err := Search(v, "S01.11.01", SearchOpts{Backlinks: true})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearchIntegration_BacklinksScopeFilter(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// S01.11.11 is linked from S02.11.17, but scope filter S01 should exclude it
	results, err := Search(v, "S01.11.11", SearchOpts{Backlinks: true, Scope: "S01"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	for _, r := range results {
		if !strings.HasPrefix(r.Ref, "S01") {
			t.Errorf("scope filter S01 but got result from %s", r.Ref)
		}
	}
}

// Acceptance tests

func TestAcceptance_Backlinks_ResultFields(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, "S01.11.11", SearchOpts{Backlinks: true})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	for _, r := range results {
		if r.Ref == "" {
			t.Error("Ref should not be empty")
		}
		if r.Name == "" {
			t.Error("Name should not be empty")
		}
		if r.MatchLine == "" {
			t.Error("MatchLine should contain the linking line")
		}
		if r.Breadcrumb == "" {
			t.Error("Breadcrumb should not be empty")
		}
		if r.Type != "id" {
			t.Errorf("Type should be 'id', got %q", r.Type)
		}
	}
}
