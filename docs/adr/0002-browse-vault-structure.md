# ADR-0002: Browse vault structure

## Status

Accepted

## Context

Users need to visualize the JD hierarchy to navigate and understand what's in the vault. This is the first consumer of the parser from ADR-0001.

## Decision

Provide a `Browse` function that renders the parsed vault tree as an indented, human-readable string. Support optional filtering to focus on a specific scope, area, or category.

### Filter syntax

| Filter        | Example      | Behavior                          |
|---------------|--------------|-----------------------------------|
| None          | `""`         | Show full tree                    |
| Scope         | `S01`        | Show only that scope              |
| Area          | `S01.10-19`  | Show only that area               |
| Category      | `S01.11`     | Show only that category and IDs   |
| Invalid       | `xyz`        | Return error                      |
| No match      | `S99`        | Return error                      |

### Output format

Two-space indentation per level:

```
S01 Me
  S01.00-09 Management for S01
    S01.01 Inbox for S01.00-09
  S01.10-19 Lifestyle
    S01.11 Entertainment
      S01.11.01 Inbox for S01.11
      S01.11.11 Theatre, 2025 Season
```

### Edge cases

- Empty vault (no scopes) → empty string, no error.
- Category with no IDs → category line with no children underneath.
- Filter matches nothing → return error.

## Expected behavior

### Function signature

```go
func Browse(v *Vault, filter string) (string, error)
```

### Test plan

**Unit tests** (`internal/vault/browse_test.go`):
- Empty vault → empty string.
- Single scope with one area, one category, one ID → correct indentation.
- Filter by scope → only that scope's tree.
- Filter by area → only that area's tree.
- Filter by category → only that category and its IDs.
- Invalid filter → error.
- Valid filter syntax but no match → error.

**Integration tests** (`internal/vault/browse_integration_test.go`):
- Full fixture vault with no filter → expected full output.
- Filter `S01` → only S01 tree.
- Filter `S01.10-19` → only Lifestyle area.
- Filter `S01.11` → only Entertainment category and IDs.

**Acceptance tests** (`internal/vault/browse_acceptance_test.go`):
- Full output contains every scope name from the fixture.
- No non-JD entries appear in the output.
- Indentation is consistent (2 spaces per level).
- Filtered output does not contain entries outside the filter.

## Consequences

- Browse becomes the primary navigation tool for both CLI and MCP.
- The filter syntax establishes a convention reusable by other operations (search, read, etc.).
