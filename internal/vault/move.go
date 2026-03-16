package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// RenameResult contains information about a renamed JD item.
type RenameResult struct {
	Ref          string
	OldName      string
	NewName      string
	OldPath      string
	NewPath      string
	LinksUpdated int
}

// MoveResult contains information about a moved JD item.
type MoveResult struct {
	OldRef       string
	NewRef       string
	OldPath      string
	NewPath      string
	LinksUpdated int
}

// Rename changes the human-readable name of a JD item.
func Rename(v *Vault, ref string, newName string) (*RenameResult, error) {
	if ref == "" {
		return nil, fmt.Errorf("empty reference")
	}
	if newName == "" {
		return nil, fmt.Errorf("empty name")
	}

	// Determine level and find the item
	if m := searchIDRe.FindStringSubmatch(ref); m != nil {
		return renameID(v, m, newName)
	}
	if m := searchCategoryRe.FindStringSubmatch(ref); m != nil {
		return renameCategory(v, m, newName)
	}
	if m := searchAreaRe.FindStringSubmatch(ref); m != nil {
		return renameArea(v, m, newName)
	}
	if m := filterScopeRe.FindStringSubmatch(ref); m != nil {
		return renameScope(v, m, newName)
	}

	return nil, fmt.Errorf("invalid reference: %q", ref)
}

func renameID(v *Vault, m []string, newName string) (*RenameResult, error) {
	scopeNum, _ := strconv.Atoi(m[1])
	catNum, _ := strconv.Atoi(m[2])
	idNum, _ := strconv.Atoi(m[3])

	id, err := findID(v, scopeNum, catNum, idNum)
	if err != nil {
		return nil, err
	}

	oldName := id.Name
	oldPath := id.Path
	ref := fmt.Sprintf("S%02d.%02d.%02d", scopeNum, catNum, idNum)
	oldFolderName := filepath.Base(oldPath)
	newFolderName := fmt.Sprintf("%s %s", ref, newName)
	newPath := filepath.Join(filepath.Dir(oldPath), newFolderName)

	// Rename folder
	if err := os.Rename(oldPath, newPath); err != nil {
		return nil, fmt.Errorf("renaming folder: %w", err)
	}

	// Rename JDex file inside
	oldJDex := filepath.Join(newPath, oldFolderName+".md")
	newJDex := filepath.Join(newPath, newFolderName+".md")
	if _, err := os.Stat(oldJDex); err == nil {
		if err := os.Rename(oldJDex, newJDex); err != nil {
			return nil, fmt.Errorf("renaming JDex file: %w", err)
		}
		// Update JDex frontmatter
		updateJDexContent(newJDex, oldFolderName, newFolderName)
	}

	// Update wiki links
	replacements := map[string]string{
		oldFolderName: newFolderName,
		oldName:       newName,
	}
	linkCount, _ := UpdateWikiLinks(v.Root, replacements)

	return &RenameResult{
		Ref:          ref,
		OldName:      oldName,
		NewName:      newName,
		OldPath:      oldPath,
		NewPath:      newPath,
		LinksUpdated: linkCount,
	}, nil
}

func renameCategory(v *Vault, m []string, newName string) (*RenameResult, error) {
	scopeNum, _ := strconv.Atoi(m[1])
	catNum, _ := strconv.Atoi(m[2])

	cat, err := findCategory(v, scopeNum, catNum)
	if err != nil {
		return nil, err
	}

	oldName := cat.Name
	oldPath := cat.Path
	ref := fmt.Sprintf("S%02d.%02d", scopeNum, catNum)
	oldFolderName := filepath.Base(oldPath)
	newFolderName := fmt.Sprintf("%s %s", ref, newName)
	newPath := filepath.Join(filepath.Dir(oldPath), newFolderName)

	if err := os.Rename(oldPath, newPath); err != nil {
		return nil, fmt.Errorf("renaming folder: %w", err)
	}

	replacements := map[string]string{
		oldFolderName: newFolderName,
		oldName:       newName,
	}
	linkCount, _ := UpdateWikiLinks(v.Root, replacements)

	return &RenameResult{
		Ref: ref, OldName: oldName, NewName: newName,
		OldPath: oldPath, NewPath: newPath, LinksUpdated: linkCount,
	}, nil
}

