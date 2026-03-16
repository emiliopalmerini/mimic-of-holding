package vault

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// SearchResult represents a single match from a vault search.
type SearchResult struct {
	Type       string // "scope", "area", "category", "id"
	Ref        string // "S01", "S01.10-19", "S01.11", "S01.11.11"
	Name       string
	Path       string
	Breadcrumb string // human-readable hierarchy path, e.g., "S01 Me > S01.11 Entertainment > ..."
	MatchLine  string // non-empty only for content matches, format: "filename: line"
}

// SearchOpts configures search behavior.
type SearchOpts struct {
	Content   bool   // if true, search file content instead of names
	Scope     string // optional scope filter (e.g., "S01")
	Meta      bool   // if true, query is "key:value" format for frontmatter search
	Backlinks bool   // if true, query is a JD ref; returns notes linking to it
	Tags      bool   // if true, list tags (empty query) or find notes by tag (with query)
}

var (
	searchScopeRe    = regexp.MustCompile(`^S(\d{2})$`)
	searchAreaRe     = regexp.MustCompile(`^S(\d{2})\.(\d{2})-(\d{2})$`)
	searchCategoryRe = regexp.MustCompile(`^S(\d{2})\.(\d{2})$`)
	searchIDRe       = regexp.MustCompile(`^S(\d{2})\.(\d{2})\.(\d{2})$`)
)

const maxLinesPerFile = 3

func scopeBreadcrumb(s Scope) string {
	return fmt.Sprintf("S%02d %s", s.Number, s.Name)
}

func areaBreadcrumb(s Scope, a Area) string {
	return fmt.Sprintf("%s > S%02d.%02d-%02d %s", scopeBreadcrumb(s), a.ScopeNumber, a.RangeStart, a.RangeEnd, a.Name)
}

func categoryBreadcrumb(s Scope, a Area, c Category) string {
	return fmt.Sprintf("%s > S%02d.%02d %s", areaBreadcrumb(s, a), c.ScopeNumber, c.Number, c.Name)
}

func idBreadcrumb(s Scope, a Area, c Category, id ID) string {
	return fmt.Sprintf("%s > S%02d.%02d.%02d %s", categoryBreadcrumb(s, a, c), id.ScopeNumber, id.CategoryNum, id.Number, id.Name)
}

// Search finds items in the vault matching the given query.
func Search(v *Vault, query string, opts SearchOpts) ([]SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("empty search query")
	}

	// Validate scope filter
	scopes, err := filterScopes(v, opts.Scope)
	if err != nil {
		return nil, err
	}

	if opts.Tags {
		return searchTags(scopes, query)
	}

	if opts.Backlinks {
		return searchBacklinks(v, scopes, query)
	}

	if opts.Meta {
		return searchMeta(scopes, query)
	}

	if opts.Content {
		return searchContent(scopes, query)
	}

	// JD reference search (not affected by scope filter — exact match)
	if results, ok := searchByRef(v, query); ok {
		return results, nil
	}

	// Name search (affected by scope filter)
	return searchByName(scopes, query), nil
}

func filterScopes(v *Vault, scope string) ([]Scope, error) {
	if scope == "" {
		return v.Scopes, nil
	}
	m := filterScopeRe.FindStringSubmatch(scope)
	if m == nil {
		return nil, fmt.Errorf("invalid scope filter: %q", scope)
	}
	num, _ := strconv.Atoi(m[1])
	for _, s := range v.Scopes {
		if s.Number == num {
			return []Scope{s}, nil
		}
	}
	return nil, fmt.Errorf("scope S%02d not found", num)
}

