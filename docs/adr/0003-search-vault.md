# ADR-0003: Search vault by ID, name, or content

## Status

Accepted

## Context

Users need to find items in the vault without knowing the exact location. Search is the primary discovery mechanism for both CLI and MCP.

## Decision

Provide a `Search` function that accepts a query string and returns matching items across all levels of the JD hierarchy.

### Search modes (auto-detected)

| Mode       | Query pattern                  | Example            | Behavior                                      |
|------------|--------------------------------|--------------------|-----------------------------------------------|
| By JD ref  | `S\d{2}`, `S\d{2}.\d{2}`, `S\d{2}.\d{2}.\d{2}` | `S01.11` | Exact match on the corresponding level |
| By name    | Plain text                     | `Entertainment`    | Case-insensitive substring match on all levels |
| By content | Prefix `?`                     | `?pasta recipe`    | Search inside markdown files for matching text |

### Edge cases

- Empty query → return error.
- No matches → return empty list, no error.
- Content search on an unreadable file → skip it, no error.
- Name search matches at multiple levels → return all matches.

## Expected behavior

### Function signature

```go
func Search(v *Vault, query string) ([]SearchResult, error)
```

### Output type

```go
type SearchResult struct {
    Type      string // "scope", "area", "category", "id"
    Ref       string // "S01", "S01.10-19", "S01.11", "S01.11.11"
    Name      string
    Path      string
    MatchLine string // non-empty only for content matches
}
```

### Test plan

**Unit tests** (`internal/vault/search_test.go`):
- Empty query → error.
- JD ref query `S01` → matches scope.
- JD ref query `S01.11` → matches category.
- JD ref query `S01.11.11` → matches ID.
- JD ref query `S99` → empty list, no error.
- Name query `Entertainment` → matches category.
- Name query `management` (lowercase) → matches case-insensitively.
- Name query matches multiple levels → returns all.
- Name query with no matches → empty list, no error.

**Integration tests** (`internal/vault/search_integration_test.go`):
- JD ref search against fixture vault → correct result.
- Name search against fixture vault → correct results.
- Content search `?` against fixture vault with a known string in a markdown file → returns result with MatchLine.
- Content search with no matches → empty list.
- Content search skips unreadable files gracefully.

**Acceptance tests** (`internal/vault/search_acceptance_test.go`):
- Every SearchResult has a non-empty Ref, Name, Path, and valid Type.
- JD ref search returns exactly one result of the correct Type.
- Content search results include MatchLine.
- Name search results do not include MatchLine.

## Consequences

- Search becomes the primary discovery tool for CLI and MCP.
- The `?` prefix convention for content search is simple but may need revisiting if queries naturally start with `?`.
- Content search performance depends on vault size; acceptable for a personal vault.
