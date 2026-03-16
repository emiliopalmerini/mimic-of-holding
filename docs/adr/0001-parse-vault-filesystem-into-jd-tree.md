# ADR-0001: Parse vault filesystem into JD tree

## Status

Accepted

## Context

The foundation of mimic-of-holding is the ability to read an Obsidian vault organized with the Johnny Decimal system and produce a structured, navigable tree. Every other operation (search, create, archive, inbox) depends on this parser.

## Decision

Walk the vault filesystem and produce a typed tree: `Vault → []Scope → []Area → []Category → []ID`.

### Parsing rules

| Level    | Pattern                        | Example                          |
|----------|--------------------------------|----------------------------------|
| Scope    | `S\d{2} .+`                   | `S01 Me`                         |
| Area     | `S\d{2}\.\d{2}-\d{2} .+`     | `S01.10-19 Lifestyle`            |
| Category | `S\d{2}\.\d{2} .+`           | `S01.11 Entertainment`           |
| ID       | `S\d{2}\.\d{2}\.\d{2} .+`    | `S01.11.11 Theatre, 2025 Season` |

### Edge cases

- Non-matching entries at any level are ignored (`.obsidian/`, `Attachments/`, `README.md`, loose files).
- Standard zeros (IDs `.01`-`.09`) are parsed like any other ID but flagged with `IsSystemID: true`.
- Empty areas/categories are valid and represented with zero children.

### Error conditions

- Root path does not exist → return error.
- Root path is a file, not a directory → return error.

## Expected behavior

### Input

A filesystem root path pointing to a JD-organized Obsidian vault.

### Output

```go
type Vault struct {
    Root   string
    Scopes []Scope
}

type Scope struct {
    Number int    // 1, 2, 3
    Name   string // "Me", "Due Draghi", "Work"
    Path   string
    Areas  []Area
}

type Area struct {
    ScopeNumber int
    RangeStart  int    // 10
    RangeEnd    int    // 19
    Name        string // "Lifestyle"
    Path        string
    Categories  []Category
}

type Category struct {
    ScopeNumber int
    Number      int    // 11
    Name        string // "Entertainment"
    Path        string
    IDs         []ID
}

type ID struct {
    ScopeNumber int
    CategoryNum int
    Number      int    // 11
    Name        string // "Theatre, 2025 Season"
    Path        string
    IsSystemID  bool   // true for .01-.09
}
```

### Test plan

**Unit tests** (`internal/vault/parser_test.go`):
- Parse a scope folder name → correct Scope struct.
- Parse an area folder name → correct Area struct.
- Parse a category folder name → correct Category struct.
- Parse an ID folder name → correct ID struct, with IsSystemID flag.
- Non-matching folder names return nil/error and are skipped.

**Integration tests** (`internal/vault/parser_integration_test.go`):
- Parse `testdata/` fixture vault → verify full tree structure.
- Non-existent root → error.
- Root is a file → error.
- Vault with empty areas/categories → valid tree with zero children.

**Acceptance tests** (`internal/vault/parser_acceptance_test.go`):
- Given a realistic vault fixture, `ParseVault()` returns a tree that:
  - Contains the expected number of scopes.
  - Each scope contains the expected areas.
  - Standard zero IDs have `IsSystemID: true`.
  - Non-JD files/folders are absent from the tree.

## Consequences

- All downstream features can rely on a typed, validated tree.
- The parser is the single source of truth for "what exists in the vault."
- Changes to JD naming conventions require updating the parser regex and this ADR.