func renameArea(v *Vault, m []string, newName string) (*RenameResult, error) {
	scopeNum, _ := strconv.Atoi(m[1])
	rangeStart, _ := strconv.Atoi(m[2])
	rangeEnd, _ := strconv.Atoi(m[3])

	for _, s := range v.Scopes {
		if s.Number != scopeNum {
			continue
		}
		for _, a := range s.Areas {
			if a.RangeStart != rangeStart {
				continue
			}
			oldName := a.Name
			oldPath := a.Path
			ref := fmt.Sprintf("S%02d.%02d-%02d", scopeNum, rangeStart, rangeEnd)
			oldFolderName := filepath.Base(oldPath)
			newFolderName := fmt.Sprintf("%s %s", ref, newName)
			newPath := filepath.Join(filepath.Dir(oldPath), newFolderName)

			if err := os.Rename(oldPath, newPath); err != nil {
				return nil, fmt.Errorf("renaming folder: %w", err)
			}

			replacements := map[string]string{oldFolderName: newFolderName, oldName: newName}
			linkCount, _ := UpdateWikiLinks(v.Root, replacements)

			return &RenameResult{
				Ref: ref, OldName: oldName, NewName: newName,
				OldPath: oldPath, NewPath: newPath, LinksUpdated: linkCount,
			}, nil
		}
	}
	return nil, fmt.Errorf("area not found")
}

func renameScope(v *Vault, m []string, newName string) (*RenameResult, error) {
	num, _ := strconv.Atoi(m[1])
	for _, s := range v.Scopes {
		if s.Number != num {
			continue
		}
		oldName := s.Name
		oldPath := s.Path
		ref := fmt.Sprintf("S%02d", num)
		oldFolderName := filepath.Base(oldPath)
		newFolderName := fmt.Sprintf("%s %s", ref, newName)
		newPath := filepath.Join(filepath.Dir(oldPath), newFolderName)

		if err := os.Rename(oldPath, newPath); err != nil {
			return nil, fmt.Errorf("renaming folder: %w", err)
		}

		replacements := map[string]string{oldFolderName: newFolderName, oldName: newName}
		linkCount, _ := UpdateWikiLinks(v.Root, replacements)

		return &RenameResult{
			Ref: ref, OldName: oldName, NewName: newName,
			OldPath: oldPath, NewPath: newPath, LinksUpdated: linkCount,
		}, nil
	}
	return nil, fmt.Errorf("scope S%02d not found", num)
}

// Move relocates a JD item to a different parent.
func Move(v *Vault, ref string, to string) (*MoveResult, error) {
	if ref == "" {
		return nil, fmt.Errorf("empty reference")
	}
	if to == "" {
		return nil, fmt.Errorf("empty target")
	}

	if m := searchIDRe.FindStringSubmatch(ref); m != nil {
		return moveID(v, m, to)
	}
	if m := searchCategoryRe.FindStringSubmatch(ref); m != nil {
		return moveCategory(v, m, to)
	}

	return nil, fmt.Errorf("invalid reference for move: %q (only IDs and categories can be moved)", ref)
}

func moveID(v *Vault, m []string, to string) (*MoveResult, error) {
	scopeNum, _ := strconv.Atoi(m[1])
	catNum, _ := strconv.Atoi(m[2])
	idNum, _ := strconv.Atoi(m[3])

	id, err := findID(v, scopeNum, catNum, idNum)
	if err != nil {
		return nil, err
	}

	// Target must be a category
	tm := searchCategoryRe.FindStringSubmatch(to)
	if tm == nil {
		return nil, fmt.Errorf("invalid target category: %q", to)
	}
	tScopeNum, _ := strconv.Atoi(tm[1])
	tCatNum, _ := strconv.Atoi(tm[2])

	targetCat, err := findCategory(v, tScopeNum, tCatNum)
	if err != nil {
		return nil, fmt.Errorf("target category not found: %w", err)
	}

	newNum := nextRegularID(targetCat)
	oldRef := fmt.Sprintf("S%02d.%02d.%02d", scopeNum, catNum, idNum)
	newRef := fmt.Sprintf("S%02d.%02d.%02d", tScopeNum, tCatNum, newNum)
	newFolderName := fmt.Sprintf("%s %s", newRef, id.Name)
	newPath := filepath.Join(targetCat.Path, newFolderName)
	oldPath := id.Path
	oldFolderName := filepath.Base(oldPath)

	if err := os.Rename(oldPath, newPath); err != nil {
		return nil, fmt.Errorf("moving folder: %w", err)
	}

	// Rename JDex file and update content
	oldJDex := filepath.Join(newPath, oldFolderName+".md")
	newJDex := filepath.Join(newPath, newFolderName+".md")
	if _, err := os.Stat(oldJDex); err == nil {
		os.Rename(oldJDex, newJDex)
		updateJDexContent(newJDex, oldFolderName, newFolderName)
	}

	// Update wiki links
	replacements := map[string]string{oldFolderName: newFolderName}
	linkCount, _ := UpdateWikiLinks(v.Root, replacements)

	return &MoveResult{
		OldRef: oldRef, NewRef: newRef,
		OldPath: oldPath, NewPath: newPath,
		LinksUpdated: linkCount,
	}, nil
}

