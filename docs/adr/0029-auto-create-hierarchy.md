# ADR-0029: Auto-create area and category in Create

## Status

Accepted

## Context

`Create` currently requires the target category to already exist. If the caller specifies a category ref like `S01.15` and neither the area `S01.10-19` nor the category `S01.15` exist, the call fails. This forces manual folder creation before using the tool, which is friction when bootstrapping new parts of the vault.

## Decision

### 1. Add optional `category_name` and `area_name` parameters to `Create`

```go
func Create(v *Vault, categoryRef string, name string, template string, customVars map[string]string, opts CreateOpts) (*CreateResult, error)

type CreateOpts struct {
    CategoryName string // name for auto-created category (required if category doesn't exist)
    AreaName     string // name for auto-created area (required if area doesn't exist)
}
```

### 2. Auto-creation logic in `Create`

When `findCategory` fails:

1. **Find scope** — must already exist. Error if not found.
2. **Find area** — use `findAreaForCategory`. The area range is derived from the category number: `rangeStart = (catNum / 10) * 10`, `rangeEnd = rangeStart + 9`. If the area doesn't exist:
   - Error if `AreaName` is empty.
   - Create area folder: `S{scope}.{rangeStart}-{rangeEnd} {AreaName}` inside the scope directory.
3. **Create category** — if the category doesn't exist within the (now-existing) area:
   - Error if `CategoryName` is empty.
   - Create category folder: `S{scope}.{catNum} {CategoryName}` inside the area directory.
4. **Re-parse vault** to pick up the new hierarchy, then proceed with normal ID creation.

### 3. MCP tool changes

Add optional `category_name` and `area_name` string parameters to the `create` tool.

### 4. CLI changes

Add `--category-name` and `--area-name` flags to the `create` command.

### Edge cases

- Scope doesn't exist → error (scopes are never auto-created).
- Area exists but category doesn't → only `CategoryName` is required. `AreaName` is ignored.
- Both area and category exist → `CategoryName` and `AreaName` are ignored. Backward compatible.
- Area doesn't exist and `AreaName` is empty → error with message indicating the name is required.
- Category doesn't exist and `CategoryName` is empty → error with message indicating the name is required.
- Category number is in the management range (0-9) → area range is 00-09. Auto-creation works the same way.

### Test plan

**Unit tests:**
- Invalid category ref → error (unchanged).
- Scope not found → error (unchanged).

**Integration tests:**
- Category exists → normal creation, opts ignored (backward compatible).
- Category doesn't exist, area exists, `CategoryName` provided → category auto-created, ID created inside it.
- Category doesn't exist, area exists, `CategoryName` empty → error.
- Neither area nor category exist, both names provided → area and category auto-created, ID created.
- Neither exists, `AreaName` empty → error.
- Neither exists, `CategoryName` empty → error.
- Verify auto-created folders have correct naming convention.

**Acceptance tests:**
- Create ID in non-existent category → re-parse vault → new area, category, and ID all visible.

## Consequences

- `Create` signature changes (new `CreateOpts` parameter). All callers must be updated.
- Existing tests pass unchanged when `CreateOpts` is zero-value (both names empty, category must exist).
- No system IDs (inbox, templates) are auto-created in the new category — the user can add them separately.
