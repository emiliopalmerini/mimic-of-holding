# ADR-0012: Fix content search

## Status

Accepted

## Context

Content search returns a flat list of decontextualized lines with no filenames, no grouping, and no scope filtering. Results are noisy and unusable for an LLM.

## Decision

Restructure content search to return grouped results with file context, and add scope filtering.

### Changes

1. `Search` gains an optional `scope` parameter.
2. Content search results are grouped by ID, include the filename, and cap at 3 matching lines per file.
3. MCP `search` tool gets explicit `content: bool` and `scope: string` params instead of `?` prefix.
4. CLI keeps `?` prefix for content search.

### Updated SearchResult for content matches

```go
type SearchResult struct {
    Type      string   // "scope", "area", "category", "id"
    Ref       string
    Name      string
    Path      string
    MatchLine string   // for content: "filename: matching line"
}
```

MatchLine format changes from raw line to `"filename: line content"`.

### Function signature change

```go
func Search(v *Vault, query string, opts SearchOpts) ([]SearchResult, error)

type SearchOpts struct {
    Content bool   // if true, search file content instead of names
    Scope   string // optional scope filter (e.g., "S01")
}
```

### Content search improvements

- Group by ID: max 3 matching lines per file, max 1 entry per file.
- Include filename in MatchLine: `"Recipe.md: Add the pasta to boiling water"`.
- Scope filter limits which scopes are searched.

### Test plan

**Unit tests**: Scope filter limits results. Content flag triggers content search.
**Integration tests**: Content search returns filename in MatchLine. Max 3 lines per file. Scope filter works with content search.
**Acceptance tests**: Content results always have filename prefix in MatchLine. Scope filter excludes other scopes.

## Consequences

- Breaking change to `Search` signature — all callers need updating.
- MCP gets better params (`content`, `scope`). CLI uses opts internally.
