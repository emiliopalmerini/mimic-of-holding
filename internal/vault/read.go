package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

// ReadResult contains the content and metadata for a JD reference at any level.
type ReadResult struct {
	Type     string   // "scope", "area", "category", "id", "file"
	Ref      string   // "S01", "S01.10-19", "S01.11", "S01.11.11"
	Name     string
	Path     string
	Content  string   // JDex content for IDs, file content for files, summary for higher levels
	Files    []string // only populated for ID-level reads
	Children []string // populated for scope/area/category (formatted as "ref name")
}

// Read returns information about any JD reference level, with optional file reading for IDs.
func Read(v *Vault, ref string, file string) (*ReadResult, error) {
	if ref == "" {
		return nil, fmt.Errorf("empty reference")
	}

	// Try each level in order from most specific to least
	if m := searchIDRe.FindStringSubmatch(ref); m != nil {
		return readID(v, m, file)
	}

	// File param only valid for IDs
	if file != "" {
		return nil, fmt.Errorf("file parameter only valid for ID references (S00.00.00)")
	}

	if m := searchCategoryRe.FindStringSubmatch(ref); m != nil {
		return readCategory(v, m)
	}
	if m := searchAreaRe.FindStringSubmatch(ref); m != nil {
		return readArea(v, m)
	}
	if m := filterScopeRe.FindStringSubmatch(ref); m != nil {
		return readScope(v, m)
	}

	return nil, fmt.Errorf("invalid reference: %q", ref)
}

func readScope(v *Vault, m []string) (*ReadResult, error) {
	num, _ := strconv.Atoi(m[1])
	for _, s := range v.Scopes {
		if s.Number == num {
			children := make([]string, len(s.Areas))
			for i, a := range s.Areas {
				children[i] = fmt.Sprintf("S%02d.%02d-%02d %s", a.ScopeNumber, a.RangeStart, a.RangeEnd, a.Name)
			}
			return &ReadResult{
				Type:     "scope",
				Ref:      fmt.Sprintf("S%02d", s.Number),
				Name:     s.Name,
				Path:     s.Path,
				Children: children,
			}, nil
		}
	}
	return nil, fmt.Errorf("scope S%02d not found", num)
}

func readArea(v *Vault, m []string) (*ReadResult, error) {
	scopeNum, _ := strconv.Atoi(m[1])
	rangeStart, _ := strconv.Atoi(m[2])
	rangeEnd, _ := strconv.Atoi(m[3])
	for _, s := range v.Scopes {
		if s.Number != scopeNum {
			continue
		}
		for _, a := range s.Areas {
			if a.RangeStart == rangeStart {
				children := make([]string, len(a.Categories))
				for i, c := range a.Categories {
					children[i] = fmt.Sprintf("S%02d.%02d %s", c.ScopeNumber, c.Number, c.Name)
				}
				return &ReadResult{
					Type:     "area",
					Ref:      fmt.Sprintf("S%02d.%02d-%02d", scopeNum, rangeStart, rangeEnd),
					Name:     a.Name,
					Path:     a.Path,
					Children: children,
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("area S%02d.%02d-%02d not found", scopeNum, rangeStart, rangeEnd)
}

func readCategory(v *Vault, m []string) (*ReadResult, error) {
	scopeNum, _ := strconv.Atoi(m[1])
	catNum, _ := strconv.Atoi(m[2])
	cat, err := findCategory(v, scopeNum, catNum)
	if err != nil {
		return nil, err
	}
	children := make([]string, len(cat.IDs))
	for i, id := range cat.IDs {
		children[i] = fmt.Sprintf("S%02d.%02d.%02d %s", id.ScopeNumber, id.CategoryNum, id.Number, id.Name)
	}
	return &ReadResult{
		Type:     "category",
		Ref:      fmt.Sprintf("S%02d.%02d", scopeNum, catNum),
		Name:     cat.Name,
		Path:     cat.Path,
		Children: children,
	}, nil
}

func readID(v *Vault, m []string, file string) (*ReadResult, error) {
	scopeNum, _ := strconv.Atoi(m[1])
	catNum, _ := strconv.Atoi(m[2])
	idNum, _ := strconv.Atoi(m[3])

	id, err := findID(v, scopeNum, catNum, idNum)
	if err != nil {
		return nil, err
	}

	ref := fmt.Sprintf("S%02d.%02d.%02d", id.ScopeNumber, id.CategoryNum, id.Number)

	// If file requested, read that specific file
	if file != "" {
		filePath := filepath.Join(id.Path, file)
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("file %q not found in %s", file, ref)
		}
		return &ReadResult{
			Type:    "file",
			Ref:     ref,
			Name:    file,
			Path:    filePath,
			Content: string(data),
		}, nil
	}

	result := &ReadResult{
		Type: "id",
		Ref:  ref,
		Name: id.Name,
		Path: id.Path,
	}

	// Read JDex file
	folderName := filepath.Base(id.Path)
	jdexName := folderName + ".md"
	jdexPath := filepath.Join(id.Path, jdexName)
	if data, err := os.ReadFile(jdexPath); err == nil {
		result.Content = string(data)
	}

	// List other files
	entries, err := os.ReadDir(id.Path)
	if err != nil {
		return result, nil
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if e.Name() == jdexName {
			continue
		}
		result.Files = append(result.Files, e.Name())
	}
	sort.Strings(result.Files)

	return result, nil
}

func findID(v *Vault, scopeNum, catNum, idNum int) (*ID, error) {
	for _, s := range v.Scopes {
		if s.Number != scopeNum {
			continue
		}
		for _, a := range s.Areas {
			for _, c := range a.Categories {
				if c.Number != catNum {
					continue
				}
				for i := range c.IDs {
					if c.IDs[i].Number == idNum {
						return &c.IDs[i], nil
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("ID S%02d.%02d.%02d not found", scopeNum, catNum, idNum)
}
