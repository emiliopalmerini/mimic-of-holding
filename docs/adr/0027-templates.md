# ADR-0027: Template support for file creation

## Status

Accepted

## Context

When creating new IDs (`Create`) or files (`WriteFile`), the content is either hardcoded (JDex boilerplate) or supplied by the caller. The vault already has a convention for templates: each category has a `.03` ID (`SXX.YY.03 Templates for SXX.YY`) containing `.md` template files (e.g., `Recipe Template.md`, `Lazy GM Session Prep.md`).

Templates may contain Obsidian-style `{{variables}}`. Some of these are domain-specific placeholders left for the user to fill in (e.g., `{{porzioni}}`), while others should be automatically substituted at creation time (e.g., `{{ref}}`, `{{name}}`, `{{date}}`).

## Decision

### 1. `ListTemplates` — discover available templates

```go
func ListTemplates(v *Vault, categoryRef string) ([]TemplateInfo, error)

type TemplateInfo struct {
    Name     string // filename stem, e.g. "Recipe Template"
    Filename string // full filename, e.g. "Recipe Template.md"
    Path     string // absolute path
    Source   string // which level it came from: "category", "area", "scope"
    SourceRef string // ref of the templates ID, e.g. "S01.12.03"
}
```

Lookup is hierarchical (bottom-up). For a given category `SXX.YY`:

1. **Category**: find ID `.03` in category `SXX.YY` → list non-JDex `.md` files
2. **Area**: find ID `.03` in the area's management category (`SXX.X0`) → list non-JDex `.md` files
3. **Scope**: find ID `.03` in the scope's management area's management category (`SXX.01`) → list non-JDex `.md` files

Results from all levels are returned, with closer levels listed first. If the same filename appears at multiple levels, only the closest one is returned (category shadows area shadows scope).

A non-JDex file is any `.md` file whose stem does not match the folder name.

#### CLI

```
mimic templates <category-ref>
```

Output:
```
Templates for S01.12:
  Recipe Template (S01.12.03, category)
  Default Note (S01.10.03, area)
```

#### MCP tool

```json
{
  "name": "templates",
  "parameters": {
    "category": "string (required) — category ref (e.g. S01.12)"
  }
}
```

### 2. Modify `Create` — optional template for JDex

Add an optional `template` parameter. When provided, the named template is resolved via the same hierarchical lookup, read, and used as the JDex file content (with variable substitution). When omitted, the current hardcoded JDex boilerplate is used (backward compatible).

```go
func Create(v *Vault, categoryRef string, name string, template string) (*CreateResult, error)
```

#### CLI

```
mimic create <category> <name> [--template <template-name>]
```

#### MCP tool

Add optional `template` string parameter to existing `create` tool.

### 3. Modify `WriteFile` — optional template for new files

Add an optional `template` parameter. When provided and `content` is empty, the named template is resolved, read, and used as the file content (with variable substitution). When `content` is non-empty, it is used as-is (template ignored).

```go
func WriteFile(v *Vault, ref string, filename string, content string, template string) (string, error)
```

#### CLI

```
mimic write <id> <filename> [content] [--template <template-name>]
```

Content becomes optional when `--template` is provided.

#### MCP tool

Add optional `template` string parameter to existing `write` tool.

### 4. Template variable substitution

Only a known set of variables is substituted. All other `{{...}}` patterns are left untouched for the user to fill in manually.

| Variable | Value | Example |
|----------|-------|---------|
| `{{ref}}` | JD reference | `S01.12.11` |
| `{{name}}` | human-readable name | `Carbonara` |
| `{{title}}` | full JD title | `S01.12.11 Carbonara` |
| `{{date}}` | today's date (YYYY-MM-DD) | `2026-03-26` |

```go
func ApplyTemplate(content string, vars TemplateVars) string

type TemplateVars struct {
    Ref   string
    Name  string
    Title string
    Date  string
}
```

### 5. Template resolution

`resolveTemplate(v *Vault, categoryRef string, templateName string) (string, error)` finds the template file by name (stem match, case-insensitive) using the hierarchical lookup described above. Returns the file content. Error if not found at any level.

### Edge cases

- `ListTemplates`: invalid category ref → error. Category not found → error. No templates found → empty list (not an error).
- `Create` with template: template not found → error (folder is not created). Template read fails → error.
- `WriteFile` with template: template not found → error. Both content and template provided → content wins, template ignored.
- `WriteFile` with neither content nor template → error (empty content not allowed without a template).
- Template file contains no known `{{variables}}` → file is copied verbatim (no error).
- `.03` ID does not exist at a given level → skip that level silently.

### Test plan

**Unit tests (ListTemplates)**:
- Invalid category ref → error.
- Category not found → error.

**Unit tests (ApplyTemplate)**:
- All known variables substituted correctly.
- Unknown `{{variables}}` left untouched.
- No variables in content → content unchanged.

**Integration tests (ListTemplates)**:
- Category with templates in `.03` → templates listed.
- Category without `.03` → empty list.
- Templates at multiple levels → all returned, closest first, shadowing by filename.

**Integration tests (Create with template)**:
- Template found → JDex uses template content with variables substituted.
- Template not found → error, no folder created.
- No template param → default hardcoded JDex (backward compatible).

**Integration tests (WriteFile with template)**:
- Template found, empty content → file uses template with variables substituted.
- Template found, content provided → content used, template ignored.
- Template not found → error.

**Acceptance tests**:
- Create ID with template → re-parse, read JDex → template content with substituted variables.
- Write file with template → read file → template content with substituted variables.

## Consequences

- `Create` and `WriteFile` signatures change (new optional parameter). All callers (CLI, MCP) must be updated.
- Existing tests for `Create` and `WriteFile` must pass unchanged (template = "" preserves current behavior).
- The `.03` convention is now load-bearing — the tool depends on it.
- `ListTemplates` gives Claude the ability to discover templates before using them, avoiding guesswork.
