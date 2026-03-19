# ADR-0025: Edit file (search-and-replace)

## Status

Proposed

## Context

The `write` tool requires the LLM to emit the entire file content as a string parameter. For modifications to existing files, this means the agent must regenerate every character — O(file size) output tokens — even to change a single line. This is the dominant source of latency for write-heavy workflows.

ADR-0011 noted: "No partial edits — the LLM must read-then-write for modifications." This ADR removes that limitation.

## Decision

Add an `EditFile` function that performs exact string replacement within an existing file, analogous to a search-and-replace. The agent sends only the old and new fragments — O(diff) output tokens instead of O(file).

### Function signature

```go
// EditFile replaces the first occurrence of oldString with newString in a file
// within a JD ID folder. The file must exist. oldString must appear exactly
// once in the file (to prevent ambiguous edits). Returns the absolute path.
func EditFile(v *Vault, ref, filename, oldString, newString string) (string, error)
```

### MCP tool integration

Register as a new `edit` tool:

```json
{
  "ref": "string (required) — JD ID reference (e.g., S01.11.11)",
  "file": "string (required) — filename to edit",
  "old_string": "string (required) — exact text to find",
  "new_string": "string (required) — replacement text"
}
```

### CLI integration

```
mimic edit <id> <filename> <old_string> <new_string>
```

### Uniqueness constraint

`oldString` must occur exactly once in the file. This prevents ambiguous edits where the caller intended to change one occurrence but the string appears multiple times. If the caller needs to replace all occurrences, they should use `write` with full content.

### Edge cases

- Empty ref → error.
- Non-ID ref → error.
- ID not found → error.
- Empty filename → error.
- File does not exist → error.
- Empty old_string → error.
- old_string not found in file → error.
- old_string found more than once → error ("ambiguous edit: old_string appears N times").
- old_string == new_string → no-op, return path.
- Empty new_string → allowed (deletes the matched text).

### Test plan

**Unit tests**:
- Empty ref → error.
- Non-ID ref → error.
- Empty filename → error.
- Empty old_string → error.
- old_string == new_string → no-op.

**Integration tests**:
- Replace text in existing file → content updated on disk.
- old_string not found → error, file unchanged.
- old_string appears twice → error, file unchanged.
- Empty new_string → matched text deleted.
- Multiline old_string/new_string → works correctly.
- Edit preserves rest of file content unchanged.

**Acceptance tests**:
- Edit a file, read it back via `Read` → updated content returned.

## Consequences

- Reduces agent output tokens from O(file) to O(diff) for the common modify-existing-file case.
- Adds one MCP tool (`edit`), one CLI command.
- The uniqueness constraint means callers must provide enough context in old_string to be unambiguous. This is the same trade-off Claude Code's own Edit tool makes.
