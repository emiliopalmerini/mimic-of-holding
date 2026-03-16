package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

var (
	scopeRe    = regexp.MustCompile(`^S(\d{2}) (.+)$`)
	areaRe     = regexp.MustCompile(`^S(\d{2})\.(\d{2})-(\d{2}) (.+)$`)
	categoryRe = regexp.MustCompile(`^S(\d{2})\.(\d{2}) (.+)$`)
	idRe       = regexp.MustCompile(`^S(\d{2})\.(\d{2})\.(\d{2}) (.+)$`)
)

func parseScopeName(name string) (*Scope, error) {
	m := scopeRe.FindStringSubmatch(name)
	if m == nil {
		return nil, fmt.Errorf("not a scope: %q", name)
	}
	num, _ := strconv.Atoi(m[1])
	return &Scope{Number: num, Name: m[2]}, nil
}

func parseAreaName(name string) (*Area, error) {
	m := areaRe.FindStringSubmatch(name)
	if m == nil {
		return nil, fmt.Errorf("not an area: %q", name)
	}
	scopeNum, _ := strconv.Atoi(m[1])
	start, _ := strconv.Atoi(m[2])
	end, _ := strconv.Atoi(m[3])
	return &Area{ScopeNumber: scopeNum, RangeStart: start, RangeEnd: end, Name: m[4]}, nil
}

func parseCategoryName(name string) (*Category, error) {
	// Reject ID format first (has three dot-separated numbers)
	if idRe.MatchString(name) {
		return nil, fmt.Errorf("not a category: %q", name)
	}
	// Reject area format (has dash-separated range)
	if areaRe.MatchString(name) {
		return nil, fmt.Errorf("not a category: %q", name)
	}
	m := categoryRe.FindStringSubmatch(name)
	if m == nil {
		return nil, fmt.Errorf("not a category: %q", name)
	}
	scopeNum, _ := strconv.Atoi(m[1])
	num, _ := strconv.Atoi(m[2])
	return &Category{ScopeNumber: scopeNum, Number: num, Name: m[3]}, nil
}

func parseIDName(name string) (*ID, error) {
	m := idRe.FindStringSubmatch(name)
	if m == nil {
		return nil, fmt.Errorf("not an ID: %q", name)
	}
	scopeNum, _ := strconv.Atoi(m[1])
	catNum, _ := strconv.Atoi(m[2])
	num, _ := strconv.Atoi(m[3])
	return &ID{
		ScopeNumber: scopeNum,
		CategoryNum: catNum,
		Number:      num,
		Name:        m[4],
		IsSystemID:  num >= 1 && num <= 9,
	}, nil
}

// ParseVault parses a JD-organized Obsidian vault at the given root path.
func ParseVault(root string) (*Vault, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("cannot access vault root: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("vault root is not a directory: %s", root)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve absolute path: %w", err)
	}

	vault := &Vault{Root: absRoot}

	scopeEntries, err := os.ReadDir(absRoot)
	if err != nil {
		return nil, fmt.Errorf("cannot read vault root: %w", err)
	}

	for _, se := range scopeEntries {
		if !se.IsDir() {
			continue
		}
		scope, err := parseScopeName(se.Name())
		if err != nil {
			continue
		}
		scope.Path = filepath.Join(absRoot, se.Name())
		scope.Areas, err = parseAreas(scope.Path)
		if err != nil {
			return nil, fmt.Errorf("parsing areas in %s: %w", se.Name(), err)
		}
		vault.Scopes = append(vault.Scopes, *scope)
	}

	sort.Slice(vault.Scopes, func(i, j int) bool {
		return vault.Scopes[i].Number < vault.Scopes[j].Number
	})

	return vault, nil
}

func parseAreas(scopePath string) ([]Area, error) {
	entries, err := os.ReadDir(scopePath)
	if err != nil {
		return nil, err
	}

	var areas []Area
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		area, err := parseAreaName(e.Name())
		if err != nil {
			continue
		}
		area.Path = filepath.Join(scopePath, e.Name())
		area.Categories, err = parseCategories(area.Path)
		if err != nil {
			return nil, fmt.Errorf("parsing categories in %s: %w", e.Name(), err)
		}
		areas = append(areas, *area)
	}

	sort.Slice(areas, func(i, j int) bool {
		return areas[i].RangeStart < areas[j].RangeStart
	})

	return areas, nil
}

func parseCategories(areaPath string) ([]Category, error) {
	entries, err := os.ReadDir(areaPath)
	if err != nil {
		return nil, err
	}

	var categories []Category
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		cat, err := parseCategoryName(e.Name())
		if err != nil {
			continue
		}
		cat.Path = filepath.Join(areaPath, e.Name())
		cat.IDs, err = parseIDs(cat.Path)
		if err != nil {
			return nil, fmt.Errorf("parsing IDs in %s: %w", e.Name(), err)
		}
		categories = append(categories, *cat)
	}

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Number < categories[j].Number
	})

	return categories, nil
}

func parseIDs(categoryPath string) ([]ID, error) {
	entries, err := os.ReadDir(categoryPath)
	if err != nil {
		return nil, err
	}

	var ids []ID
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id, err := parseIDName(e.Name())
		if err != nil {
			continue
		}
		id.Path = filepath.Join(categoryPath, e.Name())
		ids = append(ids, *id)
	}

	sort.Slice(ids, func(i, j int) bool {
		return ids[i].Number < ids[j].Number
	})

	return ids, nil
}
