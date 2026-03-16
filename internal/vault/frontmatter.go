package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// SetFrontmatter sets a scalar frontmatter field in a file within a JD ID folder.
// Creates frontmatter if the file has none.
func SetFrontmatter(v *Vault, ref, file, key, value string) (string, error) {
	path, err := resolveFilePath(v, ref, file)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}

	content := string(data)
	content = setFrontmatterField(content, key, value)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}
	return path, nil
}

// AddToFrontmatterList appends a value to a list field in frontmatter.
// Idempotent: does nothing if the value already exists.
func AddToFrontmatterList(v *Vault, ref, file, key, value string) (string, error) {
	path, err := resolveFilePath(v, ref, file)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}

	content := string(data)
	content = addToFrontmatterList(content, key, value)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}
	return path, nil
}

// RemoveFromFrontmatterList removes a value from a list field in frontmatter.
// Idempotent: does nothing if the value doesn't exist.
func RemoveFromFrontmatterList(v *Vault, ref, file, key, value string) (string, error) {
	path, err := resolveFilePath(v, ref, file)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}

	content := string(data)
	content = removeFromFrontmatterList(content, key, value)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}
	return path, nil
}

func resolveFilePath(v *Vault, ref, file string) (string, error) {
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

	path := filepath.Join(id.Path, file)
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("file not found: %s", path)
	}
	return path, nil
}

// parseFrontmatter splits content into (before-frontmatter, frontmatter-lines, after-frontmatter).
// If no frontmatter exists, returns ("", nil, content).
func parseFrontmatter(content string) (string, []string, string) {
	if !strings.HasPrefix(content, "---\n") {
		return "", nil, content
	}

	rest := content[4:] // skip opening ---\n
	endIdx := strings.Index(rest, "\n---\n")
	if endIdx == -1 {
		// Check if it ends with \n---
		if strings.HasSuffix(rest, "\n---") {
			fmContent := rest[:len(rest)-4]
			lines := strings.Split(fmContent, "\n")
			return "", lines, ""
		}
		return "", nil, content
	}

	fmContent := rest[:endIdx]
	body := rest[endIdx+5:] // skip \n---\n
	lines := strings.Split(fmContent, "\n")
	return "", lines, body
}

func buildContent(lines []string, body string) string {
	var b strings.Builder
	b.WriteString("---\n")
	for _, l := range lines {
		b.WriteString(l)
		b.WriteString("\n")
	}
	b.WriteString("---\n")
	b.WriteString(body)
	return b.String()
}

func setFrontmatterField(content, key, value string) string {
	_, lines, body := parseFrontmatter(content)

	if lines == nil {
		// No frontmatter — create it
		return buildContent([]string{key + ": " + value}, body)
	}

	// Look for existing key
	found := false
	for i, line := range lines {
		if isKeyLine(line, key) {
			lines[i] = key + ": " + value
			found = true
			break
		}
	}
	if !found {
		// Insert before closing, after last key
		lines = append(lines, key+": "+value)
	}

	return buildContent(lines, body)
}

func addToFrontmatterList(content, key, value string) string {
	_, lines, body := parseFrontmatter(content)

	if lines == nil {
		// No frontmatter — create it with a list
		return buildContent([]string{key + ":", "  - " + value}, body)
	}

	// Find the key and its list items
	keyIdx := -1
	lastListIdx := -1
	for i, line := range lines {
		if isKeyLine(line, key) {
			keyIdx = i
			lastListIdx = i
			continue
		}
		if keyIdx >= 0 && i > keyIdx {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "- ") {
				// Check for duplicate
				existing := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
				if existing == value {
					return buildContent(lines, body) // already exists
				}
				lastListIdx = i
			} else {
				break
			}
		}
	}

	if keyIdx == -1 {
		// Key doesn't exist — add it with the value
		lines = append(lines, key+":", "  - "+value)
	} else {
		// Insert after the last list item
		newLine := "  - " + value
		newLines := make([]string, 0, len(lines)+1)
		newLines = append(newLines, lines[:lastListIdx+1]...)
		newLines = append(newLines, newLine)
		newLines = append(newLines, lines[lastListIdx+1:]...)
		lines = newLines
	}

	return buildContent(lines, body)
}

func removeFromFrontmatterList(content, key, value string) string {
	_, lines, body := parseFrontmatter(content)

	if lines == nil {
		return content
	}

	keyIdx := -1
	removeIdx := -1
	for i, line := range lines {
		if isKeyLine(line, key) {
			keyIdx = i
			continue
		}
		if keyIdx >= 0 && i > keyIdx {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "- ") {
				existing := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
				if existing == value {
					removeIdx = i
					break
				}
			} else {
				break
			}
		}
	}

	if removeIdx == -1 {
		return buildContent(lines, body) // not found, no-op
	}

	lines = append(lines[:removeIdx], lines[removeIdx+1:]...)
	return buildContent(lines, body)
}

// isKeyLine checks if a frontmatter line is a top-level key matching the given key name.
func isKeyLine(line, key string) bool {
	if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
		return false
	}
	parts := strings.SplitN(line, ":", 2)
	if len(parts) < 2 {
		return false
	}
	return strings.TrimSpace(parts[0]) == key
}
