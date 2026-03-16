package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// ArchiveResult contains information about an archived JD item.
type ArchiveResult struct {
	Ref     string // original ref
	NewPath string // path after archiving
}

// Archive moves a JD item to its parent's archive folder.
func Archive(v *Vault, ref string) (*ArchiveResult, error) {
	if ref == "" {
		return nil, fmt.Errorf("empty reference")
	}

	// Try as ID first
	if m := searchIDRe.FindStringSubmatch(ref); m != nil {
		return archiveID(v, m)
	}

	// Try as category
	if m := searchCategoryRe.FindStringSubmatch(ref); m != nil {
		return archiveCategory(v, m)
	}

	// Scope or area or invalid
	if searchScopeRe.MatchString(ref) {
		return nil, fmt.Errorf("cannot archive a scope")
	}
	if searchAreaRe.MatchString(ref) {
		return nil, fmt.Errorf("cannot archive an area")
	}

	return nil, fmt.Errorf("invalid reference: %q", ref)
}

func archiveID(v *Vault, m []string) (*ArchiveResult, error) {
	scopeNum, _ := strconv.Atoi(m[1])
	catNum, _ := strconv.Atoi(m[2])
	idNum, _ := strconv.Atoi(m[3])

	id, err := findID(v, scopeNum, catNum, idNum)
	if err != nil {
		return nil, err
	}

	// Archive folder is S0X.XX.09 Archive for S0X.XX inside the category
	categoryPath := filepath.Dir(id.Path)
	archiveFolderName := fmt.Sprintf("S%02d.%02d.09 Archive for S%02d.%02d", scopeNum, catNum, scopeNum, catNum)
	archivePath := filepath.Join(categoryPath, archiveFolderName)

	if err := os.MkdirAll(archivePath, 0o755); err != nil {
		return nil, fmt.Errorf("creating archive folder: %w", err)
	}

	// Rename to [Archived] Name
	newName := fmt.Sprintf("[Archived] %s", id.Name)
	newPath := filepath.Join(archivePath, newName)

	if err := os.Rename(id.Path, newPath); err != nil {
		return nil, fmt.Errorf("moving to archive: %w", err)
	}

	ref := fmt.Sprintf("S%02d.%02d.%02d", scopeNum, catNum, idNum)
	return &ArchiveResult{Ref: ref, NewPath: newPath}, nil
}

func archiveCategory(v *Vault, m []string) (*ArchiveResult, error) {
	scopeNum, _ := strconv.Atoi(m[1])
	catNum, _ := strconv.Atoi(m[2])

	cat, err := findCategory(v, scopeNum, catNum)
	if err != nil {
		return nil, err
	}

	// Archive folder is S0X.X0.09 Archive for S0X.X0-X9 inside the area
	areaPath := filepath.Dir(cat.Path)
	mgmtNum := (catNum / 10) * 10 // e.g., 11 → 10, 21 → 20
	rangeEnd := mgmtNum + 9
	archiveFolderName := fmt.Sprintf("S%02d.%02d.09 Archive for S%02d.%02d-%02d", scopeNum, mgmtNum, scopeNum, mgmtNum, rangeEnd)
	archivePath := filepath.Join(areaPath, archiveFolderName)

	if err := os.MkdirAll(archivePath, 0o755); err != nil {
		return nil, fmt.Errorf("creating archive folder: %w", err)
	}

	// Category keeps its ID
	folderName := filepath.Base(cat.Path)
	newPath := filepath.Join(archivePath, folderName)

	if err := os.Rename(cat.Path, newPath); err != nil {
		return nil, fmt.Errorf("moving to archive: %w", err)
	}

	ref := fmt.Sprintf("S%02d.%02d", scopeNum, catNum)
	return &ArchiveResult{Ref: ref, NewPath: newPath}, nil
}
