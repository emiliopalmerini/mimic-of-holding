package vault

import (
	"testing"
)

func TestReplaceWikiLinks_Simple(t *testing.T) {
	input := "See [[S01.11.11 Theatre, 2025 Season]] for details."
	replacements := map[string]string{
		"S01.11.11 Theatre, 2025 Season": "S01.11.11 Cinema, 2025 Season",
	}
	got, count := replaceWikiLinksInText(input, replacements)
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
	want := "See [[S01.11.11 Cinema, 2025 Season]] for details."
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestReplaceWikiLinks_Piped(t *testing.T) {
	input := "See [[S01.11.11 Theatre, 2025 Season|the theatre season]] for details."
	replacements := map[string]string{
		"S01.11.11 Theatre, 2025 Season": "S01.11.11 Cinema, 2025 Season",
	}
	got, count := replaceWikiLinksInText(input, replacements)
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
	want := "See [[S01.11.11 Cinema, 2025 Season|the theatre season]] for details."
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestReplaceWikiLinks_Multiple(t *testing.T) {
	input := "Links: [[Old Name]] and [[Old Name|display]] and [[Unrelated]]."
	replacements := map[string]string{
		"Old Name": "New Name",
	}
	got, count := replaceWikiLinksInText(input, replacements)
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
	want := "Links: [[New Name]] and [[New Name|display]] and [[Unrelated]]."
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestReplaceWikiLinks_NoMatch(t *testing.T) {
	input := "No links here, or [[Something Else]]."
	replacements := map[string]string{
		"Old Name": "New Name",
	}
	got, count := replaceWikiLinksInText(input, replacements)
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
	if got != input {
		t.Errorf("text should be unchanged")
	}
}

func TestReplaceWikiLinks_MultipleReplacements(t *testing.T) {
	input := "See [[Alpha]] and [[Beta]]."
	replacements := map[string]string{
		"Alpha": "Alpha Prime",
		"Beta":  "Beta Prime",
	}
	got, count := replaceWikiLinksInText(input, replacements)
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
	want := "See [[Alpha Prime]] and [[Beta Prime]]."
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
