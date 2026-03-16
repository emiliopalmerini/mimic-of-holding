# ADR-0021: Tags search mode

## Status

Accepted

## Context

The vault uses both YAML frontmatter `tags:` and inline `#tag` markers. There's no way to list all tags or find all notes with a specific tag without manually searching.

## Decision

Add a `Tags` mode to `SearchOpts`. Behavior depends on whether a query is provided:

- **No query** (empty string with `Tags: true`): List all unique tags in the vault with counts, sorted by count descending.
- **With query**: Return all notes that have the specified tag (in frontmatter `tags:` list or inline `#tag`).

### Function signature change

```go
type SearchOpts struct {
    Content   bool
    Scope     string
    Meta      bool
    Backlinks bool
    Tags      bool // if true, list tags (no query) or find notes by tag (with query)
}
```

### Tag extraction

- YAML frontmatter: `tags:` field, either inline (`tags: [a, b]`) or list format (`- a`).
- Inline: `#tag` in body text (not inside code blocks). Match `#\w+`.

### Output for tag listing (no query)

Returns a single `SearchResult` with Type `"tags"` and the formatted list in `Name`:
```
#jdex (4)
#index (3)
#draft (1)
```

### Output for tag filter (with query)

Returns `SearchResult` entries for each note matching the tag, similar to name search results.

### Edge cases

- Tag query with `#` prefix → strip it (search for `jdex`, not `#jdex`).
- No tags in vault → "No tags found." for listing, empty results for filter.
- Scope filter applies to both modes.

### Test plan

**Acceptance tests**: Tag listing returns sorted counts. Tag filter returns correct notes.
**Unit tests**: Tag extraction from frontmatter and inline. Query normalization (strip `#`).
**Integration tests**: List tags across vault. Filter by `jdex` → returns tagged notes. Scope filter works.

## Consequences

- Tag extraction scans all `.md` files in ID folders.
- Simple regex-based extraction, not a full YAML parser.
