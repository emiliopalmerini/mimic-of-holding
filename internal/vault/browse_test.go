package vault

import (
	"strings"
	"testing"
)

func TestBrowse_EmptyVault(t *testing.T) {
	v := &Vault{Root: "/tmp"}
	got, err := Browse(v, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestBrowse_SingleScopeTree(t *testing.T) {
	v := &Vault{
		Root: "/tmp",
		Scopes: []Scope{
			{
				Number: 1, Name: "Me",
				Areas: []Area{
					{
						ScopeNumber: 1, RangeStart: 10, RangeEnd: 19, Name: "Lifestyle",
						Categories: []Category{
							{
								ScopeNumber: 1, Number: 11, Name: "Entertainment",
								IDs: []ID{
									{ScopeNumber: 1, CategoryNum: 11, Number: 1, Name: "Inbox for S01.11", IsSystemID: true},
									{ScopeNumber: 1, CategoryNum: 11, Number: 11, Name: "Theatre, 2025 Season", IsSystemID: false},
								},
							},
						},
					},
				},
			},
		},
	}

	got, err := Browse(v, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := strings.Join([]string{
		"S01 Me",
		"  S01.10-19 Lifestyle",
		"    S01.11 Entertainment",
		"      S01.11.01 Inbox for S01.11",
		"      S01.11.11 Theatre, 2025 Season",
	}, "\n")

	if got != want {
		t.Errorf("got:\n%s\n\nwant:\n%s", got, want)
	}
}

func TestBrowse_CategoryWithNoIDs(t *testing.T) {
	v := &Vault{
		Root: "/tmp",
		Scopes: []Scope{
			{
				Number: 1, Name: "Me",
				Areas: []Area{
					{
						ScopeNumber: 1, RangeStart: 10, RangeEnd: 19, Name: "Lifestyle",
						Categories: []Category{
							{ScopeNumber: 1, Number: 12, Name: "Food"},
						},
					},
				},
			},
		},
	}

	got, err := Browse(v, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := strings.Join([]string{
		"S01 Me",
		"  S01.10-19 Lifestyle",
		"    S01.12 Food",
	}, "\n")

	if got != want {
		t.Errorf("got:\n%s\n\nwant:\n%s", got, want)
	}
}

func TestBrowse_FilterByScope(t *testing.T) {
	v := &Vault{
		Root: "/tmp",
		Scopes: []Scope{
			{Number: 1, Name: "Me", Areas: []Area{
				{ScopeNumber: 1, RangeStart: 10, RangeEnd: 19, Name: "Lifestyle"},
			}},
			{Number: 2, Name: "Due Draghi", Areas: []Area{
				{ScopeNumber: 2, RangeStart: 10, RangeEnd: 19, Name: "Due Draghi al Microfono"},
			}},
		},
	}

	got, err := Browse(v, "S01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := strings.Join([]string{
		"S01 Me",
		"  S01.10-19 Lifestyle",
	}, "\n")

	if got != want {
		t.Errorf("got:\n%s\n\nwant:\n%s", got, want)
	}
}

func TestBrowse_FilterByArea(t *testing.T) {
	v := &Vault{
		Root: "/tmp",
		Scopes: []Scope{
			{Number: 1, Name: "Me", Areas: []Area{
				{ScopeNumber: 1, RangeStart: 0, RangeEnd: 9, Name: "Management for S01"},
				{ScopeNumber: 1, RangeStart: 10, RangeEnd: 19, Name: "Lifestyle", Categories: []Category{
					{ScopeNumber: 1, Number: 11, Name: "Entertainment"},
				}},
			}},
		},
	}

	got, err := Browse(v, "S01.10-19")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := strings.Join([]string{
		"S01.10-19 Lifestyle",
		"  S01.11 Entertainment",
	}, "\n")

	if got != want {
		t.Errorf("got:\n%s\n\nwant:\n%s", got, want)
	}
}

func TestBrowse_FilterByCategory(t *testing.T) {
	v := &Vault{
		Root: "/tmp",
		Scopes: []Scope{
			{Number: 1, Name: "Me", Areas: []Area{
				{ScopeNumber: 1, RangeStart: 10, RangeEnd: 19, Name: "Lifestyle", Categories: []Category{
					{ScopeNumber: 1, Number: 11, Name: "Entertainment", IDs: []ID{
						{ScopeNumber: 1, CategoryNum: 11, Number: 1, Name: "Inbox for S01.11", IsSystemID: true},
						{ScopeNumber: 1, CategoryNum: 11, Number: 11, Name: "Theatre, 2025 Season"},
					}},
					{ScopeNumber: 1, Number: 12, Name: "Food"},
				}},
			}},
		},
	}

	got, err := Browse(v, "S01.11")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := strings.Join([]string{
		"S01.11 Entertainment",
		"  S01.11.01 Inbox for S01.11",
		"  S01.11.11 Theatre, 2025 Season",
	}, "\n")

	if got != want {
		t.Errorf("got:\n%s\n\nwant:\n%s", got, want)
	}
}

func TestBrowse_InvalidFilter(t *testing.T) {
	v := &Vault{Root: "/tmp"}
	_, err := Browse(v, "xyz")
	if err == nil {
		t.Fatal("expected error for invalid filter")
	}
}

func TestBrowse_FilterNoMatch(t *testing.T) {
	v := &Vault{
		Root: "/tmp",
		Scopes: []Scope{
			{Number: 1, Name: "Me"},
		},
	}

	_, err := Browse(v, "S99")
	if err == nil {
		t.Fatal("expected error when filter matches nothing")
	}
}
