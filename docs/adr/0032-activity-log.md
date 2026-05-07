# ADR-0032: Per-scope activity log

## Status

Proposed

## Context

The vault is intended to function as an LLM-curated knowledge base in the spirit of Karpathy's LLM-Wiki pattern: the user curates, the LLM does the bookkeeping. Today there is no chronological record of what changed in the vault and when. Git history exists, but it is coarse-grained, mixes user edits with LLM edits, and is not addressable from inside the vault.

A dedicated, vault-resident, append-only activity log per scope provides:

- a chronological trail the LLM can read to understand recent state changes before answering queries;
- a substrate the future `ingest` skill (ADR TBD) writes to when it consolidates Blackhole sources into wiki pages;
- a parseable record (grep-friendly H2 headers) for ad hoc inspection.

Scope is **per-vault-scope** (S01, S02, S03), not global, because activity in S02 (podcast) is unrelated to activity in S01 (personal) and bundling them into a single log obscures rather than clarifies.

## Decision

Add a `log` command group to the CLI with two subcommands and an auto-log hook on every mutating operation.

### Log file location

The log lives inside a JD ID, not loose in the area folder. A new scope-level standard zero slot is reserved for this purpose: **`.07 Log`**.

```
S01 Me/S01.00-09 Management for S01/S01.07 Log for S01.00-09/
├── S01.07 Log for S01.00-09.md   ← JDex (describes the ID)
└── log.md                          ← append-only activity data

S02 Due Draghi/S02.00-09 Management for S02/S02.07 Log for S02.00-09/
├── S02.07 Log for S02.00-09.md
└── log.md

S03 Work/S03.00-09 Management for S03/S03.07 Log for S03.00-09/
├── S03.07 Log for S03.00-09.md
└── log.md
```

**Slot choice rationale.** Standard zero slots `.01`–`.04`, `.08`, `.09` are reserved by current convention. `.05` is already used (`S01.05 Scripts`). `.07` is unused across all scopes. Semantically, an append-only chronological record fits near `.08 Someday` and `.09 Archive`, both of which are also temporal/historical.

**JDex content.** The `.07 Log` ID has a normal JDex matching the existing template, with a description pointing readers to `log.md` for actual entries:

```markdown
---
aliases:
  - S01.07 Log for S01.00-09
location: Obsidian
tags:
  - jdex
  - index
---
# S01.07 Log for S01.00-09

Append-only chronological record of vault mutations and skill activity for scope S01. Maintained by `mimic`. See `log.md` in this folder for entries.
```

**Lazy creation.** S02 and S03 currently have no `S0X.00-09 Management for S0X/` area. Mimic creates the full chain (`S0X.00-09 Management for S0X/` → `S0X.07 Log for S0X.00-09/` → JDex + `log.md`) on the first mutation that targets that scope. This reuses the existing auto-create-hierarchy logic (ADR-0029) where applicable.

**CLAUDE.md update.** This ADR also adds a row to the Reserved Slots table in the vault's CLAUDE.md:

| Slot  | Purpose                                                              |
| ----- | -------------------------------------------------------------------- |
| `.07` | Log (scope-level only) — append-only activity record managed by mimic |

Category-level `.07` remains free for future use; logs are scope-level only.

### Entry format

Each entry is a markdown H2 header followed by an optional details block:

```markdown
## [2026-05-07 14:32] create | S01.21.14 AI Patterns

Created JDex.
```

```markdown
## [2026-05-07 14:35] archive | S01.11.15 Theatre → S01.11.09/[Archived] Theatre

Item-level archive. 4 files moved.
```

```markdown
## [2026-05-07 15:02] ingest | Karpathy LLM Wiki

Source: Notion Blackhole. Touched: S01.21.11 CSharp, S01.21.14 AI Patterns.
```

Header grammar:

```
## [YYYY-MM-DD HH:MM] <op> | <target> [→ <secondary-target>]
```

- `op`: short verb. Set: `create`, `archive`, `move`, `move-file`, `rename`, `rename-file`, `frontmatter`, `ingest`, `lint` (future). Lowercase.
- `target`: the primary JD ref or human-readable label affected.
- `secondary-target`: optional, used for ops with src→dst semantics (archive, move, rename).
- `details` body: one or more lines of plain markdown after a blank line. Optional.

Timestamps are local time in 24h. ISO date.

### Auto-logging on mutations

The following commands auto-append a log entry on success (after the mutation completes, before exit):

