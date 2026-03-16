# ADR-0008: CLI layer with Cobra

## Status

Accepted

## Context

The CLI is the first user-facing interface for mimic-of-holding. It wraps the core domain operations (ADR-0001 through ADR-0007) in a Cobra command tree.

## Decision

Create a `mimic` binary with subcommands that map 1:1 to domain operations.

### Commands

| Command                          | Domain function | Description                     |
|----------------------------------|-----------------|---------------------------------|
| `mimic browse [filter]`         | `Browse`        | Tree view, optional filter      |
| `mimic search <query>`          | `Search`        | By JD ref, name, or `?content`  |
| `mimic read <id>`               | `Read`          | JDex entry + file listing       |
| `mimic create <category> <name>`| `Create`        | Create new ID                   |
| `mimic archive <ref>`           | `Archive`       | Archive ID or category          |
| `mimic inbox [scope]`           | `Inbox`         | List inbox items                |

### Configuration

- Vault path hardcoded to `~/Documents/bag_of_holding`.
- Overridable via `--vault` flag for testing.

### Output

- Plain text to stdout.
- Errors to stderr with exit code 1.

### Test plan

**Integration tests** (`cmd/mimic/cmd_test.go`):
- Each command tested with `--vault` pointing to testdata fixture.
- `browse` → output contains scope names.
- `browse S01` → output contains S01, not S02.
- `search Entertainment` → output contains category result.
- `search S01.11.11` → output contains ID result.
- `search ?Italian` → output contains content match.
- `read S01.11.11` → output contains JDex content.
- `read S01` → error exit.
- `create S01.12 "Pasta"` → success output with new ref.
- `archive S01.11.11` → success output.
- `inbox` → output contains inbox items.
- `inbox S01` → filtered output.
- Missing args → error exit.

## Consequences

- CLI is a thin layer — no business logic, just argument parsing and output formatting.
- `--vault` flag enables testability without mocking.
