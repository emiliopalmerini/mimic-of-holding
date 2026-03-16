# ADR-0010: Expand read to all JD levels + individual files

## Status

Accepted

## Context

The current `Read` function only accepts ID-level references (`S00.00.00`). An LLM using the MCP server constantly tries to read scopes, areas, and categories and gets rejected. Additionally, `Read` lists files inside an ID but provides no way to read their content.

## Decision

Expand `Read` to accept any JD reference level and add an optional file parameter to read individual files within an ID.

### Behavior by level

| Input | Output |
|-------|--------|
| `S01` | Scope summary: name, path, list of areas with category counts |
| `S01.10-19` | Area summary: name, path, list of categories with ID counts |
| `S01.12` | Category summary: name, path, list of IDs with names |
| `S01.12.11` | ID detail: JDex content + file listing (current behavior) |
| `S01.12.11` + file `Amatriciana V1.md` | Content of that specific file |

### Updated types

```go
type ReadResult struct {
    Type     string   // "scope", "area", "category", "id", "file"
    Ref      string
    Name     string
    Path     string
    Content  string   // JDex content for IDs, file content for files, summary for higher levels
    Files    []string // only populated for ID-level reads
    Children []string // populated for scope/area/category (area names, category names, ID names)
}
```

### Edge cases

- File param with non-ID ref → error (files only exist inside IDs).
- File not found inside ID → error.
- File param is the JDex file itself → read it normally.

### MCP params

```json
{ "ref": "string (required)", "file": "string (optional)" }
```

### Test plan

**Unit tests**:
- Scope ref → valid, Type="scope".
- Area ref → valid, Type="area".
- Category ref → valid, Type="category".
- ID ref → valid, Type="id" (existing behavior).
- File param with scope ref → error.
- Empty ref → error.

**Integration tests**:
- Read scope from fixture → Children lists areas.
- Read area → Children lists categories.
- Read category → Children lists IDs.
- Read ID → JDex + files (unchanged).
- Read ID + file → file Content populated.
- Read ID + nonexistent file → error.

**Acceptance tests**:
- Every level returns correct Type.
- Children at each level match the actual fixture structure.
- File content matches what's on disk.

## Consequences

- `Read` replaces the need for separate "summary" commands.
- The LLM can navigate top-down: browse → read scope → read category → read ID → read file.
- Breaking change to `ReadResult` struct — CLI and MCP handlers need updating.
