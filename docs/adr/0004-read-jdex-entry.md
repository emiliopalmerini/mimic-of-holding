# ADR-0004: Read JDex entry and contents

## Status

Accepted

## Context

Users need to inspect the contents of a specific JD ID — its JDex metadata file and any other files it contains. This is the primary way to view what's inside a given item.

## Decision

Provide a `Read` function that takes a JD ID reference and returns the JDex file content plus a listing of other files in the folder.

### Resolution rules

- Only ID references (`S\d{2}.\d{2}.\d{2}`) are readable. Scopes, areas, and categories return an error.
- The JDex file is the markdown file named after the parent folder (e.g., `S01.11.11 Theatre, 2025 Season.md`).
- If the JDex file doesn't exist, return an empty JDex string but still list other files.

### Edge cases

- Empty ref → error.
- Ref is not an ID format → error.
- ID not found in vault → error.
- Folder is empty → valid result with empty JDex and empty Files.
- Folder has files but no JDex → empty JDex, files listed.

## Expected behavior

### Function signature

```go
func Read(v *Vault, ref string) (*ReadResult, error)
```

### Output type

```go
type ReadResult struct {
    Ref   string   // "S01.11.11"
    Name  string   // "Theatre, 2025 Season"
    Path  string   // absolute path to folder
    JDex  string   // content of JDex file, empty if missing
    Files []string // other files in the folder (relative names, excluding JDex)
}
```

### Test plan

**Unit tests** (`internal/vault/read_test.go`):
- Empty ref → error.
- Scope ref `S01` → error.
- Area ref `S01.10-19` → error.
- Category ref `S01.11` → error.
- ID ref not found → error.

**Integration tests** (`internal/vault/read_integration_test.go`):
- Read `S01.11.11` from fixture → JDex content matches file, Files is empty.
- Read an ID with no JDex file → empty JDex, no error.
- Read an ID with extra files → Files lists them, JDex still populated.

**Acceptance tests** (`internal/vault/read_acceptance_test.go`):
- ReadResult has non-empty Ref, Name, Path.
- JDex content contains expected text from the fixture file.
- Files list does not include the JDex file itself.

## Consequences

- Read becomes the inspection tool for both CLI and MCP.
- The JDex file naming convention (file named after folder) is enforced by this function.
