package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// StatsResult holds vault-level statistics.
type StatsResult struct {
	TotalScopes       int
	TotalAreas        int
	TotalCategories   int
	TotalIDs          int
	TotalFiles        int
	EmptyCategories   []string       // refs of categories with no IDs
	OrphanIDs         []string       // refs of IDs with no inbound wiki links
	LargestCategories []CategorySize // top 5 by ID count
}

// CategorySize tracks how many IDs a category contains.
type CategorySize struct {
	Ref   string
	Name  string
	Count int
}

// Stats computes vault-level statistics.
func Stats(v *Vault) (*StatsResult, error) {
	s := &StatsResult{
		TotalScopes: len(v.Scopes),
	}

	// Collect all link targets across the vault
	allLinkTargets := collectAllLinkTargets(v)

	var catSizes []CategorySize

	for _, sc := range v.Scopes {
		s.TotalAreas += len(sc.Areas)
		for _, a := range sc.Areas {
			s.TotalCategories += len(a.Categories)
			for _, c := range a.Categories {
				catRef := fmt.Sprintf("S%02d.%02d", c.ScopeNumber, c.Number)

				if len(c.IDs) == 0 {
					s.EmptyCategories = append(s.EmptyCategories, catRef)
				}

				catSizes = append(catSizes, CategorySize{
					Ref:   catRef,
					Name:  c.Name,
					Count: len(c.IDs),
				})

				s.TotalIDs += len(c.IDs)
				for _, id := range c.IDs {
					mdFiles, _ := filepath.Glob(filepath.Join(id.Path, "*.md"))
					s.TotalFiles += len(mdFiles)

					// Check if orphan
					idRef := fmt.Sprintf("S%02d.%02d.%02d", id.ScopeNumber, id.CategoryNum, id.Number)
					folderName := filepath.Base(id.Path)
					if !allLinkTargets[folderName] && !hasAnyFileStemLinked(id.Path, allLinkTargets) {
						s.OrphanIDs = append(s.OrphanIDs, idRef)
					}
				}
			}
		}
	}

	sort.Slice(catSizes, func(i, j int) bool {
		return catSizes[i].Count > catSizes[j].Count
	})
	if len(catSizes) > 5 {
		catSizes = catSizes[:5]
	}
	s.LargestCategories = catSizes

	return s, nil
}

// collectAllLinkTargets scans all .md files in the vault and returns a set
// of all wiki link targets found.
func collectAllLinkTargets(v *Vault) map[string]bool {
	targets := map[string]bool{}

	for _, sc := range v.Scopes {
		for _, a := range sc.Areas {
			for _, c := range a.Categories {
				for _, id := range c.IDs {
					mdFiles, _ := filepath.Glob(filepath.Join(id.Path, "*.md"))
					for _, f := range mdFiles {
						data, err := os.ReadFile(f)
						if err != nil {
							continue
						}
						matches := wikiLinkRe.FindAllStringSubmatch(string(data), -1)
						for _, m := range matches {
							targets[m[1]] = true
						}
					}
				}
			}
		}
	}

	return targets
}

func hasAnyFileStemLinked(idPath string, targets map[string]bool) bool {
	mdFiles, _ := filepath.Glob(filepath.Join(idPath, "*.md"))
	for _, f := range mdFiles {
		stem := strings.TrimSuffix(filepath.Base(f), ".md")
		if targets[stem] {
			return true
		}
	}
	return false
}
