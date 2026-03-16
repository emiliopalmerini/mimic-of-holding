package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// RecentResult represents a recently modified file.
type RecentResult struct {
	Ref        string
	Name       string
	File       string
	ModTime    time.Time
	Breadcrumb string
}

// Recent returns the N most recently modified .md files in the vault.
func Recent(v *Vault, n int, scope string) ([]RecentResult, error) {
	if n <= 0 {
		n = 10
	}

	scopes, err := filterScopes(v, scope)
	if err != nil {
		return nil, err
	}

	var results []RecentResult

	for _, s := range scopes {
		for _, a := range s.Areas {
			for _, c := range a.Categories {
				for _, id := range c.IDs {
					mdFiles, _ := filepath.Glob(filepath.Join(id.Path, "*.md"))
					for _, f := range mdFiles {
						info, err := os.Stat(f)
						if err != nil {
							continue
						}
						results = append(results, RecentResult{
							Ref:        fmt.Sprintf("S%02d.%02d.%02d", id.ScopeNumber, id.CategoryNum, id.Number),
							Name:       id.Name,
							File:       filepath.Base(f),
							ModTime:    info.ModTime(),
							Breadcrumb: idBreadcrumb(s, a, c, id),
						})
					}
				}
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].ModTime.After(results[j].ModTime)
	})

	if len(results) > n {
		results = results[:n]
	}

	return results, nil
}
