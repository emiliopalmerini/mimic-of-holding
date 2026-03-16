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
	Type      string // "scope", "area", "category", "id"
	Ref       string // "S01", "S01.10-19", "S01.11", "S01.11.11"
	Name      string
	Path      string
	MatchLine string // non-empty only for content matches
}

var (
	searchScopeRe    = regexp.MustCompile(`^S(\d{2})$`)
	searchAreaRe     = regexp.MustCompile(`^S(\d{2})\.(\d{2})-(\d{2})$`)
	searchCategoryRe = regexp.MustCompile(`^S(\d{2})\.(\d{2})$`)
	searchIDRe       = regexp.MustCompile(`^S(\d{2})\.(\d{2})\.(\d{2})$`)
)

// Search finds items in the vault matching the given query.
func Search(v *Vault, query string) ([]SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("empty search query")
	}

	// Content search
	if strings.HasPrefix(query, "?") {
		return searchContent(v, strings.TrimPrefix(query, "?"))
	}

	// JD reference search
	if results, ok := searchByRef(v, query); ok {
		return results, nil
	}

	// Name search
	return searchByName(v, query), nil
}

func searchByRef(v *Vault, query string) ([]SearchResult, bool) {
	if m := searchScopeRe.FindStringSubmatch(query); m != nil {
		num, _ := strconv.Atoi(m[1])
		for _, s := range v.Scopes {
			if s.Number == num {
				return []SearchResult{{
					Type: "scope",
					Ref:  fmt.Sprintf("S%02d", s.Number),
					Name: s.Name,
					Path: s.Path,
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
						Type: "area",
						Ref:  fmt.Sprintf("S%02d.%02d-%02d", a.ScopeNumber, a.RangeStart, a.RangeEnd),
						Name: a.Name,
						Path: a.Path,
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
							Type: "category",
							Ref:  fmt.Sprintf("S%02d.%02d", c.ScopeNumber, c.Number),
							Name: c.Name,
							Path: c.Path,
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
								Type: "id",
								Ref:  fmt.Sprintf("S%02d.%02d.%02d", id.ScopeNumber, id.CategoryNum, id.Number),
								Name: id.Name,
								Path: id.Path,
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

func searchByName(v *Vault, query string) []SearchResult {
	q := strings.ToLower(query)
	var results []SearchResult

	for _, s := range v.Scopes {
		if strings.Contains(strings.ToLower(s.Name), q) {
			results = append(results, SearchResult{
				Type: "scope",
				Ref:  fmt.Sprintf("S%02d", s.Number),
				Name: s.Name,
				Path: s.Path,
			})
		}
		for _, a := range s.Areas {
			if strings.Contains(strings.ToLower(a.Name), q) {
				results = append(results, SearchResult{
					Type: "area",
					Ref:  fmt.Sprintf("S%02d.%02d-%02d", a.ScopeNumber, a.RangeStart, a.RangeEnd),
					Name: a.Name,
					Path: a.Path,
				})
			}
			for _, c := range a.Categories {
				if strings.Contains(strings.ToLower(c.Name), q) {
					results = append(results, SearchResult{
						Type: "category",
						Ref:  fmt.Sprintf("S%02d.%02d", c.ScopeNumber, c.Number),
						Name: c.Name,
						Path: c.Path,
					})
				}
				for _, id := range c.IDs {
					if strings.Contains(strings.ToLower(id.Name), q) {
						results = append(results, SearchResult{
							Type: "id",
							Ref:  fmt.Sprintf("S%02d.%02d.%02d", id.ScopeNumber, id.CategoryNum, id.Number),
							Name: id.Name,
							Path: id.Path,
						})
					}
				}
			}
		}
	}

	return results
}

func searchContent(v *Vault, query string) ([]SearchResult, error) {
	q := strings.ToLower(query)
	var results []SearchResult

	for _, s := range v.Scopes {
		for _, a := range s.Areas {
			for _, c := range a.Categories {
				for _, id := range c.IDs {
					matches := searchFilesInDir(id.Path, q)
					for _, matchLine := range matches {
						results = append(results, SearchResult{
							Type:      "id",
							Ref:       fmt.Sprintf("S%02d.%02d.%02d", id.ScopeNumber, id.CategoryNum, id.Number),
							Name:      id.Name,
							Path:      id.Path,
							MatchLine: matchLine,
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
		matches = append(matches, lines...)
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
		}
	}

	return matches
}
