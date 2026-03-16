# ADR-0019: Frontmatter/metadata search

## Status

Accepted

## Context

Content search hits markdown body text but can't query YAML frontmatter fields. Finding all items by tag, location, or other metadata requires manual inspection.

## Decision

Add a `meta` search mode that queries YAML frontmatter fields across all JDex files.

### Query syntax

`key:value` — case-insensitive match on frontmatter field values.

Examples:
- `location:Notion` → all items where `location` contains "Notion"
- `tags:jdex` → all items with "jdex" in their tags
- `location:Google` → all items with "Google" in location field

### Function signature change

Add `Meta` field to `SearchOpts`:

```go
type SearchOpts struct {
    Content bool
    Scope   string
    Meta    bool // if true, query is "key:value" format for frontmatter search
}
```

### SearchResult for meta matches

Uses existing `SearchResult` with `MatchLine` set to `"key: value"` showing the matched frontmatter field.

### Edge cases

- Query doesn't contain `:` when Meta=true → error.
- Key not found in frontmatter → skip (no error).
- File has no frontmatter → skip.
- Value match is substring, case-insensitive.

### Test plan

**Unit tests**: Parse `key:value` query. Invalid meta query (no colon) → error.
**Integration tests**: Search by `location:Obsidian` → returns JDex entries. Search by `tags:jdex` → returns tagged items. No matches → empty list.
**Acceptance tests**: All meta results have MatchLine with `key: value` format. Scope filter works with meta search.

## Consequences

- Only searches JDex files (files named after their parent folder), not all markdown files.
- Simple YAML parsing (line-by-line string matching, not full YAML parse). Sufficient for flat frontmatter fields and simple tag lists.
