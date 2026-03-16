# ADR-0007: List inbox items

## Status

Accepted

## Context

Users need a quick way to see what's sitting in their inboxes across the vault for triage. Inboxes are the `.01` standard zero folders.

## Decision

Provide an `Inbox` function that lists all files inside inbox folders, optionally filtered by scope.

### Inbox identification

- Inbox folders are IDs with number `.01` and "Inbox" in the name.
- Files listed exclude the JDex file (the file named after the folder).

### Edge cases

- No inboxes have items → empty list, no error.
- Invalid scope filter → error.
- Scope filter matches nothing → error.
- Empty scope filter → all scopes.

## Expected behavior

### Function signature

```go
func Inbox(v *Vault, scopeFilter string) ([]InboxItem, error)
```

### Output type

```go
type InboxItem struct {
    InboxRef  string // "S01.11.01"
    InboxName string // "Inbox for S01.11"
    File      string // filename
}
```

### Test plan

**Unit tests** (`internal/vault/inbox_test.go`):
- Invalid scope filter → error.
- Valid scope filter with no matching scope → error.

**Integration tests** (`internal/vault/inbox_integration_test.go`):
- Vault with files in an inbox → returns InboxItems.
- Vault with empty inboxes → empty list.
- Scope filter limits results to that scope.
- JDex files are excluded from results.

**Acceptance tests** (`internal/vault/inbox_acceptance_test.go`):
- Every InboxItem has non-empty InboxRef, InboxName, File.
- InboxRef always ends in `.01`.
- No scope filter returns items from multiple scopes (if present).

## Consequences

- Inbox is read-only and lightweight.
- Provides the entry point for a triage workflow.
