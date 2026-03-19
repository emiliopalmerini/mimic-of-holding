package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// EditFile replaces the first occurrence of oldString with newString in a file
// within a JD ID folder. The file must exist and oldString must appear exactly
// once (to prevent ambiguous edits). Returns the absolute path to the edited file.
func EditFile(v *Vault, ref, filename, oldString, newString string) (string, error) {
	if ref == "" {
		return "", fmt.Errorf("empty reference")
	}
	if filename == "" {
		return "", fmt.Errorf("empty filename")
	}
	if oldString == "" {
		return "", fmt.Errorf("empty old_string")
	}

	m := searchIDRe.FindStringSubmatch(ref)
	if m == nil {
		return "", fmt.Errorf("reference %q is not a valid ID (expected S00.00.00 format)", ref)
	}

	scopeNum, _ := strconv.Atoi(m[1])
	catNum, _ := strconv.Atoi(m[2])
	idNum, _ := strconv.Atoi(m[3])

	id, err := findID(v, scopeNum, catNum, idNum)
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(id.Path, filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}

	content := string(data)

	// No-op if old and new are identical
	if oldString == newString {
		return filePath, nil
	}

	count := strings.Count(content, oldString)
	if count == 0 {
		return "", fmt.Errorf("old_string not found in %s", filename)
	}
	if count > 1 {
		return "", fmt.Errorf("ambiguous edit: old_string appears %d times in %s", count, filename)
	}

	updated := strings.Replace(content, oldString, newString, 1)

	if err := os.WriteFile(filePath, []byte(updated), 0o644); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	return filePath, nil
}
