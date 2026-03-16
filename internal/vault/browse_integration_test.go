package vault

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestBrowseIntegration_FullFixture(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	got, err := Browse(v, "")
	if err != nil {
		t.Fatalf("Browse: %v", err)
	}

	// Verify key lines are present
	for _, want := range []string{
		"S01 Me",
		"  S01.00-09 Management for S01",
		"    S01.01 Inbox for S01.00-09",
		"  S01.10-19 Lifestyle",
		"    S01.11 Entertainment",
		"      S01.11.01 Inbox for S01.11",
		"      S01.11.11 Theatre, 2025 Season",
		"    S01.12 Food",
		"  S01.20-29 Learning",
		"S02 Due Draghi",
		"  S02.10-19 Due Draghi al Microfono",
		"    S02.11 Episodes",
		"      S02.11.01 Inbox for S02.11",
		"      S02.11.17 Season 7 Episode 1",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing line %q\n\ngot:\n%s", want, got)
		}
	}
}

func TestBrowseIntegration_FilterScope(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	got, err := Browse(v, "S01")
	if err != nil {
		t.Fatalf("Browse: %v", err)
	}

	if !strings.Contains(got, "S01 Me") {
		t.Error("expected S01 Me in output")
	}
	if strings.Contains(got, "S02") {
		t.Error("S02 should not appear when filtering by S01")
	}
}

func TestBrowseIntegration_FilterArea(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	got, err := Browse(v, "S01.10-19")
	if err != nil {
		t.Fatalf("Browse: %v", err)
	}

	if !strings.HasPrefix(got, "S01.10-19 Lifestyle") {
		t.Errorf("expected output to start with area header, got:\n%s", got)
	}
	if strings.Contains(got, "S01.00-09") {
		t.Error("other areas should not appear")
	}
}

func TestBrowseIntegration_FilterCategory(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}

	got, err := Browse(v, "S01.11")
	if err != nil {
		t.Fatalf("Browse: %v", err)
	}

	if !strings.HasPrefix(got, "S01.11 Entertainment") {
		t.Errorf("expected output to start with category header, got:\n%s", got)
	}
	if strings.Contains(got, "S01.12") {
		t.Error("other categories should not appear")
	}
}
