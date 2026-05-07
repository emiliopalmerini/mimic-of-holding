package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// LintIssue describes a single lint finding.
type LintIssue struct {
	Category string
	Ref      string
	Detail   string
}

// LintResult is the aggregate output of a lint pass.
type LintResult struct {
	Scope  string // "S01" or "" for whole-vault
	Issues []LintIssue
}

// embedRe matches Obsidian embeds like ![[Target]].
var embedRe = regexp.MustCompile(`!\[\[([^\]|#^]+)(?:[#^][^\]|]*)?(?:\|[^\]]+)?\]\]`)

// fencedBlockRe matches triple-backtick code fences (and the dataview fence form).
// Used to strip code blocks before scanning for wiki-links/embeds.
var fencedBlockRe = regexp.MustCompile("(?s)```[a-zA-Z0-9_-]*\\n.*?```")

// inlineCodeRe matches single-backtick inline code spans on a single line.
// Used to strip prose-embedded examples (e.g. `[[wiki links]]`) before scanning.
var inlineCodeRe = regexp.MustCompile("`[^`\n]*`")

// stripCodeForLinkScan removes fenced code blocks and inline code spans so
// that wiki-link patterns inside docs/examples do not produce false positives.
func stripCodeForLinkScan(text string) string {
	text = fencedBlockRe.ReplaceAllString(text, "")
	text = inlineCodeRe.ReplaceAllString(text, "")
	return text
}

// Lint runs all deterministic checks against the vault. If scope is
// non-empty (e.g. "S01"), only issues whose source ref is in that scope
// are returned, but the global link / target index spans the whole vault
// so cross-scope references resolve correctly.
func Lint(v *Vault, scope string) (*LintResult, error) {
	var scopeNum int
	scopeFilter := false
	if scope != "" {
		m := searchScopeRe.FindStringSubmatch(scope)
		if m == nil {
			return nil, fmt.Errorf("invalid scope reference: %q", scope)
		}
		scopeNum, _ = strconv.Atoi(m[1])
		scopeFilter = true
		// confirm scope exists in vault
		if _, err := findScope(v, scopeNum); err != nil {
			return nil, err
		}
	}

	idx, links, err := buildLintIndex(v)
	if err != nil {
		return nil, err
	}

	result := &LintResult{Scope: scope}

	for si := range v.Scopes {
		s := &v.Scopes[si]
		if scopeFilter && s.Number != scopeNum {
			continue
		}
		for ai := range s.Areas {
			a := &s.Areas[ai]
			for ci := range a.Categories {
				c := &a.Categories[ci]
				checkNumberingGaps(c, &result.Issues)
				for idi := range c.IDs {
					id := &c.IDs[idi]
					checkID(id, idx, links, &result.Issues)
				}
			}
		}
	}

	// Broken links: walk markdown files within scope and verify every link
	// resolves against the global target index.
	if err := checkBrokenLinks(v, scopeFilter, scopeNum, idx, &result.Issues); err != nil {
		return nil, err
	}

	sortIssues(result.Issues)
	return result, nil
}

// FormatLintReport renders a LintResult as the markdown report described in
// ADR-0033.
func FormatLintReport(r *LintResult) string {
	var b strings.Builder
	title := "All scopes"
	if r.Scope != "" {
		title = r.Scope
	}
	b.WriteString("# Lint Report — ")
	b.WriteString(title)
	b.WriteString("\n")

	if len(r.Issues) == 0 {
		b.WriteString("\nNo issues found.\n")
		return b.String()
	}

	groups := groupIssues(r.Issues)
	for _, cat := range orderedCategories() {
		issues := groups[cat]
		if len(issues) == 0 {
			continue
		}
		fmt.Fprintf(&b, "\n## %s (%d)\n", cat, len(issues))
		for _, it := range issues {
			fmt.Fprintf(&b, "- `%s`: %s\n", it.Ref, it.Detail)
		}
	}
	return b.String()
}

// --- internal: index ---

type lintIndex struct {
	// targets maps a normalized link target (stem, folder name, or alias) to true.
	targets map[string]bool
	// aliasesByID maps an ID's folder name to its frontmatter aliases.
	aliasesByID map[string][]string
}

