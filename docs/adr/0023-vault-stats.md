# ADR-0023: Vault stats in browse

## Status

Accepted

## Context

Understanding vault health — empty categories, orphan notes, size distribution — requires manual inspection.

## Decision

Add a `Stats` function that returns vault-level statistics.

### Function signature

```go
type StatsResult struct {
    TotalScopes     int
    TotalAreas      int
    TotalCategories int
    TotalIDs        int
    TotalFiles      int
    EmptyCategories []string // refs of categories with no IDs
    OrphanIDs       []string // refs of IDs with no inbound wiki links
    LargestCategories []CategorySize // top 5 by ID count
}

type CategorySize struct {
    Ref   string
    Name  string
    Count int
}

func Stats(v *Vault) (*StatsResult, error)
```

### Orphan detection

An ID is orphan if no `.md` file anywhere in the vault contains a `[[wiki link]]` to any file in that ID's folder. This reuses the wiki link scanning from backlinks.

### MCP tool integration

Add as a `stats` tool (no parameters, or optional scope filter).

### Test plan

**Acceptance tests**: Stats has correct totals. Empty categories detected. Orphan IDs detected.
**Unit tests**: CategorySize sorting.
**Integration tests**: Stats against test vault returns expected counts.

## Consequences

- Adds one new MCP tool (`stats`), bringing total to 14.
- Orphan detection scans all files — potentially slow on large vaults, but acceptable for MCP use.
