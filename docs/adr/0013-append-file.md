# ADR-0013: Append to files inside JD IDs

## Status

Accepted

## Context

The `WriteFile` function overwrites entirely. The most common operation is adding content to an existing note without rewriting it. An append mode avoids the read-then-rewrite dance.

## Decision

Provide an `AppendFile` function that appends content to an existing or new file inside a JD ID folder.

### Behavior

- If file exists, append content (with a leading newline separator if file doesn't end with one).
- If file doesn't exist, create it with the given content.
- Only write to ID-level folders.

### Edge cases

- Empty ref → error.
- Non-ID ref → error.
- ID not found → error.
- Empty filename → error.
- Empty content → no-op, return path without modifying file.
- File doesn't end with newline → add newline before appending.

## Expected behavior

### Function signature

```go
func AppendFile(v *Vault, ref string, filename string, content string) (string, error)
```

Returns the absolute path to the file.

### Test plan

**Unit tests**: Empty ref, non-ID ref, empty filename, ID not found → errors.
**Integration tests**: Append to existing file, create new file, empty content no-op, newline handling.
**Acceptance tests**: Appended content readable via Read, original content preserved.

## Consequences

- Enables incremental note-building without full rewrites.
- CLI: `mimic append S01.12.11 Recipe.md "## Step 3\n..."`.
- MCP: `{ ref, file, content }` params.