func buildLintIndex(v *Vault) (*lintIndex, map[string]bool, error) {
	idx := &lintIndex{
		targets:     map[string]bool{},
		aliasesByID: map[string][]string{},
	}
	links := map[string]bool{}

	walkErr := filepath.Walk(v.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			// Folder name itself can be a link target (Obsidian resolves
			// `[[S01.11.11 Theatre, 2025 Season]]` to its JDex).
			idx.targets[info.Name()] = true
			return nil
		}
		// Index every file as a potential link target. Markdown files are
		// stored both with and without the ".md" extension; non-markdown
		// (PDFs, images, scripts) are stored under their full filename and
		// also under their stem so `[[file.pdf]]` and `[[file]]` both resolve.
		idx.targets[info.Name()] = true
		ext := filepath.Ext(info.Name())
		stem := strings.TrimSuffix(info.Name(), ext)
		idx.targets[stem] = true

		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		data, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		text := string(data)

		// Aliases
		_, fmLines, _ := parseFrontmatter(text)
		if len(fmLines) > 0 {
			aliases := extractAliases(fmLines)
			for _, a := range aliases {
				idx.targets[a] = true
			}
			// Index aliases for this file's containing folder if it is a JDex
			folderName := filepath.Base(filepath.Dir(path))
			if stem == folderName {
				idx.aliasesByID[folderName] = aliases
			}
		}

		// Links — strip code fences first so dataview-style content is ignored.
		stripped := stripCodeForLinkScan(text)
		for _, m := range wikiLinkRe.FindAllStringSubmatch(stripped, -1) {
			head := normalizeLinkHead(m[1])
			if head != "" {
				links[head] = true
			}
		}
		for _, m := range embedRe.FindAllStringSubmatch(stripped, -1) {
			head := normalizeLinkHead(m[1])
			if head != "" {
				links[head] = true
			}
		}
		return nil
	})
	if walkErr != nil {
		return nil, nil, fmt.Errorf("walking vault: %w", walkErr)
	}
	return idx, links, nil
}

// extractAliases parses the aliases: list from frontmatter lines.
// Supports both block-style (multi-line) and inline ([a, b]) lists.
func extractAliases(lines []string) []string {
	var aliases []string
	inAliases := false
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "aliases:") {
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "aliases:"))
			if strings.HasPrefix(rest, "[") && strings.HasSuffix(rest, "]") {
				inner := strings.Trim(rest, "[]")
				for _, p := range strings.Split(inner, ",") {
					a := strings.Trim(strings.TrimSpace(p), `"'`)
					if a != "" {
						aliases = append(aliases, a)
					}
				}
				inAliases = false
				continue
			}
			inAliases = true
			continue
		}
		if inAliases {
			if strings.HasPrefix(trimmed, "- ") {
				a := strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")), `"'`)
				if a != "" {
					aliases = append(aliases, a)
				}
				continue
			}
			// Indented continuation; otherwise we have left the list.
			if !strings.HasPrefix(l, " ") && !strings.HasPrefix(l, "\t") {
				inAliases = false
			}
		}
	}
	return aliases
}

