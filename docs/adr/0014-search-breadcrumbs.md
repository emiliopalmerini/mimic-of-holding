# ADR-0014: Search result breadcrumbs

## Status

Accepted

## Context

Content search shows the ID ref but not the human-readable path through the hierarchy. `S01.12.01` is meaningless without knowing it's `S01 Me > S01.12 Food > S01.12.01 Inbox`.

## Decision

Add a `Breadcrumb` field to `SearchResult` showing the full human-readable path.

### Format

```
S01 Me > S01.10-19 Lifestyle > S01.11 Entertainment > S01.11.11 Theatre, 2025 Season
```

### Behavior

- Populated for all search results (name and content matches).
- Built from the vault tree during search.

### Test plan

**Unit tests**: Breadcrumb populated for name search results.
**Integration tests**: Breadcrumb matches expected hierarchy for fixture items.
**Acceptance tests**: Every SearchResult has non-empty Breadcrumb. Content results include full path.

## Consequences

- Breadcrumb makes search results self-contained — the LLM doesn't need to call browse to understand context.
- Minor performance cost (string building per result), negligible for personal vault sizes.
