package vault

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// BM25 hyperparameters. Standard defaults; not tuned per corpus.
const (
	bm25K1 = 1.5
	bm25B  = 0.75
)

// tokenSplitRe splits on any non-alphanumeric character.
var tokenSplitRe = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// tokenize lowercases the input, splits on non-alphanumeric, drops tokens
// shorter than 2 characters.
func tokenize(text string) []string {
	parts := tokenSplitRe.Split(strings.ToLower(text), -1)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) >= 2 {
			out = append(out, p)
		}
	}
	return out
}

// rankedDoc holds per-document state during BM25 scoring.
type rankedDoc struct {
	path   string
	idPath string // absolute path of the containing ID folder
	tf     map[string]int
	length int
	lines  []string // original lines for snippet selection
}

// searchRanked walks the scopes' .md files, scores each against the query
// using BM25, and returns the top opts.Top results sorted by score desc.
func searchRanked(v *Vault, scopes []Scope, query string, opts SearchOpts) ([]SearchResult, error) {
	terms := tokenize(query)
	if len(terms) == 0 {
		return []SearchResult{}, nil
	}
	top := opts.Top
	if top <= 0 {
		top = 10
	}

	// Build per-ID lookup so we can attribute results to JD breadcrumbs.
	type idEntry struct {
		s  Scope
		a  Area
		c  Category
		id ID
	}
	idsByPath := map[string]idEntry{}
	for _, s := range scopes {
		for _, a := range s.Areas {
			for _, c := range a.Categories {
				for _, id := range c.IDs {
					idsByPath[id.Path] = idEntry{s: s, a: a, c: c, id: id}
				}
			}
		}
	}

	// Collect docs whose containing ID folder is one we know about. We only
	// score files that live directly inside an ID folder; loose files in
	// scope/area/category folders are skipped (rare in practice).
	var docs []rankedDoc
	for idPath := range idsByPath {
		entries, err := os.ReadDir(idPath)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			path := filepath.Join(idPath, e.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			text := string(data)
			tokens := tokenize(text)
			if len(tokens) == 0 {
				continue
			}
			tf := map[string]int{}
			for _, tok := range tokens {
				tf[tok]++
			}
			docs = append(docs, rankedDoc{
				path:   path,
				idPath: idPath,
				tf:     tf,
				length: len(tokens),
				lines:  strings.Split(text, "\n"),
			})
		}
	}
	if len(docs) == 0 {
		return []SearchResult{}, nil
	}

	// Corpus stats.
	n := len(docs)
	avgdl := 0.0
	df := map[string]int{}
	for _, d := range docs {
		avgdl += float64(d.length)
		for term := range d.tf {
			df[term]++
		}
	}
	avgdl /= float64(n)

	// Score each document.
	type scoredDoc struct {
		idx   int
		score float64
	}
	scored := make([]scoredDoc, 0, len(docs))
	for i, d := range docs {
		score := 0.0
		for _, term := range terms {
			tf, ok := d.tf[term]
			if !ok {
				continue
			}
			nq := df[term]
			idf := math.Log((float64(n-nq)+0.5)/(float64(nq)+0.5) + 1)
			numer := float64(tf) * (bm25K1 + 1)
			denom := float64(tf) + bm25K1*(1-bm25B+bm25B*float64(d.length)/avgdl)
			score += idf * numer / denom
		}
		if score > 0 {
			scored = append(scored, scoredDoc{idx: i, score: score})
		}
	}

	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})
	if len(scored) > top {
		scored = scored[:top]
	}

	results := make([]SearchResult, 0, len(scored))
	for _, s := range scored {
		d := docs[s.idx]
		entry := idsByPath[d.idPath]
		ref := refFromID(entry.id)
		results = append(results, SearchResult{
			Type:       "id",
			Ref:        ref,
			Name:       entry.id.Name,
			Path:       d.path,
			Breadcrumb: idBreadcrumb(entry.s, entry.a, entry.c, entry.id),
			Score:      s.score,
			Snippet:    bestSnippet(d.lines, terms),
		})
	}
	return results, nil
}

// bestSnippet returns the line with the most distinct query terms, trimmed to
// ~200 characters centered around the first match.
func bestSnippet(lines, terms []string) string {
	termSet := make(map[string]bool, len(terms))
	for _, t := range terms {
		termSet[t] = true
	}
	bestIdx := -1
	bestHits := 0
	for i, line := range lines {
		hits := distinctTermHits(line, termSet)
		if hits > bestHits {
			bestHits = hits
			bestIdx = i
		}
	}
	if bestIdx == -1 {
		return ""
	}
	line := strings.TrimSpace(lines[bestIdx])
	const maxLen = 200
	if len(line) <= maxLen {
		return line
	}
	lower := strings.ToLower(line)
	first := -1
	for t := range termSet {
		if i := strings.Index(lower, t); i >= 0 {
			if first == -1 || i < first {
				first = i
			}
		}
	}
	if first == -1 {
		return line[:maxLen] + "…"
	}
	start := first - maxLen/2
	if start < 0 {
		start = 0
	}
	end := start + maxLen
	if end > len(line) {
		end = len(line)
		start = end - maxLen
		if start < 0 {
			start = 0
		}
	}
	prefix := ""
	if start > 0 {
		prefix = "…"
	}
	suffix := ""
	if end < len(line) {
		suffix = "…"
	}
	return prefix + line[start:end] + suffix
}

func distinctTermHits(line string, terms map[string]bool) int {
	tokens := tokenize(line)
	seen := map[string]bool{}
	for _, t := range tokens {
		if terms[t] && !seen[t] {
			seen[t] = true
		}
	}
	return len(seen)
}

func refFromID(id ID) string {
	return fmt.Sprintf("S%02d.%02d.%02d", id.ScopeNumber, id.CategoryNum, id.Number)
}
