package vault

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// wikiLinkRe matches [[target]] and [[target|display]].
var wikiLinkRe = regexp.MustCompile(`\[\[([^\]|]+)(\|[^\]]+)?\]\]`)

// replaceWikiLinksInText replaces wiki link targets in text.
// Returns the updated text and count of replacements made.
func replaceWikiLinksInText(text string, replacements map[string]string) (string, int) {
	count := 0
	result := wikiLinkRe.ReplaceAllStringFunc(text, func(match string) string {
		sub := wikiLinkRe.FindStringSubmatch(match)
		if sub == nil {
			return match
		}
		target := sub[1]
		pipe := sub[2] // includes the | prefix, or empty

		if newTarget, ok := replacements[target]; ok {
			count++
			return "[[" + newTarget + pipe + "]]"
		}
		return match
	})
	return result, count
}

// UpdateWikiLinks scans all .md files under vaultRoot and replaces wiki link targets.
// replacements maps old name → new name. Returns count of links updated.
func UpdateWikiLinks(vaultRoot string, replacements map[string]string) (int, error) {
	if len(replacements) == 0 {
		return 0, nil
	}

	totalCount := 0

	err := filepath.Walk(vaultRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable
		}
		if info.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable
		}

		text := string(data)
		newText, count := replaceWikiLinksInText(text, replacements)
		if count == 0 {
			return nil
		}

		totalCount += count
		return os.WriteFile(path, []byte(newText), info.Mode())
	})

	return totalCount, err
}
