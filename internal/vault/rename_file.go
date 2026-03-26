package vault

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// RenameFileResult contains information about a renamed file.
type RenameFileResult struct {
	OldPath        string
	NewPath        string
	LinksUpdated   int
	HeadingUpdated bool
	IsJDex         bool
	FolderRenamed  bool
}

// RenameFile renames a file within a JD ID folder and updates wikilinks across the vault.
// If the file is the JDex file (stem matches folder name), the folder is also renamed.
func RenameFile(v *Vault, ref, oldFilename, newFilename string) (*RenameFileResult, error) {
	if ref == "" {
		return nil, fmt.Errorf("empty reference")
	}
	if oldFilename == "" {
		return nil, fmt.Errorf("empty old filename")
	}
	if newFilename == "" {
		return nil, fmt.Errorf("empty new filename")
	}

	// Only ID refs are valid
	m := searchIDRe.FindStringSubmatch(ref)
	if m == nil {
		return nil, fmt.Errorf("invalid reference: %q (only ID refs are supported)", ref)
	}

	// Auto-append .md
	if !strings.HasSuffix(oldFilename, ".md") {
		oldFilename += ".md"
	}
	if !strings.HasSuffix(newFilename, ".md") {
		newFilename += ".md"
	}

	// No-op
	if oldFilename == newFilename {
		return &RenameFileResult{}, nil
	}

	scopeNum, _ := strconv.Atoi(m[1])
	catNum, _ := strconv.Atoi(m[2])
	idNum, _ := strconv.Atoi(m[3])

	id, err := findID(v, scopeNum, catNum, idNum)
	if err != nil {
		return nil, err
	}

	folderName := filepath.Base(id.Path)
	jdexFilename := folderName + ".md"
	isJDex := oldFilename == jdexFilename

	oldPath := filepath.Join(id.Path, oldFilename)
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file %q not found in %s", oldFilename, ref)
	}

	// Check target doesn't already exist
	newPath := filepath.Join(id.Path, newFilename)
	if _, err := os.Stat(newPath); err == nil {
		return nil, fmt.Errorf("file %q already exists in %s", newFilename, ref)
	}

	oldStem := strings.TrimSuffix(oldFilename, ".md")
	newStem := strings.TrimSuffix(newFilename, ".md")

	result := &RenameFileResult{
		OldPath: oldPath,
		IsJDex:  isJDex,
	}

	if isJDex {
		// JDex rename: rename folder, then file, then update frontmatter
		newFolderName := newStem
		newFolderPath := filepath.Join(filepath.Dir(id.Path), newFolderName)

		if err := os.Rename(id.Path, newFolderPath); err != nil {
			return nil, fmt.Errorf("renaming folder: %w", err)
		}
		result.FolderRenamed = true

		// Rename JDex file inside new folder
		oldJDex := filepath.Join(newFolderPath, oldFilename)
		newJDex := filepath.Join(newFolderPath, newFilename)
		if err := os.Rename(oldJDex, newJDex); err != nil {
			return nil, fmt.Errorf("renaming JDex file: %w", err)
		}

		result.NewPath = newJDex

		// Update JDex frontmatter
		updateJDexContent(newJDex, folderName, newFolderName)

		// Update H1 heading
		result.HeadingUpdated = updateH1Heading(newJDex, oldStem, newStem)

		// Update wikilinks: folder name, old name (human-readable), stem, stem.md
		replacements := map[string]string{
			folderName:   newFolderName,
			id.Name:      strings.TrimPrefix(newFolderName, fmt.Sprintf("S%02d.%02d.%02d ", scopeNum, catNum, idNum)),
			oldFilename:  newFilename,
		}
		result.LinksUpdated, _ = UpdateWikiLinks(v.Root, replacements)
	} else {
		// Regular file rename
		if err := os.Rename(oldPath, newPath); err != nil {
			return nil, fmt.Errorf("renaming file: %w", err)
		}

		result.NewPath = newPath

		// Update H1 heading
		result.HeadingUpdated = updateH1Heading(newPath, oldStem, newStem)

		// Update wikilinks: both stem and full filename
		replacements := map[string]string{
			oldStem:     newStem,
			oldFilename: newFilename,
		}
		result.LinksUpdated, _ = UpdateWikiLinks(v.Root, replacements)
	}

	return result, nil
}

// updateH1Heading replaces the first H1 heading in a file if it matches oldStem exactly.
// Returns true if the heading was updated.
func updateH1Heading(path, oldStem, newStem string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var lines []string
	updated := false
	inFrontmatter := false
	frontmatterCount := 0

	for scanner.Scan() {
		line := scanner.Text()

		// Track frontmatter
		if line == "---" {
			frontmatterCount++
			if frontmatterCount == 1 {
				inFrontmatter = true
			} else {
				inFrontmatter = false
			}
			lines = append(lines, line)
			continue
		}

		if inFrontmatter {
			lines = append(lines, line)
			continue
		}

		// Check for H1 matching old stem (only replace first occurrence)
		if !updated && line == "# "+oldStem {
			lines = append(lines, "# "+newStem)
			updated = true
			continue
		}

		lines = append(lines, line)
	}

	if !updated {
		return false
	}

	content := strings.Join(lines, "\n") + "\n"
	os.WriteFile(path, []byte(content), 0o644)
	return true
}