| Command       | Op            | Target                                | Details                                          |
| ------------- | ------------- | ------------------------------------- | ------------------------------------------------ |
| `create`      | `create`      | created JD ref                        | `Created JDex.` (and any auto-created hierarchy) |
| `archive`     | `archive`     | `<src> → <dst>`                       | level (file/id/category/area), file count        |
| `move`        | `move`        | `<src ref> → <dst ref>`               | —                                                |
| `move-file`   | `move-file`   | `<src ref>/<file> → <dst ref>/<file>` | —                                                |
| `rename`      | `rename`      | `<old name> → <new name>`             | scope of rename (id/category/area)               |
| `rename-file` | `rename-file` | `<ref>/<old> → <ref>/<new>`           | —                                                |
| `frontmatter` | `frontmatter` | `<ref>/<file>`                        | field(s) changed                                 |

Read-only commands (`browse`, `read`, `search`, `inbox`, `resolve`, `templates`, `log tail`) do **not** log.

If a mutation fails (returns non-zero), nothing is logged.

For cross-scope moves, the entry is written to the **destination scope's** log only. Rationale: the action's final state belongs to the destination; the source no longer holds the item. (If we later find this loses signal, we can add a back-reference in the source log; not now.)

### `mimic log append`

Public command for skills and scripts that perform writes outside mimic (e.g. the future `ingest` skill, which uses `Edit` on JDex pages).

```
mimic log append <scope> <op> <target> [--secondary <ref>] [--details <text>]
```

**Inputs:**

- `scope` (required): `S01`, `S02`, `S03`.
- `op` (required): one of the op set above. Free-form strings rejected to keep the log canonical.
- `target` (required): label or JD ref. Quoted if it contains spaces.
- `--secondary` (optional): adds `→ <secondary>` to the header.
- `--details` (optional): body text. Can include newlines (caller's responsibility to escape if invoked from shell).

**Examples:**

```sh
mimic log append S01 ingest "Karpathy LLM Wiki" \
  --details "Source: Blackhole. Touched: S01.21.11, S01.21.14."

mimic log append S02 lint "S02.10-19" \
  --details "3 orphan files reported."
```

**Output:** none on success. Non-zero exit on validation error.

### `mimic log tail`

Read the recent N entries of a scope's log.

```
mimic log tail <scope> [-n <N>]
```

**Inputs:**

- `scope` (required): `S01`, `S02`, `S03`.
- `-n` (optional, default 10): number of entries to return.

**Output (stdout):** the last N entries verbatim, in chronological order (oldest of the N first), separated by blank lines.

If the log file does not exist, output empty and exit 0.

### Domain functions

In `internal/vault/`:

```go
// Log appends an entry to the scope's log file. Creates the file if missing.
// op is required. secondary and details may be empty.
func Log(v *Vault, scope, op, target, secondary, details string) error

// LogTail returns the last n entries from a scope's log. Returns an empty
// slice if the log file does not exist. Each entry is the raw markdown block
// (header + body) without trailing blank line.
func LogTail(v *Vault, scope string, n int) ([]string, error)
```

CLI handlers in `cmd/mimic/log.go` are thin adapters.

Mutating CLI commands call `vault.Log(...)` after their domain function returns nil error. The op label and target/secondary/details strings are determined by each command (no central dispatcher).

### Edge cases

- **Log file missing**: `Log` creates it with a single H1 header `# Activity Log — <scope>` then the entry.
- **Scope folder missing**: `Log` returns error `scope not found`. Mutations should never trigger this since they operate on existing scopes; the case applies to manual `log append`.
- **Op not in canonical set**: `log append` rejects with error. Auto-log calls always pass canonical ops.
- **`tail` with N larger than entry count**: returns all available entries, no error.
- **`tail` with N <= 0**: returns empty.
- **Concurrent invocations**: out of scope. Mimic is single-shot CLI; if two processes append simultaneously, last writer wins for that file write. Acceptable; the user does not run mimic concurrently.
- **Mutation fails partway** (e.g. archive moves 3 of 4 files then errors): no log entry. The caller surfaces the partial state via the error; the log stays clean.
- **Auto-log itself fails** (disk full, permissions): the mutation has already succeeded. Log failure is reported on stderr as `warning: log append failed: <err>` and the command still exits 0. We do not undo a successful mutation because logging failed.

### What is NOT changing

- Existing commands keep their current outputs and exit codes; auto-logging is additive.
- Log files are not parsed back by any other command yet (no `log query`, no log-driven undo). Future ADRs may add these.
- Git is unaffected. Log files are committed with normal vault commits.

## Consequences

- Three new files appear in the vault (one per scope's Management area). User must initialize them lazily (created on first mutation per scope).
- The `ingest` skill (future ADR) gets a clean substrate to write to without re-implementing log mechanics.
- The `lint` command (future ADR) can use the log as one of its inputs (e.g. flag pages that have not been touched in N months).
- Vault commits will include log churn. Acceptable; the log tells the story of the vault and is part of its content.
- Adding a new mutating command in the future requires the author to remember to call `vault.Log`. Mitigation: a comment in the domain layer pointing at this ADR.