// normalizeLinkHead returns the canonical lookup key for a wikilink target.
// Drops everything from the first '|', '#', or '^'; takes the last path segment.
func normalizeLinkHead(s string) string {
	s = strings.TrimSpace(s)
	for _, sep := range []string{"|", "#", "^"} {
		if i := strings.Index(s, sep); i >= 0 {
			s = s[:i]
		}
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if i := strings.LastIndex(s, "/"); i >= 0 {
		s = s[i+1:]
	}
	return strings.TrimSuffix(s, ".md")
}

// --- internal: per-ID checks ---

func checkID(id *ID, idx *lintIndex, links map[string]bool, out *[]LintIssue) {
	folderName := filepath.Base(id.Path)
	jdexPath := filepath.Join(id.Path, folderName+".md")
	jdexExists := false
	if info, err := os.Stat(jdexPath); err == nil && !info.IsDir() {
		jdexExists = true
	}

	// Standard zero IDs (.01–.09) are system folders (Inbox, Templates,
	// Archive, etc.). They are conventionally bare and not flagged for
	// missing-jdex. If they do carry a JDex, its frontmatter is still checked.
	if !jdexExists {
		if !id.IsSystemID {
			*out = append(*out, LintIssue{
				Category: "missing-jdex",
				Ref:      folderName,
				Detail:   "folder has no JDex matching its name",
			})
		}
	} else {
		checkJDexFrontmatter(jdexPath, folderName, out)
	}

	// Name-mismatch: any .md file inside the ID whose stem looks like a JD ref
	// but does not match this folder.
	entries, err := os.ReadDir(id.Path)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			stem := strings.TrimSuffix(e.Name(), ".md")
			if idRe.MatchString(stem) && stem != folderName {
				*out = append(*out, LintIssue{
					Category: "name-mismatch",
					Ref:      filepath.Join(folderName, e.Name()),
					Detail:   fmt.Sprintf("file stem %q does not match folder name", stem),
				})
			}
		}
	}

	// Orphan ID — skip standard zeros.
	if id.IsSystemID {
		return
	}
	if !idIsLinked(folderName, idx.aliasesByID[folderName], links) {
		*out = append(*out, LintIssue{
			Category: "orphan-id",
			Ref:      folderName,
			Detail:   "not referenced by any [[link]]",
		})
	}
}

func idIsLinked(folderName string, aliases []string, links map[string]bool) bool {
	if links[folderName] {
		return true
	}
	for _, a := range aliases {
		if links[a] {
			return true
		}
	}
	return false
}

func checkJDexFrontmatter(jdexPath, folderName string, out *[]LintIssue) {
	data, err := os.ReadFile(jdexPath)
	if err != nil {
		*out = append(*out, LintIssue{
			Category: "broken-frontmatter",
			Ref:      filepath.Join(folderName, folderName+".md"),
			Detail:   fmt.Sprintf("cannot read: %v", err),
		})
		return
	}
	_, lines, _ := parseFrontmatter(string(data))
	if lines == nil {
		*out = append(*out, LintIssue{
			Category: "broken-frontmatter",
			Ref:      filepath.Join(folderName, folderName+".md"),
			Detail:   "no frontmatter block",
		})
		return
	}
	hasAliases := false
	hasLocation := false
	tags := []string{}
	inTags := false
	for _, l := range lines {
		t := strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(t, "aliases:"):
			hasAliases = true
		case strings.HasPrefix(t, "location:"):
			rest := strings.TrimSpace(strings.TrimPrefix(t, "location:"))
			if rest != "" {
				hasLocation = true
			}
		case strings.HasPrefix(t, "tags:"):
			inTags = true
			rest := strings.TrimSpace(strings.TrimPrefix(t, "tags:"))
			if strings.HasPrefix(rest, "[") && strings.HasSuffix(rest, "]") {
				inner := strings.Trim(rest, "[]")
				for _, p := range strings.Split(inner, ",") {
					tags = append(tags, strings.Trim(strings.TrimSpace(p), `"'`))
				}
				inTags = false
			}
		default:
			if inTags && strings.HasPrefix(t, "- ") {
				tags = append(tags, strings.Trim(strings.TrimSpace(strings.TrimPrefix(t, "- ")), `"'`))
			} else if inTags && !strings.HasPrefix(l, " ") && !strings.HasPrefix(l, "\t") {
				inTags = false
			}
		}
	}
	missing := []string{}
	if !hasAliases {
		missing = append(missing, "aliases")
	}
	if !hasLocation {
		missing = append(missing, "location")
	}
	hasJdexTag := false
	hasIndexTag := false
	for _, tg := range tags {
		if tg == "jdex" {
			hasJdexTag = true
		}
		if tg == "index" {
			hasIndexTag = true
		}
	}
	if !hasJdexTag {
		missing = append(missing, "tags: jdex")
	}
	if !hasIndexTag {
		missing = append(missing, "tags: index")
	}
	if len(missing) > 0 {
		*out = append(*out, LintIssue{
			Category: "broken-frontmatter",
			Ref:      filepath.Join(folderName, folderName+".md"),
			Detail:   "missing " + strings.Join(missing, ", "),
		})
	}
}

