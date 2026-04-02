# ADR-0031: Drop MCP server, add resolve command

## Status

Accepted

## Context

The MCP server (`cmd/mimic-mcp/`) exposes 17 tools, many of which duplicate functionality that AI agents already have natively (Read, Write, Edit via Claude Code). Passing file content through MCP tool parameters is slow because the protocol requires the full string before execution (no streaming). The agent ends up crafting large content strings as MCP params when it could write files directly with its own optimized tools.

The CLI (`cmd/mimic/`) already covers all domain operations and is callable via Bash by any agent with shell access. Maintaining two adapters (CLI + MCP) over the same domain doubles the surface area with no added value when the consumer always has shell access.

The missing piece is path resolution: the agent doesn't know where `S01.11.11` maps on the filesystem, so it can't use native file tools without help.

## Decision

1. **Remove** `cmd/mimic-mcp/` entirely.
2. **Add** a `resolve` CLI command that translates JD references to filesystem paths.
3. **Keep** all existing CLI commands unchanged.

### `resolve` command

```
mimic resolve <ref> [file]
```

**Inputs:**

- `ref` (required): any JD reference (scope, area, category, or ID). Examples: `S01`, `S01.10-19`, `S01.11`, `S01.11.11`.
- `file` (optional): filename within an ID. Only valid when `ref` is an ID.

**Output (stdout):**

- The absolute filesystem path to the resolved item, or to the file within it.
- Exit code 0 on success, non-zero on error.

**Examples:**

```sh
$ mimic resolve S01.11.11
/home/user/Documents/bag_of_holding/S01 Me/S01.10-19 Lifestyle/S01.11 Entertainment/S01.11.11 Theatre, 2025 Season

$ mimic resolve S01.11.11 notes.md
/home/user/Documents/bag_of_holding/S01 Me/S01.10-19 Lifestyle/S01.11 Entertainment/S01.11.11 Theatre, 2025 Season/notes.md

$ mimic resolve S01.11
/home/user/Documents/bag_of_holding/S01 Me/S01.10-19 Lifestyle/S01.11 Entertainment
```

### Agent workflow (before vs. after)

**Before (MCP):**

```
agent calls MCP write(ref="S01.11.11", file="review.md", content="<entire file content>")
```

**After (CLI + native tools):**

```
agent runs: mimic resolve S01.11.11 review.md
agent gets: /home/.../S01.11.11 Theatre, 2025 Season/review.md
agent uses native Write tool on that path (streaming, fast)
```

### Domain function

```go
// Resolve returns the absolute filesystem path for a JD reference.
// If file is non-empty and ref is an ID, returns the path to the file within it.
func Resolve(v *Vault, ref string, file string) (string, error)
```

### Edge cases

- `ref` is empty: return error "empty reference".
- `ref` does not match any vault item: return error "not found".
- `file` provided with a non-ID ref (scope, area, category): return error "file argument is only valid for ID references".
- `file` does not exist on disk: return the path anyway (the agent may be about to create it). This matches `resolve` semantics (path resolution, not existence check).
- `ref` matches multiple items (ambiguous): not possible in JD; references are unique.

### What is NOT changing

- All existing CLI commands (`browse`, `search`, `read`, `write`, `edit`, `append`, `create`, `archive`, `move`, `rename`, `move_file`, `rename_file`, `inbox`, `templates`) remain unchanged. They are still useful for human users in the terminal.
- The `internal/vault/` domain layer is unchanged.

## Consequences

- The MCP server code (`cmd/mimic-mcp/`, `server.go`, `mcp_test.go`, `main.go`) is deleted.
- Any MCP client configuration pointing to `mimic-mcp` must be removed.
- Agents using this tool must have shell access (Bash) to call the CLI.
- Agents that previously used MCP for file I/O will be faster, since they use native streaming tools instead of passing content through MCP params.
- One adapter to maintain instead of two.
