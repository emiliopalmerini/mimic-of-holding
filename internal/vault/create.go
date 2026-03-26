package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// CreateResult contains information about a newly created JD ID.
type CreateResult struct {
	Ref  string // "S01.11.12"
	Name string // "Cinema"
	Path string // absolute path to created folder
}

// Create creates a new JD ID in the given category with the given name.
// If template is non-empty, the named template is resolved and used as JDex content.
func Create(v *Vault, categoryRef string, name string, template string) (*CreateResult, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	m := searchCategoryRe.FindStringSubmatch(categoryRef)
	if m == nil {
		return nil, fmt.Errorf("invalid category reference: %q (expected S00.00 format)", categoryRef)
	}

	scopeNum, _ := strconv.Atoi(m[1])
	catNum, _ := strconv.Atoi(m[2])

	cat, err := findCategory(v, scopeNum, catNum)
	if err != nil {
		return nil, err
	}

	nextNum := nextRegularID(cat)
	ref := fmt.Sprintf("S%02d.%02d.%02d", scopeNum, catNum, nextNum)
	folderName := fmt.Sprintf("%s %s", ref, name)
	folderPath := filepath.Join(cat.Path, folderName)

	// Resolve template before creating folder (fail early)
	var jdexContent string
	if template != "" {
		tmplContent, err := resolveTemplate(v, scopeNum, catNum, template)
		if err != nil {
			return nil, err
		}
		jdexContent = ApplyTemplate(tmplContent, templateVarsForID(ref, name))
	} else {
		jdexContent = fmt.Sprintf(`---
aliases:
  - %s %s
location: Obsidian
tags:
  - jdex
  - index
---
# %s %s

## Contents
`, ref, name, ref, name)
	}

	if err := os.MkdirAll(folderPath, 0o755); err != nil {
		return nil, fmt.Errorf("creating folder: %w", err)
	}

	jdexPath := filepath.Join(folderPath, folderName+".md")
	if err := os.WriteFile(jdexPath, []byte(jdexContent), 0o644); err != nil {
		return nil, fmt.Errorf("writing JDex file: %w", err)
	}

	return &CreateResult{
		Ref:  ref,
		Name: name,
		Path: folderPath,
	}, nil
}

func findCategory(v *Vault, scopeNum, catNum int) (*Category, error) {
	for _, s := range v.Scopes {
		if s.Number != scopeNum {
			continue
		}
		for _, a := range s.Areas {
			for i := range a.Categories {
				if a.Categories[i].Number == catNum {
					return &a.Categories[i], nil
				}
			}
		}
	}
	return nil, fmt.Errorf("category S%02d.%02d not found", scopeNum, catNum)
}

func nextRegularID(cat *Category) int {
	max := 10 // so first regular ID will be 11
	for _, id := range cat.IDs {
		if id.Number >= 10 && id.Number > max {
			max = id.Number
		}
	}
	return max + 1
}