func moveCategory(v *Vault, m []string, to string) (*MoveResult, error) {
	scopeNum, _ := strconv.Atoi(m[1])
	catNum, _ := strconv.Atoi(m[2])

	cat, err := findCategory(v, scopeNum, catNum)
	if err != nil {
		return nil, err
	}

	// Target must be an area
	tm := searchAreaRe.FindStringSubmatch(to)
	if tm == nil {
		return nil, fmt.Errorf("invalid target area: %q", to)
	}
	tScopeNum, _ := strconv.Atoi(tm[1])
	tRangeStart, _ := strconv.Atoi(tm[2])
	tRangeEnd, _ := strconv.Atoi(tm[3])

	// Find target area
	var targetArea *Area
	for _, s := range v.Scopes {
		if s.Number != tScopeNum {
			continue
		}
		for i, a := range s.Areas {
			if a.RangeStart == tRangeStart {
				targetArea = &s.Areas[i]
				break
			}
		}
	}
	if targetArea == nil {
		return nil, fmt.Errorf("target area S%02d.%02d-%02d not found", tScopeNum, tRangeStart, tRangeEnd)
	}

	// Assign new category number (first available in target range after management)
	newCatNum := nextCategoryNum(targetArea, tRangeStart, tRangeEnd)

	oldRef := fmt.Sprintf("S%02d.%02d", scopeNum, catNum)
	newRef := fmt.Sprintf("S%02d.%02d", tScopeNum, newCatNum)
	oldFolderName := filepath.Base(cat.Path)
	newFolderName := fmt.Sprintf("%s %s", newRef, cat.Name)
	newPath := filepath.Join(targetArea.Path, newFolderName)

	if err := os.Rename(cat.Path, newPath); err != nil {
		return nil, fmt.Errorf("moving folder: %w", err)
	}

	// Recursively update child ID prefixes
	replacements := map[string]string{oldFolderName: newFolderName}
	updateChildPrefixes(newPath, oldRef, newRef, replacements)

	linkCount, _ := UpdateWikiLinks(v.Root, replacements)

	return &MoveResult{
		OldRef: oldRef, NewRef: newRef,
		OldPath: cat.Path, NewPath: newPath,
		LinksUpdated: linkCount,
	}, nil
}

func nextCategoryNum(area *Area, rangeStart, rangeEnd int) int {
	used := make(map[int]bool)
	for _, c := range area.Categories {
		used[c.Number] = true
	}
	// Start after management (.X0)
	mgmt := rangeStart
	for i := mgmt + 1; i <= rangeEnd; i++ {
		if !used[i] {
			return i
		}
	}
	return rangeEnd // fallback
}

func updateChildPrefixes(parentPath, oldPrefix, newPrefix string, replacements map[string]string) {
	entries, err := os.ReadDir(parentPath)
	if err != nil {
		return
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, oldPrefix) {
			continue
		}
		newName := strings.Replace(name, oldPrefix, newPrefix, 1)
		oldPath := filepath.Join(parentPath, name)
		newChildPath := filepath.Join(parentPath, newName)

		os.Rename(oldPath, newChildPath)
		replacements[name] = newName

		// Rename JDex file inside
		oldJDex := filepath.Join(newChildPath, name+".md")
		newJDex := filepath.Join(newChildPath, newName+".md")
		if _, err := os.Stat(oldJDex); err == nil {
			os.Rename(oldJDex, newJDex)
			updateJDexContent(newJDex, name, newName)
		}
	}
}

// MoveFile moves a file from one ID to another.
func MoveFile(v *Vault, fromRef string, filename string, toRef string) (string, error) {
	if fromRef == "" || filename == "" || toRef == "" {
		return "", fmt.Errorf("fromRef, filename, and toRef are all required")
	}

	fromM := searchIDRe.FindStringSubmatch(fromRef)
	if fromM == nil {
		return "", fmt.Errorf("invalid source ID: %q", fromRef)
	}
	toM := searchIDRe.FindStringSubmatch(toRef)
	if toM == nil {
		return "", fmt.Errorf("invalid target ID: %q", toRef)
	}

	fromScopeNum, _ := strconv.Atoi(fromM[1])
	fromCatNum, _ := strconv.Atoi(fromM[2])
	fromIDNum, _ := strconv.Atoi(fromM[3])
	fromID, err := findID(v, fromScopeNum, fromCatNum, fromIDNum)
	if err != nil {
		return "", err
	}

	toScopeNum, _ := strconv.Atoi(toM[1])
	toCatNum, _ := strconv.Atoi(toM[2])
	toIDNum, _ := strconv.Atoi(toM[3])
	toID, err := findID(v, toScopeNum, toCatNum, toIDNum)
	if err != nil {
		return "", err
	}

	srcPath := filepath.Join(fromID.Path, filename)
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file %q not found in %s", filename, fromRef)
	}

	dstPath := filepath.Join(toID.Path, filename)
	if err := os.Rename(srcPath, dstPath); err != nil {
		return "", fmt.Errorf("moving file: %w", err)
	}

	return dstPath, nil
}

func updateJDexContent(path, oldName, newName string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	content := strings.ReplaceAll(string(data), oldName, newName)
	os.WriteFile(path, []byte(content), 0o644)
}
