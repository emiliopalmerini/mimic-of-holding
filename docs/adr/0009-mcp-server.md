# ADR-0009: MCP server

## Status

Accepted

## Context

The MCP server exposes the same vault operations as the CLI but as MCP tools, allowing Claude Code to call them natively without shelling out.

## Decision

Create a `mimic-mcp` binary that serves MCP tools over stdio using `github.com/mark3labs/mcp-go`.

### Tools

| Tool      | Parameters                       | Returns                          |
|-----------|----------------------------------|----------------------------------|
| `browse`  | `filter?: string`               | Tree string                      |
| `search`  | `query: string`                 | Array of search results          |
| `read`    | `ref: string`                   | JDex content + file listing      |
| `create`  | `category: string, name: string`| Created ref, name, path          |
| `archive` | `ref: string`                   | Original ref, new path           |
| `inbox`   | `scope?: string`                | Array of inbox items             |

### Configuration

- Vault path defaults to `~/Documents/bag_of_holding`.
- Overridable via `--vault` flag.

### Test plan

**Integration tests** (`cmd/mimic-mcp/mcp_test.go`):
- Each tool tested by calling the handler directly with `--vault` pointing to testdata fixture.
- `browse` → returns tree containing scope names.
- `browse` with filter → returns filtered tree.
- `search` by name → returns matching results.
- `search` by ref → returns exact match.
- `search` by content → returns content match.
- `search` missing query → error.
- `read` → returns JDex content.
- `read` invalid ref → error.
- `create` → returns new ref.
- `archive` → returns archived confirmation.
- `inbox` → returns inbox items.
- `inbox` with scope → filtered items.

## Consequences

- MCP server is a thin adapter layer, same as the CLI.
- Claude Code can be configured to use this server via `claude mcp add`.
