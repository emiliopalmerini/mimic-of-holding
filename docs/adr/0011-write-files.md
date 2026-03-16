# ADR-0011: Write files inside JD IDs

## Status

Accepted

## Context

The MCP server can read vault content but cannot create or update notes inside ID folders. Writing content is the most common vault operation.

## Decision

Provide a `WriteFile` function that creates or overwrites a file inside a JD ID folder.

### Behavior

- Only write to ID-level folders (`S00.00.00`).
- Full file writes only — no partial edits.
- If file exists, overwrite. If not, create.
- Parent ID must exist in the vault.

### Edge cases

- Empty ref → error.
- Non-ID ref → error.
- ID not found → error.
- Empty filename → error.
- Empty content → allowed (creates empty file).

## Expected behavior

### Function signature

```go
func WriteFile(v *Vault, ref string, filename string, content string) (string, error)
```

Returns the absolute path to the written file.

### MCP params

```json
{ "ref": "string (required)", "file": "string (required)", "content": "string (required)" }
```

### Test plan

**Unit tests**:
- Empty ref → error.
- Non-ID ref → error.
- Empty filename → error.
- ID not found → error.

**Integration tests**:
- Write new file → file exists on disk with correct content.
- Overwrite existing file → content updated.
- Write empty content → file exists, is empty.

**Acceptance tests**:
- Written file readable via `Read` with file param.
- Written file appears in `Read` file listing.

## Consequences

- Enables the full create-write workflow: `create S01.12 Pasta` → `write S01.12.11 "Recipe.md" "content"`.
- No partial edits — the LLM must read-then-write for modifications.
