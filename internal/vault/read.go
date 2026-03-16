package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

// ReadResult contains the JDex entry and file listing for a JD ID.
type ReadResult struct {
	Ref   string   // "S01.11.11"
	Name  string   // "Theatre, 2025 Season"
	Path  string   // absolute path to folder
	JDex  string   // content of JDex file, empty if missing
	Files []string // other files in the folder (relative names, excluding JDex)
}

// Read returns the JDex entry and file listing for the given JD ID reference.
func Read(v *Vault, ref string) (*ReadResult, error) {
	if ref == "" {
		return nil, fmt.Errorf("empty reference")
	}

	m := searchIDRe.FindStringSubmatch(ref)
	if m == nil {
		return nil, fmt.Errorf("reference %q is not a valid ID (expected S00.00.00 format)", ref)
	}

	scopeNum, _ := strconv.Atoi(m[1])
	catNum, _ := strconv.Atoi(m[2])
	idNum, _ := strconv.Atoi(m[3])

	id, err := findID(v, scopeNum, catNum, idNum)
	if err != nil {
		return nil, err
	}

	result := &ReadResult{
		Ref:  fmt.Sprintf("S%02d.%02d.%02d", id.ScopeNumber, id.CategoryNum, id.Number),
		Name: id.Name,
		Path: id.Path,
	}

	// The JDex file is named after the folder
	folderName := filepath.Base(id.Path)
	jdexName := folderName + ".md"
	jdexPath := filepath.Join(id.Path, jdexName)

	if data, err := os.ReadFile(jdexPath); err == nil {
		result.JDex = string(data)
	}

	// List other files in the folder
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