func checkNumberingGaps(c *Category, out *[]LintIssue) {
	var nums []int
	for _, id := range c.IDs {
		if id.IsSystemID {
			continue
		}
		nums = append(nums, id.Number)
	}
	if len(nums) < 2 {
		return
	}
	sort.Ints(nums)
	for i := 1; i < len(nums); i++ {
		prev, cur := nums[i-1], nums[i]
		if cur > prev+1 {
			gaps := []string{}
			for n := prev + 1; n < cur; n++ {
				gaps = append(gaps, fmt.Sprintf(".%02d", n))
			}
			ref := fmt.Sprintf("S%02d.%02d", c.ScopeNumber, c.Number)
			*out = append(*out, LintIssue{
				Category: "numbering-gap",
				Ref:      ref,
				Detail: fmt.Sprintf("gap at %s (between .%02d and .%02d)",
					strings.Join(gaps, ", "), prev, cur),
			})
		}
	}
}

// --- internal: broken-link walk ---

func checkBrokenLinks(v *Vault, scopeFilter bool, scopeNum int, idx *lintIndex, out *[]LintIssue) error {
	root := v.Root
	if scopeFilter {
		s, _ := findScope(v, scopeNum)
		root = s.Path
	}
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}
		data, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		text := stripCodeForLinkScan(string(data))
		rel, _ := filepath.Rel(v.Root, path)

		seen := map[string]bool{}
		flag := func(target string) {
			if seen[target] {
				return
			}
			seen[target] = true
			*out = append(*out, LintIssue{
				Category: "broken-link",
				Ref:      rel,
				Detail:   fmt.Sprintf("`[[%s]]` does not resolve", target),
			})
		}
		for _, m := range wikiLinkRe.FindAllStringSubmatch(text, -1) {
			head := normalizeLinkHead(m[1])
			if head == "" {
				continue
			}
			if !idx.targets[head] {
				flag(head)
			}
		}
		for _, m := range embedRe.FindAllStringSubmatch(text, -1) {
			head := normalizeLinkHead(m[1])
			if head == "" {
				continue
			}
			if !idx.targets[head] {
				flag(head)
			}
		}
		return nil
	})
}

// --- internal: report formatting ---

func orderedCategories() []string {
	return []string{"missing-jdex", "broken-frontmatter", "name-mismatch", "numbering-gap", "broken-link", "orphan-id"}
}

func groupIssues(issues []LintIssue) map[string][]LintIssue {
	g := map[string][]LintIssue{}
	for _, it := range issues {
		g[it.Category] = append(g[it.Category], it)
	}
	for k := range g {
		sort.Slice(g[k], func(i, j int) bool {
			if g[k][i].Ref != g[k][j].Ref {
				return g[k][i].Ref < g[k][j].Ref
			}
			return g[k][i].Detail < g[k][j].Detail
		})
	}
	return g
}

func sortIssues(issues []LintIssue) {
	order := map[string]int{}
	for i, c := range orderedCategories() {
		order[c] = i
	}
	sort.SliceStable(issues, func(i, j int) bool {
		oi, oj := order[issues[i].Category], order[issues[j].Category]
		if oi != oj {
			return oi < oj
		}
		if issues[i].Ref != issues[j].Ref {
			return issues[i].Ref < issues[j].Ref
		}
		return issues[i].Detail < issues[j].Detail
	})
}

// SummarizeLintIssues returns a one-line summary suitable for the activity log.
func SummarizeLintIssues(issues []LintIssue) string {
	if len(issues) == 0 {
		return "No issues found."
	}
	counts := map[string]int{}
	for _, it := range issues {
		counts[it.Category]++
	}
	cats := orderedCategories()
	parts := []string{}
	used := 0
	for _, c := range cats {
		if n := counts[c]; n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, c))
			used++
		}
	}
	return fmt.Sprintf("%d issues across %d categories: %s.", len(issues), used, strings.Join(parts, ", "))
}
