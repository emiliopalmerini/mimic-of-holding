# ADR-0016: Move and rename JD items

## Status

Accepted

## Context

Reorganizing vault entries currently requires manual filesystem operations and Obsidian. The tool should support moving items between parents and renaming them.

## Decision

Provide `Rename`, `Move`, and `MoveFile` functions.

### Rename

Changes the human-readable name of any JD item (scope, area, category, ID). The JD number stays the same.

- Renames the folder on disk.
- For IDs: renames the JDex file and updates its frontmatter (aliases, heading).
- For categories/areas: recursively updates child folder prefixes and JDex files.

```go
func Rename(v *Vault, ref string, newName string) (*RenameResult, error)

type RenameResult struct {
    Ref     string
    OldName string
    NewName string
    OldPath string
    NewPath string
    LinksUpdated int // wiki links updated (ADR-0017)
}
```

### Move

Relocates a JD item to a different parent.

| Source | Target | Behavior |
|--------|--------|----------|
| ID (`S01.11.11`) | Category (`S01.12`) | Gets next available number in target category |
| Category (`S01.11`) | Area (`S01.20-29`) | Keeps number if available, else next available |
| Area (`S01.10-19`) | Scope (`S02`) | Keeps range if available |

- Updates folder name prefix to match new parent.
- For categories/areas: recursively updates all child prefixes.
- Updates JDex files (aliases, heading) for moved item and all descendants.

```go
func Move(v *Vault, ref string, to string) (*MoveResult, error)

type MoveResult struct {
    OldRef  string
    NewRef  string
    OldPath string
    NewPath string
    LinksUpdated int
}
```

### MoveFile

Moves a file from one ID to another.

```go
func MoveFile(v *Vault, fromRef string, filename string, toRef string) (string, error)
```

Returns new file path.

### Edge cases

- Rename: empty name → error. Ref not found → error.
- Move: target not found → error. Target is same parent → error. Number collision handled by auto-assignment.
- MoveFile: file not found → error. Source or target ID not found → error.

### Test plan

**Unit tests**: Error cases for all three functions.

**Integration tests (Rename)**:
- Rename an ID → folder renamed, JDex file renamed, frontmatter updated.
- Rename a category → folder renamed, child ID prefixes updated.

**Integration tests (Move)**:
- Move ID to different category → new number assigned, prefix updated.
- Move category to different area → prefix updated, children updated.

**Integration tests (MoveFile)**:
- Move file between IDs → file at new location, gone from old.

**Acceptance tests**:
- After rename, vault re-parses correctly with new name.
- After move, item findable at new ref via Search.

## Consequences

- Most complex operation — recursive prefix updates across children.
- Depends on ADR-0017 for wiki link updates.
- Move invalidates the current Vault struct — callers should re-parse after.
