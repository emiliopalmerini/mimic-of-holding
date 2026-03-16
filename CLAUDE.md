# CLAUDE.md

## Project Overview

Mimic of Holding is a Go CLI + MCP server for interacting with a Johnny Decimal organized Obsidian vault (Bag of Holding).

## Architecture

- `internal/vault/` — Core domain. All business logic lives here. No external dependencies.
- `cmd/mimic/` — Cobra CLI. Thin adapter over domain functions.
- `cmd/mimic-mcp/` — MCP server via `mcp-go`. Thin adapter over domain functions.
- `docs/adr/` — Architecture Decision Records. Each feature has a pre-registered spec.
- `testdata/vault/` — JD vault fixture used by all tests.

## Development Workflow

This project uses pre-registration TDD:

1. Write an ADR in `docs/adr/` defining inputs, outputs, edge cases
2. Write tests: acceptance → unit → integration
3. Verify tests fail
4. Implement to make tests pass
5. Mark ADR as Accepted

## Commands

```sh
go test ./...              # Run all tests
go build ./cmd/mimic       # Build CLI
go build ./cmd/mimic-mcp   # Build MCP server
```

## Vault Path

Default vault path: `~/Documents/bag_of_holding`. Override with `--vault` flag (CLI) or `--vault` arg (MCP binary).

## Version Control

This project uses **jj (Jujutsu)** for version control (colocated with git).
