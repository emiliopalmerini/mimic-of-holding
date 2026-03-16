package vault

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// InboxItem represents a file sitting in an inbox folder.
type InboxItem struct {
	InboxRef  string // "S01.11.01"
	InboxName string // "Inbox for S01.11"
	File      string // filename
	Preview   string // first 3 non-empty body lines, joined with " | "
}

// Inbox lists all files in inbox folders, optionally filtered by scope.
func Inbox(v *Vault, scopeFilter string) ([]InboxItem, error) {
	scopes, err := resolveScopes(v, scopeFilter)
	if err != nil {
		return nil, err
	}

	var items []InboxItem

	for _, s := range scopes {
		for _, a := range s.Areas {
			for _, c := range a.Categories {
				for _, id := range c.IDs {
					if id.Number != 1 || !strings.Contains(strings.ToLower(id.Name), "inbox") {
						continue
					}
					files, err := listInboxFiles(id)
					if err != nil {
						continue
					}
					ref := fmt.Sprintf("S%02d.%02d.%02d", id.ScopeNumber, id.CategoryNum, id.Number)
					for _, f := range files {
						preview := filePreview(filepath.Join(id.Path, f), 3)
						items = append(items, InboxItem{
							InboxRef:  ref,
							InboxName: id.Name,
							File:      f,
							Preview:   preview,
						})
					}
				}
			}
		}
	}

	return items, nil
}

func resolveScopes(v *Vault, scopeFilter string) ([]Scope, error) {
	if scopeFilter == "" {
		return v.Scopes, nil
	}

	m := filterScopeRe.FindStringSubmatch(scopeFilter)
	if m == nil {
		return nil, fmt.Errorf("invalid scope filter: %q", scopeFilter)
	}

	num, _ := strconv.Atoi(m[1])
	for _, s := range v.Scopes {
		if s.Number == num {
			return []Scope{s}, nil
		}
	}

	return nil, fmt.Errorf("scope S%02d not found", num)
}

// filePreview returns the first n non-empty body lines of a markdown file,
// skipping YAML frontmatter. Lines are joined with " | ".
func filePreview(path string, n int) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inFrontmatter := false
	frontmatterSeen := false
	var lines []string

	for scanner.Scan() && len(lines) < n {
		line := strings.TrimSpace(scanner.Text())

		// Handle YAML frontmatter
		if line == "---" {
			if !frontmatterSeen {
				inFrontmatter = true
				frontmatterSeen = true
				continue
			}
			if inFrontmatter {
				inFrontmatter = false
				continue
			}
		}
		if inFrontmatter {
			continue
		}

		if line != "" {
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, " | ")
}

func listInboxFiles(id ID) ([]string, error) {
	entries, err := os.ReadDir(id.Path)
	if err != nil {
		return nil, err
	}

	jdexName := filepath.Base(id.Path) + ".md"

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if e.Name() == jdexName {
			continue
		}
		files = append(files, e.Name())
	}

	return files, nil
}
