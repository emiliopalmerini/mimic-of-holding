# ADR-0018: Inbox preview

## Status

Accepted

## Context

Inbox listing shows filenames but not content. Triaging requires a separate read for each file. Adding a preview reduces this to one call.

## Decision

Add a `Preview` field to `InboxItem` containing the first 3 non-empty lines of the file.

### Updated type

```go
type InboxItem struct {
    InboxRef  string
    InboxName string
    File      string
    Preview   string // first 3 non-empty lines, joined with " | "
}
```

### Edge cases

- Binary or unreadable file → empty preview.
- File shorter than 3 lines → show what's available.
- Empty file → empty preview.
- YAML frontmatter (`---` blocks) → skip frontmatter, preview body only.

### Test plan

**Integration tests**: Inbox with files → Preview populated. Preview skips frontmatter. Short files → partial preview.
**Acceptance tests**: Preview never empty for files with body content. Preview doesn't include `---` frontmatter delimiters.

## Consequences

- Slight performance cost (reading file content during inbox listing). Acceptable for personal vaults.
- MCP and CLI output become more useful for triage without extra calls.
