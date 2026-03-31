# ADR-0028: Custom template variables

## Status

Accepted

## Context

ADR-0027 introduced template support for `Create` and `WriteFile`, with four built-in variables (`{{ref}}`, `{{name}}`, `{{title}}`, `{{date}}`). All other `{{...}}` placeholders are left untouched for the user to fill in manually.

In practice, vault templates contain many domain-specific placeholders (e.g., `{{description}}`, `{{porzioni}}`, `{{company}}`, `{{role}}`). When an LLM uses the MCP `write` tool with a template, it must follow up with multiple `edit` or `frontmatter` calls to fill these in — making the workflow slow and clunky.

## Decision

### 1. Add `CustomVars` to `TemplateVars`

```go
type TemplateVars struct {
    Ref        string
    Name       string
    Title      string
    Date       string
    CustomVars map[string]string // NEW
}
```

### 2. Modify `ApplyTemplate` — substitute custom vars after built-ins

After substituting built-in variables, iterate over `CustomVars` and replace `{{key}}` → `value`. Built-in keys (`ref`, `name`, `title`, `date`) take precedence — if a custom var collides with a built-in key, the custom var is ignored.

```go
func ApplyTemplate(content string, vars TemplateVars) string
```

Empty string values are valid and replace the placeholder with an empty string.
Unknown custom vars (key not present in template) are harmlessly ignored.

### 3. Modify `WriteFile` — accept custom vars

```go
func WriteFile(v *Vault, ref string, filename string, content string, template string, customVars map[string]string) (string, error)
```

When `template` is provided, `customVars` are merged into the `TemplateVars` before applying. When `content` is non-empty, `customVars` are ignored (content takes precedence).

### 4. Modify `Create` — accept custom vars

```go
func Create(v *Vault, categoryRef string, name string, template string, customVars map[string]string) (*CreateResult, error)
```

Same behavior: `customVars` only apply when a template is used.

### 5. MCP tool changes

Add optional `vars` object parameter (JSON object with string values) to `write` and `create` tools.

### 6. CLI changes

Add `--var key=value` repeatable flag to `write` and `create` commands. Multiple `--var` flags build the map.

### Edge cases

- `customVars` is `nil` or empty → no additional substitution (backward compatible).
- Custom var key collides with built-in key → built-in wins, custom var skipped.
- Custom var with empty string value → placeholder replaced with empty string.
- Custom var key not present in template → ignored (no error).
- `vars` provided without `template` → ignored (no error).
- `vars` provided with `content` → ignored (content takes precedence).

### Test plan

**Unit tests (ApplyTemplate)**:
- Custom vars substituted correctly.
- Built-in keys take precedence over custom vars with same key.
- Empty custom vars map → no change (backward compatible).
- Empty string value replaces placeholder with empty string.
- Custom var key not in template → content unchanged.

**Integration tests (WriteFile)**:
- Template + custom vars → placeholders replaced.
- Template + no custom vars → domain-specific placeholders left untouched (backward compatible).
- Content + custom vars → content used as-is, vars ignored.

**Integration tests (Create)**:
- Template + custom vars → JDex has substituted content.
- No template → default boilerplate unchanged (backward compatible).

**Acceptance tests**:
- Write with template + vars → read → content matches expected.
- Create with template + vars → re-parse + read → content matches expected.

## Consequences

- `WriteFile` and `Create` signatures change (new parameter). All callers must be updated.
- Existing tests pass unchanged when `customVars` is `nil`.
- Template authors can now use arbitrary `{{placeholder}}` names knowing the LLM can fill them in via `vars`.