func searchByRef(v *Vault, query string) ([]SearchResult, bool) {
	if m := searchScopeRe.FindStringSubmatch(query); m != nil {
		num, _ := strconv.Atoi(m[1])
		for _, s := range v.Scopes {
			if s.Number == num {
				return []SearchResult{{
					Type:       "scope",
					Ref:        fmt.Sprintf("S%02d", s.Number),
					Name:       s.Name,
					Path:       s.Path,
					Breadcrumb: scopeBreadcrumb(s),
				}}, true
			}
		}
		return []SearchResult{}, true
	}

	if m := searchAreaRe.FindStringSubmatch(query); m != nil {
		scopeNum, _ := strconv.Atoi(m[1])
		rangeStart, _ := strconv.Atoi(m[2])
		for _, s := range v.Scopes {
			if s.Number != scopeNum {
				continue
			}
			for _, a := range s.Areas {
				if a.RangeStart == rangeStart {
					return []SearchResult{{
						Type:       "area",
						Ref:        fmt.Sprintf("S%02d.%02d-%02d", a.ScopeNumber, a.RangeStart, a.RangeEnd),
						Name:       a.Name,
						Path:       a.Path,
						Breadcrumb: areaBreadcrumb(s, a),
					}}, true
				}
			}
		}
		return []SearchResult{}, true
	}

	if m := searchCategoryRe.FindStringSubmatch(query); m != nil {
		scopeNum, _ := strconv.Atoi(m[1])
		catNum, _ := strconv.Atoi(m[2])
		for _, s := range v.Scopes {
			if s.Number != scopeNum {
				continue
			}
			for _, a := range s.Areas {
				for _, c := range a.Categories {
					if c.Number == catNum {
						return []SearchResult{{
							Type:       "category",
							Ref:        fmt.Sprintf("S%02d.%02d", c.ScopeNumber, c.Number),
							Name:       c.Name,
							Path:       c.Path,
							Breadcrumb: categoryBreadcrumb(s, a, c),
						}}, true
					}
				}
			}
		}
		return []SearchResult{}, true
	}

	if m := searchIDRe.FindStringSubmatch(query); m != nil {
		scopeNum, _ := strconv.Atoi(m[1])
		catNum, _ := strconv.Atoi(m[2])
		idNum, _ := strconv.Atoi(m[3])
		for _, s := range v.Scopes {
			if s.Number != scopeNum {
				continue
			}
			for _, a := range s.Areas {
				for _, c := range a.Categories {
					if c.Number != catNum {
						continue
					}
					for _, id := range c.IDs {
						if id.Number == idNum {
							return []SearchResult{{
								Type:       "id",
								Ref:        fmt.Sprintf("S%02d.%02d.%02d", id.ScopeNumber, id.CategoryNum, id.Number),
								Name:       id.Name,
								Path:       id.Path,
								Breadcrumb: idBreadcrumb(s, a, c, id),
							}}, true
						}
					}
				}
			}
		}
		return []SearchResult{}, true
	}

	return nil, false
}

func searchByName(scopes []Scope, query string) []SearchResult {
	q := strings.ToLower(query)
	var results []SearchResult

	for _, s := range scopes {
		if strings.Contains(strings.ToLower(s.Name), q) {
			results = append(results, SearchResult{
				Type:       "scope",
				Ref:        fmt.Sprintf("S%02d", s.Number),
				Name:       s.Name,
				Path:       s.Path,
				Breadcrumb: scopeBreadcrumb(s),
			})
		}
		for _, a := range s.Areas {
			if strings.Contains(strings.ToLower(a.Name), q) {
				results = append(results, SearchResult{
					Type:       "area",
					Ref:        fmt.Sprintf("S%02d.%02d-%02d", a.ScopeNumber, a.RangeStart, a.RangeEnd),
					Name:       a.Name,
					Path:       a.Path,
					Breadcrumb: areaBreadcrumb(s, a),
				})
			}
			for _, c := range a.Categories {
				if strings.Contains(strings.ToLower(c.Name), q) {
					results = append(results, SearchResult{
						Type:       "category",
						Ref:        fmt.Sprintf("S%02d.%02d", c.ScopeNumber, c.Number),
						Name:       c.Name,
						Path:       c.Path,
						Breadcrumb: categoryBreadcrumb(s, a, c),
					})
				}
				for _, id := range c.IDs {
					if strings.Contains(strings.ToLower(id.Name), q) {
						results = append(results, SearchResult{
							Type:       "id",
							Ref:        fmt.Sprintf("S%02d.%02d.%02d", id.ScopeNumber, id.CategoryNum, id.Number),
							Name:       id.Name,
							Path:       id.Path,
							Breadcrumb: idBreadcrumb(s, a, c, id),
						})
					}
				}
			}
		}
	}

	return results
}

func searchContent(scopes []Scope, query string) ([]SearchResult, error) {
	q := strings.ToLower(query)
	var results []SearchResult

	for _, s := range scopes {
		for _, a := range s.Areas {
			for _, c := range a.Categories {
				for _, id := range c.IDs {
					matches := searchFilesInDir(id.Path, q)
					ref := fmt.Sprintf("S%02d.%02d.%02d", id.ScopeNumber, id.CategoryNum, id.Number)
					bc := idBreadcrumb(s, a, c, id)
					for _, m := range matches {
						results = append(results, SearchResult{
							Type:       "id",
							Ref:        ref,
							Name:       id.Name,
							Path:       id.Path,
							Breadcrumb: bc,
							MatchLine:  m,
						})
					}
				}
			}
		}
	}

	return results, nil
}

func searchFilesInDir(dir, query string) []string {
	var matches []string

	mdFiles, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return nil
	}

	for _, path := range mdFiles {
		lines := searchFileContent(path, query)
		filename := filepath.Base(path)
		for _, line := range lines {
			matches = append(matches, fmt.Sprintf("%s: %s", filename, line))
		}
	}

	return matches
}

