package vault

import (
	"path/filepath"
	"testing"
)

// Unit tests

func TestStats_CategorySizeSorting(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	stats, err := Stats(v)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	for i := 1; i < len(stats.LargestCategories); i++ {
		if stats.LargestCategories[i].Count > stats.LargestCategories[i-1].Count {
			t.Errorf("LargestCategories not sorted: [%d].Count=%d > [%d].Count=%d",
				i, stats.LargestCategories[i].Count, i-1, stats.LargestCategories[i-1].Count)
		}
	}
}

// Integration tests

func TestStatsIntegration_Totals(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	stats, err := Stats(v)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats.TotalScopes != 2 {
		t.Errorf("TotalScopes: got %d, want 2", stats.TotalScopes)
	}
	if stats.TotalAreas == 0 {
		t.Error("TotalAreas should not be 0")
	}
	if stats.TotalCategories == 0 {
		t.Error("TotalCategories should not be 0")
	}
	if stats.TotalIDs == 0 {
		t.Error("TotalIDs should not be 0")
	}
	if stats.TotalFiles == 0 {
		t.Error("TotalFiles should not be 0")
	}
}

func TestStatsIntegration_EmptyCategories(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	stats, err := Stats(v)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	// S01.12 Food has no IDs (only .gitkeep)
	found := false
	for _, ref := range stats.EmptyCategories {
		if ref == "S01.12" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected S01.12 in EmptyCategories, got %v", stats.EmptyCategories)
	}
}

func TestStatsIntegration_OrphanIDs(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	stats, err := Stats(v)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}

	// S01.11.11 and S02.11.17 link to each other, so neither should be orphan
	for _, ref := range stats.OrphanIDs {
		if ref == "S01.11.11" || ref == "S02.11.17" {
			t.Errorf("%s should not be orphan (it has inbound links)", ref)
		}
	}
}

// Acceptance tests

func TestAcceptance_Stats_AllFieldsPopulated(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	stats, err := Stats(v)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats.TotalScopes == 0 {
		t.Error("TotalScopes should not be 0")
	}
	if stats.LargestCategories == nil {
		t.Error("LargestCategories should not be nil")
	}
}
