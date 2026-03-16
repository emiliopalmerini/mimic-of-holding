package vault

import (
	"path/filepath"
	"strings"
	"testing"
)

// Unit tests

func TestRecent_DefaultLimit(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	// n=0 should default to 10
	results, err := Recent(v, 0, "")
	if err != nil {
		t.Fatalf("Recent: %v", err)
	}
	if len(results) > 10 {
		t.Errorf("default limit should be 10, got %d results", len(results))
	}
}

func TestRecent_InvalidScope(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	_, err = Recent(v, 5, "bad")
	if err == nil {
		t.Fatal("expected error for invalid scope")
	}
}

// Integration tests

func TestRecentIntegration_ReturnsFiles(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Recent(v, 20, "")
	if err != nil {
		t.Fatalf("Recent: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
}

func TestRecentIntegration_ScopeFilter(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Recent(v, 20, "S01")
	if err != nil {
		t.Fatalf("Recent: %v", err)
	}
	for _, r := range results {
		if !strings.HasPrefix(r.Ref, "S01") {
			t.Errorf("scope filter S01 but got result from %s", r.Ref)
		}
	}
}

func TestRecentIntegration_Limit(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Recent(v, 2, "")
	if err != nil {
		t.Fatalf("Recent: %v", err)
	}
	if len(results) > 2 {
		t.Errorf("expected at most 2 results, got %d", len(results))
	}
}

// Acceptance tests

func TestAcceptance_Recent_SortedByModTime(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Recent(v, 20, "")
	if err != nil {
		t.Fatalf("Recent: %v", err)
	}
	for i := 1; i < len(results); i++ {
		if results[i].ModTime.After(results[i-1].ModTime) {
			t.Errorf("results not sorted by ModTime descending: [%d]=%v > [%d]=%v",
				i, results[i].ModTime, i-1, results[i-1].ModTime)
		}
	}
}

func TestAcceptance_Recent_ResultFields(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	results, err := Recent(v, 20, "")
	if err != nil {
		t.Fatalf("Recent: %v", err)
	}
	for _, r := range results {
		if r.Ref == "" {
			t.Error("Ref should not be empty")
		}
		if r.File == "" {
			t.Error("File should not be empty")
		}
		if r.Breadcrumb == "" {
			t.Error("Breadcrumb should not be empty")
		}
		if r.ModTime.IsZero() {
			t.Error("ModTime should not be zero")
		}
	}
}
