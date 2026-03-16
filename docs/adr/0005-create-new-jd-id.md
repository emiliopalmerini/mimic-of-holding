# ADR-0005: Create new JD ID

## Status

Accepted

## Context

Users need to create new items in the vault following the JDex workflow: find the next available ID, create the folder, and generate the JDex entry file with standard frontmatter.

## Decision

Provide a `Create` function that takes a category reference and a name, and creates the next available ID folder with a JDex file.

### ID numbering rules

- Regular IDs start at `.11` (`.01-.09` are reserved for system zeros).
- Next ID = highest existing regular ID + 1.
- If no regular IDs exist, start at `.11`.

### JDex template

```yaml
---
aliases:
  - S01.11.12 Cinema
location: Obsidian
tags:
  - jdex
  - index
---
# S01.11.12 Cinema

## Contents
```

### Edge cases

- Category has no regular IDs → first ID is `.11`.
- Category has only system IDs → first regular ID is `.11`.
- Empty name → error.
- Invalid category ref → error.
- Category not found → error.

## Expected behavior

### Function signature

```go
func Create(v *Vault, categoryRef string, name string) (*CreateResult, error)
```

### Output type

```go
type CreateResult struct {
    Ref  string // "S01.11.12"
    Name string // "Cinema"
    Path string // absolute path to created folder
}
```

### Test plan

**Unit tests** (`internal/vault/create_test.go`):
- Empty name → error.
- Invalid category ref (e.g., `S01`, `xyz`) → error.
- Category not found (`S99.99`) → error.

**Integration tests** (`internal/vault/create_integration_test.go`):
- Create in category with existing IDs → next number assigned correctly.
- Create in empty category → ID is `.11`.
- Create in category with only system IDs → ID is `.11`.
- Folder and JDex file exist on disk after creation.
- JDex file has correct frontmatter content.

**Acceptance tests** (`internal/vault/create_acceptance_test.go`):
- Created folder name matches `S\d{2}\.\d{2}\.\d{2} .+` pattern.
- JDex file is named after the folder.
- JDex file contains aliases, location, tags, and heading.
- Re-parsing the vault after creation includes the new ID.

## Consequences

- Create is a write operation — integration tests need a temporary copy of the fixture vault.
- The JDex template is hardcoded; could be made configurable later if needed.
