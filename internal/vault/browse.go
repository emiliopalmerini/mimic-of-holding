package vault

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	filterScopeRe    = regexp.MustCompile(`^S(\d{2})$`)
	filterAreaRe     = regexp.MustCompile(`^S(\d{2})\.(\d{2})-(\d{2})$`)
	filterCategoryRe = regexp.MustCompile(`^S(\d{2})\.(\d{2})$`)
)

// Browse renders the vault tree as a human-readable indented string.
// An optional filter narrows output to a scope (S01), area (S01.10-19), or category (S01.11).
func Browse(v *Vault, filter string) (string, error) {
	if filter == "" {
		return browseAll(v), nil
	}

	if m := filterScopeRe.FindStringSubmatch(filter); m != nil {
		num, _ := strconv.Atoi(m[1])
		return browseScope(v, num)
	}
	if m := filterAreaRe.FindStringSubmatch(filter); m != nil {
		scopeNum, _ := strconv.Atoi(m[1])
		rangeStart, _ := strconv.Atoi(m[2])
		return browseArea(v, scopeNum, rangeStart)
	}
	if m := filterCategoryRe.FindStringSubmatch(filter); m != nil {
		scopeNum, _ := strconv.Atoi(m[1])
		catNum, _ := strconv.Atoi(m[2])
		return browseCategory(v, scopeNum, catNum)
	}

	return "", fmt.Errorf("invalid filter: %q", filter)
}

func browseAll(v *Vault) string {
	var b strings.Builder
	for i, s := range v.Scopes {
		if i > 0 {
			b.WriteString("\n")
		}
		writeScope(&b, s, 0)
	}
	return b.String()
}

func browseScope(v *Vault, num int) (string, error) {
	for _, s := range v.Scopes {
		if s.Number == num {
			var b strings.Builder
			writeScope(&b, s, 0)
			return b.String(), nil
		}
	}
	return "", fmt.Errorf("scope S%02d not found", num)
}

func browseArea(v *Vault, scopeNum, rangeStart int) (string, error) {
	for _, s := range v.Scopes {
		if s.Number != scopeNum {
			continue
		}
		for _, a := range s.Areas {
			if a.RangeStart == rangeStart {
				var b strings.Builder
				writeArea(&b, a, 0)
				return b.String(), nil
			}
		}
	}
	return "", fmt.Errorf("area S%02d.%02d-* not found", scopeNum, rangeStart)
}

func browseCategory(v *Vault, scopeNum, catNum int) (string, error) {
	for _, s := range v.Scopes {
		if s.Number != scopeNum {
			continue
		}
		for _, a := range s.Areas {
			for _, c := range a.Categories {
				if c.Number == catNum {
					var b strings.Builder
					writeCategory(&b, c, 0)
					return b.String(), nil
				}
			}
		}
	}
	return "", fmt.Errorf("category S%02d.%02d not found", scopeNum, catNum)
}

func writeScope(b *strings.Builder, s Scope, indent int) {
	writeIndent(b, indent)
	fmt.Fprintf(b, "S%02d %s", s.Number, s.Name)
	for _, a := range s.Areas {
		b.WriteString("\n")
		writeArea(b, a, indent+2)
	}
}

func writeArea(b *strings.Builder, a Area, indent int) {
	writeIndent(b, indent)
	fmt.Fprintf(b, "S%02d.%02d-%02d %s", a.ScopeNumber, a.RangeStart, a.RangeEnd, a.Name)
	for _, c := range a.Categories {
		b.WriteString("\n")
		writeCategory(b, c, indent+2)
	}
}

func writeCategory(b *strings.Builder, c Category, indent int) {
	writeIndent(b, indent)
	fmt.Fprintf(b, "S%02d.%02d %s", c.ScopeNumber, c.Number, c.Name)
	for _, id := range c.IDs {
		b.WriteString("\n")
		writeID(b, id, indent+2)
	}
}

func writeID(b *strings.Builder, id ID, indent int) {
	writeIndent(b, indent)
	fmt.Fprintf(b, "S%02d.%02d.%02d %s", id.ScopeNumber, id.CategoryNum, id.Number, id.Name)
}

func writeIndent(b *strings.Builder, n int) {
	for range n {
		b.WriteByte(' ')
	}
}
