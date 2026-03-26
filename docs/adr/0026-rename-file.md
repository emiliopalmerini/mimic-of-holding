# ADR-0026: Rename file within a JD ID

## Status

Accepted

## Context

`Rename` (ADR-0016) changes the human-readable name of a JD item (scope, area, category, ID) and updates wikilinks. There is no way to rename an individual `.md` file inside an ID folder. Renaming a file manually leaves stale wikilinks throughout the vault and, if the file is the JDex file, creates a mismatch between the folder name and its index file.

## Decision

Add a `RenameFile` function and expose it as CLI command `mimic rename-file` and MCP tool `rename_file`.

### Function signature

```go
func RenameFile(v *Vault, ref, oldFilename, newFilename string) (*RenameFileResult, error)

type RenameFileResult struct {
    OldPath      string
    NewPath      string
    LinksUpdated int
    HeadingUpdated bool
    IsJDex       bool   // true if the renamed file was the JDex file
    FolderRenamed bool  // true if the folder was also renamed (JDex case)
}
```

### Behavior

1. **Resolve** `ref` to a JD ID. Non-ID refs are an error.
2. **Locate** `oldFilename` inside the ID folder. Not found is an error.
3. **JDex detection**: if `oldFilename` matches the folder's JDex file (stem == folder name), this is a JDex rename.
   - Rename the folder to use the new name (keeping the JD prefix).
   - Rename the JDex file to match the new folder name.
   - Update the JDex frontmatter (aliases, heading) as `Rename` already does.
   - Set `IsJDex = true`, `FolderRenamed = true`.
4. **Regular file**: rename the file on disk (`oldFilename` -> `newFilename`).
5. **Update wikilinks** across the entire vault. The replacements map must include:
   - Stem without extension: `"old-stem" -> "new-stem"`
   - Full filename with extension: `"old-stem.md" -> "new-stem.md"`
   - For JDex files, also the folder-name form (as `Rename` does).
6. **Update H1 heading**: if the file's first `# ` heading is identical to the old stem, replace it with the new stem.

### CLI

```
mimic rename-file <id-ref> <old-filename> <new-filename>
```

Output:
```
Renamed file: "old-name.md" -> "new-name.md"
Path: /path/to/new/file.md
Updated N wiki links
Heading updated: yes/no
```

### MCP tool

```json
{
  "name": "rename_file",
  "parameters": {
    "ref": "string (required) — JD ID reference (e.g. S01.11.11)",
    "old_name": "string (required) — current filename",
    "new_name": "string (required) — desired filename"
  }
}
```

### Edge cases

- Empty ref -> error.
- Non-ID ref (scope, area, category) -> error.
- ID not found -> error.
- Empty old_name or new_name -> error.
- old_name == new_name -> no-op, return current path.
- old_name not found in ID folder -> error.
- new_name already exists in ID folder -> error.
- new_name missing `.md` extension -> auto-append `.md`.
- old_name missing `.md` extension -> auto-append `.md`.
- H1 heading differs from old stem -> heading left unchanged, `HeadingUpdated = false`.
- JDex rename: folder rename fails (e.g. target exists) -> error, no partial changes.

### Test plan

**Unit tests**:
- Empty ref -> error.
- Non-ID ref -> error.
- Empty old_name -> error.
- Empty new_name -> error.
- old_name == new_name -> no-op.

**Integration tests**:
- Rename regular file -> file renamed on disk, wikilinks updated (both stem and with `.md`).
- Rename JDex file -> folder renamed, JDex file renamed, frontmatter updated, wikilinks updated.
- old_name not found -> error, nothing changed.
- new_name already exists -> error, nothing changed.
- File with H1 matching old stem -> heading updated.
- File with H1 not matching old stem -> heading unchanged.
- Piped wikilinks `[[old|display]]` -> target updated, display preserved.
- Missing `.md` extension in args -> auto-appended, rename succeeds.

**Acceptance tests**:
- Rename a file, re-parse vault and read it back -> correct content at new path.
- Rename JDex file, browse vault -> new name reflected in structure.

## Consequences

- Extends ADR-0016's rename capability to individual files.
- Reuses `UpdateWikiLinks` from ADR-0017.
- JDex detection adds complexity but prevents folder/file name mismatches.
- H1 update is a convenience — keeps Obsidian's display name in sync.