func searchMeta(scopes []Scope, query string) ([]SearchResult, error) {
	// Parse key:value
	parts := strings.SplitN(query, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("meta query must be 'key:value' format, got %q", query)
	}
	key := strings.ToLower(strings.TrimSpace(parts[0]))
	value := strings.ToLower(strings.TrimSpace(parts[1]))

	var results []SearchResult

	for _, s := range scopes {
		for _, a := range s.Areas {
			for _, c := range a.Categories {
				for _, id := range c.IDs {
					ref := fmt.Sprintf("S%02d.%02d.%02d", id.ScopeNumber, id.CategoryNum, id.Number)
					bc := idBreadcrumb(s, a, c, id)

					// Scan all .md files in the ID folder
					mdFiles, _ := filepath.Glob(filepath.Join(id.Path, "*.md"))
					for _, mdFile := range mdFiles {
						matchLine := searchFrontmatter(mdFile, key, value)
						if matchLine == "" {
							continue
						}
						filename := filepath.Base(mdFile)
						results = append(results, SearchResult{
							Type:       "id",
							Ref:        ref,
							Name:       id.Name,
							Path:       id.Path,
							Breadcrumb: bc,
							MatchLine:  fmt.Sprintf("%s: %s", filename, matchLine),
						})
						break // one match per ID is enough
					}
				}
			}
		}
	}

	return results, nil
}

// searchFrontmatter reads YAML frontmatter from a file and checks if a key contains the value.
// Returns the matching "key: value" line, or empty string.
func searchFrontmatter(path, key, value string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inFrontmatter := false
	currentKey := ""

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			}
			break // end of frontmatter
		}
		if !inFrontmatter {
			continue
		}

		// Check if this is a key line (not indented list item)
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && strings.Contains(line, ":") {
			parts := strings.SplitN(trimmed, ":", 2)
			currentKey = strings.ToLower(strings.TrimSpace(parts[0]))
			val := strings.TrimSpace(parts[1])
			if currentKey == key && val != "" && strings.Contains(strings.ToLower(val), value) {
				return trimmed
			}
		} else if strings.HasPrefix(trimmed, "- ") && currentKey == key {
			// List item under the current key
			listVal := strings.TrimPrefix(trimmed, "- ")
			if strings.Contains(strings.ToLower(listVal), value) {
				return fmt.Sprintf("%s: %s", key, listVal)
			}
		}
	}

	return ""
}

var inlineTagRe = regexp.MustCompile(`(?:^|\s)#(\w+)`)

// extractTags collects tags from a markdown file (frontmatter tags: field and inline #tags).
func extractTags(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	seen := map[string]bool{}
	var tags []string
	scanner := bufio.NewScanner(f)
	inFrontmatter := false
	inTagsField := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			}
			inFrontmatter = false
			continue
		}

		if inFrontmatter {
			// Check for tags key
			if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && strings.Contains(line, ":") {
				parts := strings.SplitN(trimmed, ":", 2)
				key := strings.ToLower(strings.TrimSpace(parts[0]))
				inTagsField = key == "tags"
				if inTagsField {
					// Check inline format: tags: [a, b] or tags: a
					val := strings.TrimSpace(parts[1])
					if val != "" {
						// Handle bracket format or single value
						val = strings.Trim(val, "[]")
						for _, t := range strings.Split(val, ",") {
							t = strings.TrimSpace(t)
							if t != "" && !seen[t] {
								seen[t] = true
								tags = append(tags, t)
							}
						}
					}
				}
			} else if inTagsField && strings.HasPrefix(trimmed, "- ") {
				t := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
				if t != "" && !seen[t] {
					seen[t] = true
					tags = append(tags, t)
				}
			} else {
				inTagsField = false
			}
			continue
		}

		// Body: look for inline #tags
		matches := inlineTagRe.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			t := m[1]
			if !seen[t] {
				seen[t] = true
				tags = append(tags, t)
			}
		}
	}

	return tags
}

func searchTags(scopes []Scope, query string) ([]SearchResult, error) {
	query = strings.TrimSpace(query)
	query = strings.TrimPrefix(query, "#")

	if query == "" {
		return listAllTags(scopes)
	}
	return filterByTag(scopes, query)
}

