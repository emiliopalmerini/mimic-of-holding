# ADR-0024: Frontmatter field editing

## Status

Accepted

## Context

Updating a single frontmatter field (e.g., adding a tag) requires reading the entire file, modifying it, and writing it back. This is error-prone and verbose for the LLM.

## Decision

Add a `SetFrontmatter` function that sets a specific YAML frontmatter field in a file.

### Function signature

```go
// SetFrontmatter sets a frontmatter field in a file within a JD ID folder.
// If the file has no frontmatter, it adds one.
// For list fields (like tags), use AddToFrontmatterList / RemoveFromFrontmatterList.
func SetFrontmatter(v *Vault, ref, file, key, value string) (string, error)

// AddToFrontmatterList appends a value to a list field in frontmatter.
func AddToFrontmatterList(v *Vault, ref, file, key, value string) (string, error)

// RemoveFromFrontmatterList removes a value from a list field in frontmatter.
func RemoveFromFrontmatterList(v *Vault, ref, file, key, value string) (string, error)
```

### MCP tool integration

Extend write tool or add a `frontmatter` tool with an `action` parameter:
- `set` — set scalar field
- `add` — add to list field
- `remove` — remove from list field

Single tool with action parameter to avoid tool proliferation.

### Frontmatter parsing

Line-by-line string manipulation (consistent with existing `searchFrontmatter`):
1. Find `---` delimiters.
2. For `set`: find existing key line and replace value, or append new key.
3. For `add`: find key's list, append `- value`.
4. For `remove`: find and remove matching `- value` line.
5. If no frontmatter exists, create `---\nkey: value\n---\n` prefix.

### Edge cases

- File doesn't exist → error.
- Key doesn't exist for `set` → add it.
- Key doesn't exist for `add` → create list with single item.
- Value already in list for `add` → no-op (idempotent).
- Value not in list for `remove` → no-op (idempotent).
- No frontmatter for any action → create it.

### Test plan

**Acceptance tests**: Set/add/remove produce valid frontmatter. Existing content preserved.
**Unit tests**: Frontmatter parsing edge cases. No-frontmatter creation.
**Integration tests**: Set field, read back. Add tag, verify. Remove tag, verify. Idempotent operations.

## Consequences

- Adds one new MCP tool (`frontmatter`), bringing total to 15.
- Line-by-line YAML manipulation is fragile for complex YAML but sufficient for the flat frontmatter used in Obsidian vaults.
