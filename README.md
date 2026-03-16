# Mimic of Holding

A CLI and MCP server for interacting with a [Johnny Decimal](https://johnnydecimal.com/) organized Obsidian vault.

Built to give [Claude Code](https://claude.ai/code) native access to the [Bag of Holding](https://github.com/emiliopalmerini/bag_of_holding) vault.

## Commands

| Command | Description |
|---------|-------------|
| `mimic browse [filter]` | Display the vault tree (filter by scope/area/category) |
| `mimic search <query>` | Search by JD ref, name, or content (`?query`) |
| `mimic read <id>` | Read a JDex entry and file listing |
| `mimic create <category> <name>` | Create a new JD ID |
| `mimic archive <ref>` | Archive an ID or category |
| `mimic inbox [scope]` | List inbox items across scopes |

## Install

```sh
go install ./cmd/mimic
go install ./cmd/mimic-mcp
```

## MCP Server

Register with Claude Code:

```sh
claude mcp add mimic-of-holding -- /path/to/mimic-mcp
```

This exposes `browse`, `search`, `read`, `create`, `archive`, and `inbox` as MCP tools.

## Architecture

```
cmd/mimic/       CLI (Cobra)
cmd/mimic-mcp/   MCP server (mcp-go, stdio transport)
internal/vault/  Core domain: parse, browse, search, read, create, archive, inbox
docs/adr/        Architecture Decision Records (pre-registered specs)
testdata/        JD vault fixture for tests
```

## Development

Tests follow a pre-registration TDD workflow: ADR spec → acceptance/unit/integration tests → implementation.

```sh
go test ./...
```
