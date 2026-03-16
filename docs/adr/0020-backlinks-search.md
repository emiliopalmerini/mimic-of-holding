# ADR-0020: Backlinks search mode

## Status

Accepted

## Context

Users can search by name, content, and frontmatter, but cannot discover what links *to* a given note. Backlinks are essential for understanding how knowledge is connected across the vault.

## Decision

Add a `Backlinks` mode to `SearchOpts`. When enabled, the query is a JD reference (e.g., `S01.11.11`), and the search returns all notes containing wiki links (`[[...]]`) that point to any file in that ID's folder.

### Function signature change

```go
type SearchOpts struct {
    Content   bool
    Scope     string
    Meta      bool
    Backlinks bool // if true, query is a JD ref; returns notes linking to it
}
```

### Matching logic

1. Resolve the query ref to get the ID's folder name (e.g., `S01.11.11 Theatre, 2025 Season`).
2. Scan all `.md` files across the vault for wiki links matching the folder name or any file within it.
3. Return `SearchResult` entries for each linking note, with `MatchLine` showing the line containing the link.

### Edge cases

- Invalid ref → error.
- Ref not found → error.
- No backlinks → empty results.
- Self-links (file links to itself) → excluded.
- Links with display text `[[target|display]]` → matched on target.
- Scope filter applies to the *linking* files, not the target.

### Test plan

**Acceptance tests**: Backlink results have correct Ref, Type, and MatchLine. Self-links excluded.
**Unit tests**: Invalid/missing ref → error.
**Integration tests**: `S01.11.11` has backlinks from `S02.11.17`. Scope filter excludes cross-scope links. No backlinks → empty.

## Consequences

- Reuses the existing `wikiLinkRe` regex from `wikilinks.go`.
- Searches all `.md` files, not just JDex files, since any note can contain links.
