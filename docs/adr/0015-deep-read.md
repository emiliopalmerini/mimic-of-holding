# ADR-0015: Deep read for areas and categories

## Status

Accepted

## Context

Reading an area or category currently returns only a list of children. To summarize everything in an area, you need multiple sequential reads. A deep read returns all content recursively.

## Decision

Add a `deep` boolean parameter to `Read`. When true, for areas and categories, recursively include JDex content and file listings for all descendants.

### Behavior

| Level + deep | Output |
|-------------|--------|
| `read S01.10-19 --deep` | Area info + all categories with their IDs, JDex content, and file listings |
| `read S01.11 --deep` | Category info + all IDs with JDex content and file listings |
| `read S01.11.11 --deep` | Same as regular read (no change for IDs) |
| `read S01 --deep` | Scope info + all areas/categories/IDs recursively |

### Updated ReadResult

Add `DeepChildren []ReadResult` field — populated only when deep=true.

### Test plan

**Unit tests**: Deep flag with ID ref → same as regular read.
**Integration tests**: Deep read area → contains nested category and ID results. Deep read category → contains nested ID results with content.
**Acceptance tests**: Deep results contain JDex content from descendant IDs. Deep read scope returns full tree.

## Consequences

- Enables area/scope-level summarization in a single call.
- Response size can be large for deep reads of big areas — acceptable for personal vault.
