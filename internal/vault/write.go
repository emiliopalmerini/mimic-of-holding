package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// WriteFile creates or overwrites a file inside a JD ID folder.
// If template is non-empty and content is empty, the named template is used.
// Returns the absolute path to the written file.
func WriteFile(v *Vault, ref string, filename string, content string, template string) (string, error) {
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

	// Resolve content from template if needed
	if content == "" && template != "" {
		tmplContent, err := resolveTemplate(v, scopeNum, catNum, template)
		if err != nil {
			return "", err
		}
		content = ApplyTemplate(tmplContent, templateVarsForID(ref, id.Name))
	}

	filePath := filepath.Join(id.Path, filename)
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	return filePath, nil
}
