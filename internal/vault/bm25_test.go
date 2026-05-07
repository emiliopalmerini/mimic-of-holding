package vault

import (
	"testing"
)

func TestTokenize(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"Hello, World!", []string{"hello", "world"}},
		{"a b c", []string{}}, // all under 2 chars
		{"foo-bar_baz", []string{"foo", "bar", "baz"}},
		{"PASTA recipe!", []string{"pasta", "recipe"}},
		{"S01.11.11 Theatre", []string{"s01", "11", "11", "theatre"}},
		{"", []string{}},
	}
	for _, c := range cases {
		got := tokenize(c.in)
		if len(got) != len(c.want) {
			t.Errorf("tokenize(%q) = %v, want %v", c.in, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("tokenize(%q)[%d] = %q, want %q", c.in, i, got[i], c.want[i])
			}
		}
	}
}

func TestBestSnippet(t *testing.T) {
	lines := []string{
		"This is a header",
		"The quick brown fox jumps over the lazy dog",
		"Another line about pasta and sauce",
	}
	got := bestSnippet(lines, []string{"pasta", "sauce"})
	want := "Another line about pasta and sauce"
	if got != want {
		t.Errorf("bestSnippet = %q, want %q", got, want)
	}
}

func TestBestSnippet_NoHits(t *testing.T) {
	lines := []string{"foo", "bar", "baz"}
	got := bestSnippet(lines, []string{"pasta"})
	if got != "" {
		t.Errorf("expected empty snippet, got %q", got)
	}
}

func TestSearchRanked_BlankQueryNoResults(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	// Whitespace tokenizes to nothing; ranked search returns empty.
	results, err := Search(v, "  ", SearchOpts{Content: true, Ranked: true})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no results, got %d", len(results))
	}
}

func TestSearchRanked_FixtureFindsTheatre(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	results, err := Search(v, "theatre", SearchOpts{Content: true, Ranked: true, Top: 5})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for 'theatre'")
	}
	// Theatre JDex must be present; ordering depends on BM25 (term frequency
	// + length normalization), so we don't assert position.
	found := false
	for _, r := range results {
		if r.Ref == "S01.11.11" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected S01.11.11 in results, got refs: %v", refsOf(results))
	}
	if results[0].Score <= 0 {
		t.Errorf("expected positive score, got %v", results[0].Score)
	}
}

func refsOf(rs []SearchResult) []string {
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = r.Ref
	}
	return out
}

func TestSearchRanked_TopLimit(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	// Use a term that should match many docs ('the' would be filtered, so use a word likely present).
	results, err := Search(v, "season", SearchOpts{Content: true, Ranked: true, Top: 1})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) > 1 {
		t.Errorf("expected at most 1 result with Top=1, got %d", len(results))
	}
}

func TestSearchRanked_DefaultTopWhenZero(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	// Should not error or return an absurd number; default 10.
	results, err := Search(v, "the", SearchOpts{Content: true, Ranked: true, Top: 0})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) > 10 {
		t.Errorf("expected default Top=10, got %d", len(results))
	}
}

func TestSearchRanked_AllTokensTooShort(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	// Single-character tokens are dropped; result should be empty.
	results, err := Search(v, "a b c", SearchOpts{Content: true, Ranked: true})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}
