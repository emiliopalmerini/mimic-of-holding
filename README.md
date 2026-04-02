# Mimic of Holding

A CLI for interacting with a [Johnny Decimal](https://johnnydecimal.com/) organized Obsidian vault.

Built to give [Claude Code](https://claude.ai/code) native access to the [Bag of Holding](https://github.com/emiliopalmerini/bag_of_holding) vault.

## Commands

| Command | Description |
|---------|-------------|
| `mimic browse [filter]` | Display the vault tree (filter by scope/area/category) |
| `mimic search <query>` | Search by JD ref, name, or content (`?query`) |
| `mimic read <ref> [file]` | Read any JD level or a specific file within an ID |
| `mimic resolve <ref> [file]` | Resolve a JD reference to its filesystem path |
| `mimic create <category> <name>` | Create a new JD ID |
| `mimic move <ref> <to>` | Move a JD item to a different parent |
| `mimic rename <ref> <name>` | Rename a JD item (updates wiki links) |
| `mimic move-file <from> <file> <to>` | Move a file between IDs |
| `mimic rename-file <ref> <old> <new>` | Rename a file within an ID (updates wiki links) |
| `mimic frontmatter <action> <ref> <file> <key> <value>` | Edit YAML frontmatter (set, add, remove) |
| `mimic archive <ref>` | Archive an ID or category |
| `mimic inbox [scope]` | List inbox items across scopes |
| `mimic templates <category>` | List available templates for a category |

## Install

```sh
go install ./cmd/mimic
```

## Architecture

```
cmd/mimic/       CLI (Cobra)
internal/vault/  Core domain: parse, browse, search, read, resolve, create, move, rename, archive, frontmatter
docs/adr/        Architecture Decision Records (pre-registered specs)
testdata/        JD vault fixture for tests
```

## Development

Tests follow a pre-registration TDD workflow: ADR spec -> acceptance/unit/integration tests -> implementation.

```sh
go test ./...
```
