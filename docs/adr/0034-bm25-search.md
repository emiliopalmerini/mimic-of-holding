# ADR-0034: BM25-ranked content search

## Status

Accepted

## Context

`mimic search ?<query>` does line-level case-insensitive substring matching:
every matching line in every file is returned, in vault traversal order. This
works for "find a specific quote" but is poor for "find pages most about X" —
a question the LLM (and the user) increasingly asks as the wiki grows.

A BM25-ranked content search returns top-K _files_ by relevance, with a
snippet, instead of an unranked stream of lines.

We keep the existing substring search; add a new mode for ranked search.

## Decision

Add a `??<query>` CLI prefix that triggers BM25-ranked search. Single `?` keeps
its current line-level substring behavior.

### CLI syntax

```
mimic search ??<query> [--top N] [--vault ...]
```

- `??` prefix → BM25 ranked search (new).
- `?` prefix → existing substring search (unchanged).
- No prefix → name/ref search (unchanged).
- `--top N` → number of results (default 10). Only applies to `??` mode.

### Domain layer

New field on `SearchOpts`:

```go
type SearchOpts struct {
    // ... existing fields ...
    Ranked bool // if true, content search uses BM25 ranking
    Top    int  // max results in Ranked mode (default 10)
}
```

`Search(v, query, opts)` dispatches: when `opts.Ranked` is true and `opts.Content` is true, it calls a new `searchRanked` path; otherwise the existing flow runs unchanged.

`SearchResult` gains two optional fields populated only in ranked mode:

```go
type SearchResult struct {
    // ... existing fields ...
    Score   float64 // BM25 score (ranked mode only)
    Snippet string  // single-line excerpt (ranked mode only)
}
```

Existing fields (`Type`, `Ref`, `Name`, `Path`, `Breadcrumb`, `MatchLine`) stay
populated as before; ranked-mode results set `Type = "id"`, leave `MatchLine`
empty, and use `Snippet` for the excerpt.

### BM25 implementation

Hand-rolled, no external dependency. Lives in `internal/vault/bm25.go`.

**Tokenization.** Lowercase the input, split on any character that is not a
letter or a digit. Drop tokens shorter than 2 characters. No stemming, no
stopwords in v1.

**Scoring.** Standard BM25 with k1 = 1.5, b = 0.75:

```
score(D, Q) = Σ_t in Q  IDF(t) · (tf(t,D) · (k1+1)) / (tf(t,D) + k1 · (1 - b + b · |D|/avgdl))

IDF(t) = ln((N - n(t) + 0.5) / (n(t) + 0.5) + 1)
```

Where `N` is the number of documents indexed for this search, `n(t)` is the
number of documents containing term `t`, `|D|` is the token count of document
`D`, and `avgdl` is the average document length across the corpus.

**Corpus.** All `.md` files inside the scope filter (or whole vault if no
scope). Hidden directories and the activity log are excluded. Frontmatter is
included in tokenization (aliases, location, tags often carry useful terms).

**Snippet.** The line in the document with the highest count of distinct query
terms; ties broken by line position. Trimmed to ~200 characters around the
match.

**Top-K.** Documents are scored, then sorted by score descending. Documents
with score == 0 (no query term present) are dropped. The top `opts.Top`
results are returned; default 10. `Top <= 0` is treated as 10.

### Performance

For each search call, the corpus is walked once: O(N) file reads. For typical
vault sizes (low thousands of files) this is well under a second. No persistent
index file in v1 — staleness is impossible, and rebuild cost is negligible
relative to interactive use.

If the vault grows past ~50k files (unlikely for personal use), we revisit
with a persistent index.

### Edge cases

- **Empty query after stripping prefix**: error "empty search query" (current
  behavior).
- **Only stopword-like terms / very short tokens**: all tokens dropped → no
  results, exit 0 with "No results found.".
- **No documents in scope**: returns empty result set.
- **A term in the query that no document contains**: contributes 0 to scores
  (IDF defined and positive but tf is 0, product is 0). Doesn't break the
  formula.
- **Documents with identical scores**: stable sort preserves vault traversal
  order as a tie-breaker.

### What is NOT included

- No persistent on-disk index. Search rebuilds the corpus each call.
- No vector / embedding similarity. Pure lexical BM25.
- No stemming, no stopwords, no language-specific tokenization. Latin-script
  words only get reasonable handling.
- No multi-field weighting (title vs body). All text contributes equally.
- No phrase queries. Tokens are matched independently. `"foo bar"` works the
  same as `foo bar`.
- No interactive UI — this is a CLI command.

## Consequences

- The LLM gets a useful "what does the vault know about X?" search. The
  existing line-level grep stays for the case where exact-string matching is
  the goal.
- One new file (`bm25.go`) plus modest extensions to `search.go` /
  `cmd/mimic/search.go` plus tests.
- Snippet rendering changes the CLI output shape only when `??` is used; old
  callers and the `mimic` skill reference still work as documented.
