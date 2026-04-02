# ADR-0030: Binary file detection in read

## Status

Proposed

## Context

`Read` uses `os.ReadFile` + `string(data)` for any file, including images, PDFs, and other binaries. This produces garbage output that wastes LLM tokens and confuses the model. Additionally, the file listing (`Files []string`) provides no metadata — no sizes, no indication of which files are binary.

## Decision

Detect binary files and return a placeholder instead of raw content. Enrich the file listing with metadata.

### Changes

1. **`isBinaryFile(path) bool` helper**: Read the first 512 bytes and check for null bytes (same heuristic as `net/http.DetectContentType`).

2. **`FileInfo` struct** replaces `Files []string`:

```go
type FileInfo struct {
    Name     string
    Size     int64
    IsBinary bool
}
```

3. **`ReadResult` additions**:
   - `Files` changes from `[]string` to `[]FileInfo`.
   - New `IsBinary bool` field — set when reading a binary file directly.

4. **Behavior when reading a binary file** (`file` param points to binary):
   - Return `IsBinary: true`, `Content: ""`.
   - Type remains `"file"`.

5. **File listing** in ID reads:
   - Each entry includes name, size in bytes, and binary flag.
   - Sorted alphabetically (unchanged).

### Edge cases

- Empty file (0 bytes) → not binary.
- File with only null bytes → binary.
- File that starts with text but has embedded nulls → binary (conservative).

### MCP rendering

- Binary file read: `[Binary file: photo.jpg (450.0 KB)]`
- File listing entry: `photo.jpg (binary, 450.0 KB)` or `notes.md (234 B)`

### CLI rendering

Same format as MCP.

### Test plan

**Testdata**: Add a small binary file (a few bytes with null characters) to `testdata/vault/S01 Me/S01.10-19 Lifestyle/S01.11 Entertainment/S01.11.11 Theatre, 2025 Season/`.

**Unit tests**:
- `isBinaryFile` correctly identifies binary vs. text files.
- `isBinaryFile` returns false for empty files.

**Integration tests**:
- `Read(v, "S01.11.11", "binary-file.bin")` returns `IsBinary: true`, empty `Content`.
- `Read(v, "S01.11.11", "")` file listing includes `FileInfo` with correct name, non-zero size, and binary flag.
- `Read(v, "S01.11.11", "notes.md")` returns `IsBinary: false`.

**Acceptance tests**:
- File listing for an ID contains both text and binary files with correct metadata.
- Binary file read returns empty content with IsBinary flag.

## Consequences

- `Files` type change from `[]string` to `[]FileInfo` is a breaking change to the Go struct — CLI and MCP adapters must be updated simultaneously.
- MCP output is still text, so MCP clients see improved text output, not a breaking API change.
- Prevents wasted tokens from binary garbage in LLM conversations.
