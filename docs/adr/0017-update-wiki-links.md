# ADR-0017: Update wiki links on move/rename

## Status

Accepted

## Context

Obsidian uses `[[wiki links]]` to reference notes. When a file or folder is renamed/moved via the filesystem, these links break. Obsidian auto-updates links only when changes are made through its UI.

## Decision

After any move or rename, scan all `.md` files in the vault for matching `[[...]]` patterns and update them.

### Link patterns to match

| Pattern | Example | When updated |
|---------|---------|-------------|
| Full folder name | `[[S01.11.11 Theatre, 2025 Season]]` | Folder rename or move |
| Bare name | `[[Theatre, 2025 Season]]` | Rename only (name changes) |
| JDex file reference | `[[S01.11.11 Theatre, 2025 Season.md]]` | Folder rename or move |
| Piped link display text | `[[S01.11.11 Theatre\|Theatre]]` | Target part updated, display text left alone |

### Function signature

```go
func UpdateWikiLinks(vaultRoot string, replacements map[string]string) (int, error)
```

- `replacements`: map of old name → new name.
- Returns count of links updated across all files.
- Scans all `.md` files recursively under `vaultRoot`.
- Does NOT modify the file being moved/renamed itself.

### Integration with Move/Rename

`Rename` and `Move` call `UpdateWikiLinks` internally and return the count in their result.

### Edge cases

- No links to update → returns 0, no error.
- File can't be read → skip, no error.
- Replacement creates no change → file not rewritten.

### Test plan

**Unit tests**:
- Replace link in a string with known patterns.
- Piped links: only target updated, display text preserved.

**Integration tests**:
- Create fixture files with wiki links, run UpdateWikiLinks, verify links updated.
- File with no matching links → not modified.
- Piped link updated correctly.

**Acceptance tests**:
- After Rename, wiki links across vault point to new name.
- After Move, wiki links across vault point to new ref.

## Consequences

- Full vault scan on every move/rename. Acceptable for personal vaults.
- Links inside the moved/renamed item itself are NOT updated (they're relative and still valid).