func listAllTags(scopes []Scope) ([]SearchResult, error) {
	counts := map[string]int{}

	for _, s := range scopes {
		for _, a := range s.Areas {
			for _, c := range a.Categories {
				for _, id := range c.IDs {
					mdFiles, _ := filepath.Glob(filepath.Join(id.Path, "*.md"))
					for _, f := range mdFiles {
						for _, tag := range extractTags(f) {
							counts[tag]++
						}
					}
				}
			}
		}
	}

	if len(counts) == 0 {
		return []SearchResult{{Type: "tags", Name: "No tags found."}}, nil
	}

	// Sort by count descending
	type tagCount struct {
		tag   string
		count int
	}
	var sorted []tagCount
	for t, c := range counts {
		sorted = append(sorted, tagCount{t, c})
	}
	for i := range sorted {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].count > sorted[i].count || (sorted[j].count == sorted[i].count && sorted[j].tag < sorted[i].tag) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	var b strings.Builder
	for _, tc := range sorted {
		fmt.Fprintf(&b, "#%s (%d)\n", tc.tag, tc.count)
	}

	return []SearchResult{{
		Type: "tags",
		Name: strings.TrimSpace(b.String()),
	}}, nil
}

func filterByTag(scopes []Scope, tag string) ([]SearchResult, error) {
	tag = strings.ToLower(tag)
	var results []SearchResult

	for _, s := range scopes {
		for _, a := range s.Areas {
			for _, c := range a.Categories {
				for _, id := range c.IDs {
					if idHasTag(id, tag) {
						ref := fmt.Sprintf("S%02d.%02d.%02d", id.ScopeNumber, id.CategoryNum, id.Number)
						results = append(results, SearchResult{
							Type:       "id",
							Ref:        ref,
							Name:       id.Name,
							Path:       id.Path,
							Breadcrumb: idBreadcrumb(s, a, c, id),
						})
					}
				}
			}
		}
	}

	return results, nil
}

func idHasTag(id ID, tag string) bool {
	mdFiles, _ := filepath.Glob(filepath.Join(id.Path, "*.md"))
	for _, f := range mdFiles {
		for _, t := range extractTags(f) {
			if strings.ToLower(t) == tag {
				return true
			}
		}
	}
	return false
}

func searchBacklinks(v *Vault, scopes []Scope, query string) ([]SearchResult, error) {
	// Validate and resolve the target ref
	m := searchIDRe.FindStringSubmatch(query)
	if m == nil {
		return nil, fmt.Errorf("backlinks query must be a valid JD ID reference (S00.00.00), got %q", query)
	}
	scopeNum, _ := strconv.Atoi(m[1])
	catNum, _ := strconv.Atoi(m[2])
	idNum, _ := strconv.Atoi(m[3])

	targetID, err := findID(v, scopeNum, catNum, idNum)
	if err != nil {
		return nil, err
	}

	// Build set of link targets: the folder name and any .md file stems within it
	targetName := filepath.Base(targetID.Path)
	targets := map[string]bool{targetName: true}
	mdFiles, _ := filepath.Glob(filepath.Join(targetID.Path, "*.md"))
	for _, f := range mdFiles {
		stem := strings.TrimSuffix(filepath.Base(f), ".md")
		targets[stem] = true
	}

	var results []SearchResult

	for _, s := range scopes {
		for _, a := range s.Areas {
			for _, c := range a.Categories {
				for _, id := range c.IDs {
					// Skip the target itself
					if id.Path == targetID.Path {
						continue
					}
					ref := fmt.Sprintf("S%02d.%02d.%02d", id.ScopeNumber, id.CategoryNum, id.Number)
					bc := idBreadcrumb(s, a, c, id)

					idMdFiles, _ := filepath.Glob(filepath.Join(id.Path, "*.md"))
					for _, mdFile := range idMdFiles {
						matches := findBacklinkLines(mdFile, targets)
						for _, matchLine := range matches {
							filename := filepath.Base(mdFile)
							results = append(results, SearchResult{
								Type:       "id",
								Ref:        ref,
								Name:       id.Name,
								Path:       id.Path,
								Breadcrumb: bc,
								MatchLine:  fmt.Sprintf("%s: %s", filename, matchLine),
							})
						}
					}
				}
			}
		}
	}

	return results, nil
}

// findBacklinkLines scans a file for wiki links matching any of the targets.
func findBacklinkLines(path string, targets map[string]bool) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var matches []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		allLinks := wikiLinkRe.FindAllStringSubmatch(line, -1)
		for _, sub := range allLinks {
			if targets[sub[1]] {
				matches = append(matches, strings.TrimSpace(line))
				break
			}
		}
		if len(matches) >= maxLinesPerFile {
			break
		}
	}
	return matches
}

func searchFileContent(path, query string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var matches []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(strings.ToLower(line), query) {
			matches = append(matches, strings.TrimSpace(line))
			if len(matches) >= maxLinesPerFile {
				break
			}
		}
	}

	return matches
}
