package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// TemplateInfo describes a template file available for use.
type TemplateInfo struct {
	Name      string // filename stem, e.g. "Recipe Template"
	Filename  string // full filename, e.g. "Recipe Template.md"
	Path      string // absolute path
	Source    string // "category", "area", or "scope"
	SourceRef string // ref of the templates ID, e.g. "S01.12.03"
}

// TemplateVars holds the known variables for template substitution.
type TemplateVars struct {
	Ref   string
	Name  string
	Title string
	Date  string
}

// ApplyTemplate substitutes known {{variables}} in content.
// Unknown variables are left untouched.
func ApplyTemplate(content string, vars TemplateVars) string {
	known := map[string]string{
		"{{ref}}":   vars.Ref,
		"{{name}}":  vars.Name,
		"{{title}}": vars.Title,
		"{{date}}":  vars.Date,
	}
	for placeholder, value := range known {
		if value != "" {
			content = strings.ReplaceAll(content, placeholder, value)
		}
	}
	return content
}

// ListTemplates returns available templates for a category, with hierarchical lookup.
func ListTemplates(v *Vault, categoryRef string) ([]TemplateInfo, error) {
	m := searchCategoryRe.FindStringSubmatch(categoryRef)
	if m == nil {
		return nil, fmt.Errorf("invalid category reference: %q (expected S00.00 format)", categoryRef)
	}

	scopeNum, _ := strconv.Atoi(m[1])
	catNum, _ := strconv.Atoi(m[2])

	// Verify category exists
	_, err := findCategory(v, scopeNum, catNum)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var result []TemplateInfo

	// 1. Category level: SXX.YY.03
	addTemplatesFromID(v, scopeNum, catNum, 3, "category", &result, seen)

	// 2. Area level: find the area's management category (SXX.X0) and its .03 ID
	area := findAreaForCategory(v, scopeNum, catNum)
	if area != nil {
		addTemplatesFromID(v, scopeNum, area.RangeStart, 3, "area", &result, seen)
	}

	// 3. Scope level: SXX.01.03
	addTemplatesFromID(v, scopeNum, 1, 3, "scope", &result, seen)

	return result, nil
}

func findAreaForCategory(v *Vault, scopeNum, catNum int) *Area {
	for _, s := range v.Scopes {
		if s.Number != scopeNum {
			continue
		}
		for i, a := range s.Areas {
			if catNum >= a.RangeStart && catNum <= a.RangeEnd {
				return &s.Areas[i]
			}
		}
	}
	return nil
}

func addTemplatesFromID(v *Vault, scopeNum, catNum, idNum int, source string, result *[]TemplateInfo, seen map[string]bool) {
	id, err := findID(v, scopeNum, catNum, idNum)
	if err != nil {
		return // .03 ID doesn't exist at this level, skip silently
	}

	folderName := filepath.Base(id.Path)
	ref := fmt.Sprintf("S%02d.%02d.%02d", scopeNum, catNum, idNum)

	entries, err := os.ReadDir(id.Path)
	if err != nil {
		return
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		stem := strings.TrimSuffix(e.Name(), ".md")
		// Skip JDex file
		if stem == folderName {
			continue
		}
		// Skip if already seen (shadowing)
		if seen[stem] {
			continue
		}
		seen[stem] = true
		*result = append(*result, TemplateInfo{
			Name:      stem,
			Filename:  e.Name(),
			Path:      filepath.Join(id.Path, e.Name()),
			Source:    source,
			SourceRef: ref,
		})
	}
}

// resolveTemplate finds a template by name in the category hierarchy and returns its content.
func resolveTemplate(v *Vault, scopeNum, catNum int, templateName string) (string, error) {
	templates, err := ListTemplates(v, fmt.Sprintf("S%02d.%02d", scopeNum, catNum))
	if err != nil {
		return "", err
	}

	nameLower := strings.ToLower(templateName)
	for _, tmpl := range templates {
		if strings.ToLower(tmpl.Name) == nameLower {
			data, err := os.ReadFile(tmpl.Path)
			if err != nil {
				return "", fmt.Errorf("reading template %q: %w", tmpl.Name, err)
			}
			return string(data), nil
		}
	}

	return "", fmt.Errorf("template %q not found", templateName)
}

func templateVarsForID(ref, name string) TemplateVars {
	return TemplateVars{
		Ref:   ref,
		Name:  name,
		Title: ref + " " + name,
		Date:  time.Now().Format("2006-01-02"),
	}
}
