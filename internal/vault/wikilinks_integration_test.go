package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateWikiLinksIntegration_UpdatesAcrossVault(t *testing.T) {
	root := copyFixtureVault(t)

	replacements := map[string]string{
		"S01.11.11 Theatre, 2025 Season": "S01.11.11 Cinema, 2025 Season",
		"Theatre, 2025 Season":           "Cinema, 2025 Season",
	}
	count, err := UpdateWikiLinks(root, replacements)
	if err != nil {
		t.Fatalf("UpdateWikiLinks: %v", err)
	}
	if count == 0 {
		t.Error("expected at least one link updated")
	}

	// Check S02 JDex file was updated
	s02File := filepath.Join(root, "S02 Due Draghi", "S02.10-19 Due Draghi al Microfono",
		"S02.11 Episodes", "S02.11.17 Season 7 Episode 1", "S02.11.17 Season 7 Episode 1.md")
	data, err := os.ReadFile(s02File)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "Theatre, 2025 Season") {
		t.Error("old link text should be replaced")
	}
	if !strings.Contains(content, "Cinema, 2025 Season") {
		t.Error("new link text should be present")
	}
}

func TestUpdateWikiLinksIntegration_PipedLinkPreservesDisplay(t *testing.T) {
	root := copyFixtureVault(t)

	replacements := map[string]string{
		"S01.11.11 Theatre, 2025 Season": "S01.11.11 Cinema, 2025 Season",
		"Theatre, 2025 Season":           "Cinema, 2025 Season",
	}
	_, err := UpdateWikiLinks(root, replacements)
	if err != nil {
		t.Fatalf("UpdateWikiLinks: %v", err)
	}

	s02File := filepath.Join(root, "S02 Due Draghi", "S02.10-19 Due Draghi al Microfono",
		"S02.11 Episodes", "S02.11.17 Season 7 Episode 1", "S02.11.17 Season 7 Episode 1.md")
	data, _ := os.ReadFile(s02File)
	content := string(data)
	// Piped link should keep display text
	if !strings.Contains(content, "the theatre season") {
		t.Errorf("piped link display text should be preserved, got:\n%s", content)
	}
}

func TestUpdateWikiLinksIntegration_NoMatchesNoModification(t *testing.T) {
	root := copyFixtureVault(t)

	replacements := map[string]string{
		"Nonexistent Item": "Something Else",
	}
	count, err := UpdateWikiLinks(root, replacements)
	if err != nil {
		t.Fatalf("UpdateWikiLinks: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 updates, got %d", count)
	}
}
