# ADR-0006: Archive JD items

## Status

Accepted

## Context

Users need to archive items following the Johnny Decimal archive hierarchy. Archiving moves items to the appropriate `.09` archive folder at the parent level.

## Decision

Provide an `Archive` function that moves a JD item to its parent's archive folder.

### Archive rules

| Source   | Target                                    | Naming                              |
|----------|-------------------------------------------|-------------------------------------|
| ID       | `S0X.XX.09 Archive for S0X.XX/`          | Renamed to `[Archived] Name`        |
| Category | `S0X.X0.09 Archive for S0X.X0-X9/`      | Keeps ID and all items              |

- Only IDs and categories can be archived. Scopes and areas return an error.
- Archive folder is created if it doesn't exist.

### Edge cases

- Archive folder doesn't exist → create it.
- Item is already inside an archive folder → error.
- Ref not found → error.

## Expected behavior

### Function signature

```go
func Archive(v *Vault, ref string) (*ArchiveResult, error)
```

### Output type

```go
type ArchiveResult struct {
    Ref     string // original ref
    NewPath string // path after archiving
}
```

### Test plan

**Unit tests** (`internal/vault/archive_test.go`):
- Empty ref → error.
- Scope ref → error.
- Area ref → error.
- Invalid ref → error.

**Integration tests** (`internal/vault/archive_integration_test.go`):
- Archive an ID → moved to category's `.09` folder, renamed to `[Archived] Name`.
- Archive a category → moved to area's `.X0.09` folder, keeps ID.
- Archive creates `.09` folder if missing.
- Original path no longer exists after archive.
- Archived item exists at new path.

**Acceptance tests** (`internal/vault/archive_acceptance_test.go`):
- After archiving an ID, re-parsing the vault no longer includes it at the original location.
- Archived ID folder is named `[Archived] Name` (no JD prefix).
- After archiving a category, its contents are preserved at the new location.

## Consequences

- Archive is destructive (moves files). Integration tests use a temporary vault copy.
- Obsidian link updating is out of scope for this ADR — noted as future work.
