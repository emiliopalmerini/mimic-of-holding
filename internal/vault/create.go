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

// CreateOpts holds optional parameters for auto-creating hierarchy.
type CreateOpts struct {
	CategoryName string // name for auto-created category (required if category doesn't exist)
	AreaName     string // name for auto-created area (required if area doesn't exist)
}

// Create creates a new JD ID in the given category with the given name.
// If template is non-empty, the named template is resolved and used as JDex content.
// customVars are substituted into the template alongside built-in variables.
// If the category or area doesn't exist, they are auto-created using names from opts.
func Create(v *Vault, categoryRef string, name string, template string, customVars map[string]string, opts ...CreateOpts) (*CreateResult, error) {
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
		var o CreateOpts
		if len(opts) > 0 {
			o = opts[0]
		}
		cat, err = ensureHierarchy(v, scopeNum, catNum, o)
		if err != nil {
			return nil, err
		}
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
		vars := templateVarsForID(ref, name)
		vars.CustomVars = customVars
		jdexContent = ApplyTemplate(tmplContent, vars)
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

func findScope(v *Vault, scopeNum int) (*Scope, error) {
	for i, s := range v.Scopes {
		if s.Number == scopeNum {
			return &v.Scopes[i], nil
		}
	}
	return nil, fmt.Errorf("scope S%02d not found", scopeNum)
}

func ensureHierarchy(v *Vault, scopeNum, catNum int, opts CreateOpts) (*Category, error) {
	scope, err := findScope(v, scopeNum)
	if err != nil {
		return nil, err
	}

	// Ensure area exists
	area := findAreaForCategory(v, scopeNum, catNum)
	if area == nil {
		if opts.AreaName == "" {
			rangeStart := (catNum / 10) * 10
			rangeEnd := rangeStart + 9
			return nil, fmt.Errorf("area S%02d.%02d-%02d not found; provide area_name to auto-create", scopeNum, rangeStart, rangeEnd)
		}
		area, err = createArea(v, scope, catNum, opts.AreaName)
		if err != nil {
			return nil, err
		}
	}

	// Ensure category exists
	cat, err := findCategory(v, scopeNum, catNum)
	if err != nil {
		if opts.CategoryName == "" {
			return nil, fmt.Errorf("category S%02d.%02d not found; provide category_name to auto-create", scopeNum, catNum)
		}
		cat, err = createCategory(v, area, scopeNum, catNum, opts.CategoryName)
		if err != nil {
			return nil, err
		}
	}

	return cat, nil
}

func createArea(v *Vault, scope *Scope, catNum int, areaName string) (*Area, error) {
	rangeStart := (catNum / 10) * 10
	rangeEnd := rangeStart + 9
	folderName := fmt.Sprintf("S%02d.%02d-%02d %s", scope.Number, rangeStart, rangeEnd, areaName)
	folderPath := filepath.Join(scope.Path, folderName)

	if err := os.MkdirAll(folderPath, 0o755); err != nil {
		return nil, fmt.Errorf("creating area folder: %w", err)
	}

	area := Area{
		ScopeNumber: scope.Number,
		RangeStart:  rangeStart,
		RangeEnd:    rangeEnd,
		Name:        areaName,
		Path:        folderPath,
	}
	scope.Areas = append(scope.Areas, area)
	return &scope.Areas[len(scope.Areas)-1], nil
}

func createCategory(v *Vault, area *Area, scopeNum, catNum int, categoryName string) (*Category, error) {
	folderName := fmt.Sprintf("S%02d.%02d %s", scopeNum, catNum, categoryName)
	folderPath := filepath.Join(area.Path, folderName)

	if err := os.MkdirAll(folderPath, 0o755); err != nil {
		return nil, fmt.Errorf("creating category folder: %w", err)
	}

	cat := Category{
		ScopeNumber: scopeNum,
		Number:      catNum,
		Name:        categoryName,
		Path:        folderPath,
	}
	area.Categories = append(area.Categories, cat)
	return &area.Categories[len(area.Categories)-1], nil
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
