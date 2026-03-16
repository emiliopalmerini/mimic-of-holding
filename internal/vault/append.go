package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// AppendFile appends content to a file inside a JD ID folder.
// If the file doesn't exist, it creates it. If content is empty, it's a no-op.
// Returns the absolute path to the file.
func AppendFile(v *Vault, ref string, filename string, content string) (string, error) {
	if ref == "" {
		return "", fmt.Errorf("empty reference")
	}
	if filename == "" {
		return "", fmt.Errorf("empty filename")
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

	// Empty content is a no-op
	if content == "" {
		return filePath, nil
	}

	// If file exists, read it and append
	existing, err := os.ReadFile(filePath)
	if err == nil {
		// Add newline separator if file doesn't end with one
		if len(existing) > 0 && existing[len(existing)-1] != '\n' {
			content = "\n" + content
		}
		content = string(existing) + content
	}

	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	return filePath, nil
}
