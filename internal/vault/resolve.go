package vault

import (
	"fmt"
	"path/filepath"
	"strconv"
)

// Resolve returns the absolute filesystem path for a JD reference.
// If file is non-empty and ref is an ID, returns the path to the file within it.
// Non-existent files are not an error; the path is returned regardless.
func Resolve(v *Vault, ref string, file string) (string, error) {
	if ref == "" {
		return "", fmt.Errorf("empty reference")
	}

	// Try ID first (most specific)
	if m := searchIDRe.FindStringSubmatch(ref); m != nil {
		scopeNum, _ := strconv.Atoi(m[1])
		catNum, _ := strconv.Atoi(m[2])
		idNum, _ := strconv.Atoi(m[3])
		id, err := findID(v, scopeNum, catNum, idNum)
		if err != nil {
			return "", err
		}
		if file != "" {
			return filepath.Join(id.Path, file), nil
		}
		return id.Path, nil
	}

	// File param only valid for IDs
	if file != "" {
		return "", fmt.Errorf("file argument is only valid for ID references")
	}

	// Category
	if m := searchCategoryRe.FindStringSubmatch(ref); m != nil {
		scopeNum, _ := strconv.Atoi(m[1])
		catNum, _ := strconv.Atoi(m[2])
		cat, err := findCategory(v, scopeNum, catNum)
		if err != nil {
			return "", err
		}
		return cat.Path, nil
	}

	// Area
	if m := searchAreaRe.FindStringSubmatch(ref); m != nil {
		scopeNum, _ := strconv.Atoi(m[1])
		rangeStart, _ := strconv.Atoi(m[2])
		rangeEnd, _ := strconv.Atoi(m[3])
		for _, s := range v.Scopes {
			if s.Number != scopeNum {
				continue
			}
			for _, a := range s.Areas {
				if a.RangeStart == rangeStart && a.RangeEnd == rangeEnd {
					return a.Path, nil
				}
			}
		}
		return "", fmt.Errorf("area S%02d.%02d-%02d not found", scopeNum, rangeStart, rangeEnd)
	}

	// Scope
	if m := filterScopeRe.FindStringSubmatch(ref); m != nil {
		scopeNum, _ := strconv.Atoi(m[1])
		scope, err := findScope(v, scopeNum)
		if err != nil {
			return "", err
		}
		return scope.Path, nil
	}

	return "", fmt.Errorf("invalid reference: %q", ref)
}
