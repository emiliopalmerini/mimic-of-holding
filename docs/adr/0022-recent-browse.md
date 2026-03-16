# ADR-0022: Recent files in browse

## Status

Accepted

## Context

Users often want to know "what was I working on?" There's no way to see recently modified notes.

## Decision

Add a `Recent` mode to `Browse`. When enabled, returns the N most recently modified `.md` files across the vault, with their JD breadcrumb and modification time.

### Function signature change

New function rather than modifying Browse (different return shape):

```go
// RecentResult represents a recently modified file.
type RecentResult struct {
    Ref      string // JD ref of the parent ID
    Name     string // ID name
    File     string // filename
    ModTime  time.Time
    Breadcrumb string
}

// Recent returns the N most recently modified .md files in the vault.
func Recent(v *Vault, n int, scope string) ([]RecentResult, error)
```

### Behavior

1. Walk all ID folders, collect `.md` files with their modification times.
2. Sort by modification time, most recent first.
3. Return top N (default 10 if n <= 0).
4. Scope filter applies.

### Edge cases

- n <= 0 → default to 10.
- Empty vault → empty results.
- Scope filter with no matches → empty results.

### MCP tool integration

Add as a new tool `recent` since the output shape differs from `browse`.

### Test plan

**Acceptance tests**: Results are sorted by ModTime descending. Each result has Ref and File.
**Unit tests**: Default n. Scope filter validation.
**Integration tests**: Returns files from test vault sorted by time. Scope filter works.

## Consequences

- Adds one new MCP tool (`recent`), bringing total to 13. Acceptable since it's a genuinely different verb.
- Uses file system modification times, which are reliable for Obsidian vaults.
