package vault

import (
	"path/filepath"
	"strings"
	"testing"
)

// Unit tests

func TestSearch_TagsQueryNormalizesHash(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// Both "#jdex" and "jdex" should return the same results
	r1, err := Search(v, "#jdex", SearchOpts{Tags: true})
	if err != nil {
		t.Fatalf("Search #jdex: %v", err)
	}
	r2, err := Search(v, "jdex", SearchOpts{Tags: true})
	if err != nil {
		t.Fatalf("Search jdex: %v", err)
	}
	if len(r1) != len(r2) {
		t.Errorf("#jdex returned %d results, jdex returned %d", len(r1), len(r2))
	}
}

// Integration tests

func TestSearchIntegration_TagsListAll(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// Empty query with Tags=true lists all tags
	results, err := Search(v, " ", SearchOpts{Tags: true})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result (tag listing), got %d", len(results))
	}
	if results[0].Type != "tags" {
		t.Errorf("expected type 'tags', got %q", results[0].Type)
	}
	// Should contain known tags
	for _, tag := range []string{"jdex", "index", "draft"} {
		if !strings.Contains(results[0].Name, tag) {
			t.Errorf("tag listing should contain %q, got:\n%s", tag, results[0].Name)
		}
	}
}

func TestSearchIntegration_TagsFilterByTag(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, "jdex", SearchOpts{Tags: true})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results for tag 'jdex'")
	}
	// Both S01.11.11 and S02.11.17 have tags: [jdex]
	refs := make(map[string]bool)
	for _, r := range results {
		refs[r.Ref] = true
	}
	if !refs["S01.11.11"] {
		t.Error("expected S01.11.11 in results")
	}
	if !refs["S02.11.17"] {
		t.Error("expected S02.11.17 in results")
	}
}

func TestSearchIntegration_TagsFilterNoMatch(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, "nonexistent-tag", SearchOpts{Tags: true})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearchIntegration_TagsScopeFilter(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, "jdex", SearchOpts{Tags: true, Scope: "S01"})
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

func TestAcceptance_TagsListing_SortedByCounts(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, " ", SearchOpts{Tags: true})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	// Each line should have format "#tag (N)"
	lines := strings.Split(strings.TrimSpace(results[0].Name), "\n")
	if len(lines) == 0 {
		t.Fatal("tag listing should not be empty")
	}
	for _, line := range lines {
		if !strings.HasPrefix(line, "#") || !strings.Contains(line, "(") {
			t.Errorf("unexpected tag listing format: %q", line)
		}
	}
}

func TestAcceptance_TagsFilter_ResultFields(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Search(v, "jdex", SearchOpts{Tags: true})
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
		if r.Breadcrumb == "" {
			t.Error("Breadcrumb should not be empty")
		}
		if r.Type != "id" {
			t.Errorf("Type should be 'id', got %q", r.Type)
		}
	}
}
