package vault

import (
	"testing"
)

var searchFixture = &Vault{
	Root: "/tmp",
	Scopes: []Scope{
		{
			Number: 1, Name: "Me", Path: "/tmp/S01 Me",
			Areas: []Area{
				{
					ScopeNumber: 1, RangeStart: 0, RangeEnd: 9, Name: "Management for S01", Path: "/tmp/S01 Me/S01.00-09 Management for S01",
					Categories: []Category{
						{ScopeNumber: 1, Number: 1, Name: "Inbox for S01.00-09", Path: "/tmp/S01 Me/S01.00-09 Management for S01/S01.01 Inbox for S01.00-09"},
					},
				},
				{
					ScopeNumber: 1, RangeStart: 10, RangeEnd: 19, Name: "Lifestyle", Path: "/tmp/S01 Me/S01.10-19 Lifestyle",
					Categories: []Category{
						{
							ScopeNumber: 1, Number: 11, Name: "Entertainment", Path: "/tmp/S01 Me/S01.10-19 Lifestyle/S01.11 Entertainment",
							IDs: []ID{
								{ScopeNumber: 1, CategoryNum: 11, Number: 1, Name: "Inbox for S01.11", Path: "/tmp/S01 Me/S01.10-19 Lifestyle/S01.11 Entertainment/S01.11.01 Inbox for S01.11", IsSystemID: true},
								{ScopeNumber: 1, CategoryNum: 11, Number: 11, Name: "Theatre, 2025 Season", Path: "/tmp/S01 Me/S01.10-19 Lifestyle/S01.11 Entertainment/S01.11.11 Theatre, 2025 Season"},
							},
						},
					},
				},
			},
		},
		{
			Number: 2, Name: "Due Draghi", Path: "/tmp/S02 Due Draghi",
			Areas: []Area{
				{
					ScopeNumber: 2, RangeStart: 10, RangeEnd: 19, Name: "Due Draghi al Microfono", Path: "/tmp/S02 Due Draghi/S02.10-19 Due Draghi al Microfono",
				},
			},
		},
	},
}

func TestSearch_EmptyQuery(t *testing.T) {
	_, err := Search(searchFixture, "")
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestSearch_JDRefScope(t *testing.T) {
	results, err := Search(searchFixture, "S01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	r := results[0]
	if r.Type != "scope" || r.Ref != "S01" || r.Name != "Me" {
		t.Errorf("got %+v", r)
	}
}

func TestSearch_JDRefCategory(t *testing.T) {
	results, err := Search(searchFixture, "S01.11")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	r := results[0]
	if r.Type != "category" || r.Ref != "S01.11" || r.Name != "Entertainment" {
		t.Errorf("got %+v", r)
	}
}

func TestSearch_JDRefID(t *testing.T) {
	results, err := Search(searchFixture, "S01.11.11")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	r := results[0]
	if r.Type != "id" || r.Ref != "S01.11.11" || r.Name != "Theatre, 2025 Season" {
		t.Errorf("got %+v", r)
	}
}

func TestSearch_JDRefNoMatch(t *testing.T) {
	results, err := Search(searchFixture, "S99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
}

func TestSearch_NameExact(t *testing.T) {
	results, err := Search(searchFixture, "Entertainment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Type != "category" {
		t.Errorf("got type %q, want category", results[0].Type)
	}
}

func TestSearch_NameCaseInsensitive(t *testing.T) {
	results, err := Search(searchFixture, "management")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should match area "Management for S01" and category "Inbox for S01.00-09" (contains "management" nowhere),
	// but specifically area name and possibly others with "Management" in name
	found := false
	for _, r := range results {
		if r.Type == "area" && r.Ref == "S01.00-09" {
			found = true
		}
	}
	if !found {
		t.Error("expected area S01.00-09 Management for S01 in results")
	}
}

func TestSearch_NameMultipleLevels(t *testing.T) {
	// "Draghi" appears in scope name "Due Draghi" and area name "Due Draghi al Microfono"
	results, err := Search(searchFixture, "Draghi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	types := make(map[string]bool)
	for _, r := range results {
		types[r.Type] = true
	}
	if !types["scope"] || !types["area"] {
		t.Errorf("expected matches at scope and area level, got types: %v", types)
	}
}

func TestSearch_NameNoMatch(t *testing.T) {
	results, err := Search(searchFixture, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
}

func TestSearch_NameDoesNotIncludeMatchLine(t *testing.T) {
	results, err := Search(searchFixture, "Entertainment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if r.MatchLine != "" {
			t.Errorf("name search should not set MatchLine, got %q", r.MatchLine)
		}
	}
}
