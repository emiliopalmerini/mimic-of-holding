# ADR-0032: Stdin and $EDITOR support for write command

## Status

Proposed

## Context

The `write` command requires content as a positional argument, which is awkward for multi-line content. There is no way to pipe content from another command or open an editor for interactive composition. This makes the command slow and cumbersome for human users.

## Decision

Add stdin and `$EDITOR` support to the `write` CLI command. Content is resolved by priority:

1. **Explicit content arg** (existing behavior)
2. **`--template` flag** (existing behavior)
3. **Piped stdin** (new)
4. **`$EDITOR`** (new, interactive terminal only)

### Stdin detection

When content arg is omitted and no `--template` flag, check if stdin is a pipe:

```go
stat, _ := os.Stdin.Stat()
if (stat.Mode() & os.ModeCharDevice) == 0 {
    // stdin is piped, read from it
}
```

### Editor flow

When content arg is omitted, no template, and stdin is a terminal (not piped), open `$EDITOR` (falling back to `$VISUAL`, then `vi`):

1. Create a temp file in `os.TempDir()`
2. Open `$EDITOR` with the temp file path
3. Wait for the editor to exit
4. Read the temp file contents
5. If contents are empty, abort with an error ("empty content, aborting")
6. Clean up the temp file

### Examples

```sh
# Explicit content (unchanged)
mimic write S01.11.11 notes.md "some content"

# Template (unchanged)
mimic write S01.11.11 notes.md --template weekly

# Pipe from stdin (new)
echo "some content" | mimic write S01.11.11 notes.md
cat draft.md | mimic write S01.11.11 notes.md
pbpaste | mimic write S01.11.11 notes.md

# Editor (new, interactive terminal)
mimic write S01.11.11 notes.md
# opens $EDITOR, writes content on save+quit
```

### `--edit` / `-e` flag

When combined with `--template`, the template content is pre-filled in the editor for the user to modify before saving:

```sh
mimic write S01.11.11 review.md --template review --edit
# opens $EDITOR with template content pre-filled
```

Without `--template`, `--edit` forces the editor even if stdin is piped (overrides stdin).

### Args change

`Args` changes from `cobra.RangeArgs(2, 3)` to `cobra.RangeArgs(2, 3)` (unchanged). The third arg remains optional; the new behavior only changes what happens when it is omitted.

### Edge cases

- Stdin is piped but empty (0 bytes): write an empty file (consistent with `echo -n "" | mimic write ...`).
- `$EDITOR` is unset and `$VISUAL` is unset: fall back to `vi`.
- Editor exits with non-zero: abort with error "editor exited with error".
- User saves empty file in editor: abort with error "empty content, aborting".
- Both content arg and stdin piped: content arg wins (priority order).
- `--template` and content arg both provided: content arg wins (existing behavior, unchanged).
- `--edit` without `--template` and content arg provided: content arg wins, `--edit` is ignored.

### Domain layer

No changes. `vault.WriteFile` already accepts a `content string`; all new logic lives in `cmd/mimic/write.go`.

## Test plan

**Unit tests** (in `cmd/mimic/`):
- Content arg takes priority over stdin.
- `--template` takes priority over stdin.
- `--edit` flag is recognized.

**Integration tests** (in `cmd/mimic/`):
- Piped stdin: simulate by setting `cmd.SetIn()` to a `strings.Reader`.
- Verify written file contains stdin content.

**Acceptance tests:**
- Pipe content through stdin, verify file is readable via `mimic read`.

**Not tested** (interactive, requires TTY):
- `$EDITOR` flow (manual testing only).

## Consequences

- Human users get a much faster workflow for writing multi-line content.
- No breaking changes; all existing invocations work identically.
- Editor flow depends on the user's environment (`$EDITOR`); not testable in CI.
