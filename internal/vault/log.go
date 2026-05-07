package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const logSlot = 7

var canonicalLogOps = map[string]bool{
	"create":      true,
	"archive":     true,
	"move":        true,
	"move-file":   true,
	"rename":      true,
	"rename-file": true,
	"frontmatter": true,
	"ingest":      true,
	"lint":        true,
}

// nowFunc is overridable in tests if needed; defaults to time.Now.
var nowFunc = time.Now

// Log appends an entry to the scope's activity log. It lazily creates the
// Management area, the .07 Log ID folder, the JDex file, and log.md if any
// of them are missing.
//
// op must be one of the canonical operations (see canonicalLogOps).
// target must be non-empty. secondary and details may be empty.
//
// On success, the in-memory vault tree is also updated so the new Management
// area / Log ID becomes discoverable without a full re-parse. Callers that
// need full structural awareness can still re-parse if desired.
func Log(v *Vault, scope, op, target, secondary, details string) error {
	m := searchScopeRe.FindStringSubmatch(scope)
	if m == nil {
		return fmt.Errorf("invalid scope reference: %q", scope)
	}
	scopeNum, _ := strconv.Atoi(m[1])

	if !canonicalLogOps[op] {
		return fmt.Errorf("op %q is not in the canonical set", op)
	}
	if strings.TrimSpace(target) == "" {
		return fmt.Errorf("target cannot be empty")
	}

	s, err := findScope(v, scopeNum)
	if err != nil {
		return err
	}

	logFolder, err := ensureLogFolder(v, s, scopeNum)
	if err != nil {
		return err
	}
	logFile := filepath.Join(logFolder, "log.md")

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		header := fmt.Sprintf("# Activity Log — S%02d\n", scopeNum)
		if err := os.WriteFile(logFile, []byte(header), 0o644); err != nil {
			return fmt.Errorf("creating log file: %w", err)
		}
	}

	entry := buildLogEntry(nowFunc(), op, target, secondary, details)

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("opening log: %w", err)
	}
	defer f.Close()
	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("appending log: %w", err)
	}
	return nil
}

// LogTail returns the last n entries from the scope's log file. Each entry is
// a markdown block (H2 header plus optional details body) without the trailing
// blank line. Returns an empty slice if the log file does not exist.
func LogTail(v *Vault, scope string, n int) ([]string, error) {
	if n <= 0 {
		return nil, nil
	}
	m := searchScopeRe.FindStringSubmatch(scope)
	if m == nil {
		return nil, fmt.Errorf("invalid scope reference: %q", scope)
	}
	scopeNum, _ := strconv.Atoi(m[1])

	s, err := findScope(v, scopeNum)
	if err != nil {
		return nil, err
	}

	logFile := filepath.Join(s.Path, scopeMgmtAreaFolder(scopeNum), scopeLogIDFolder(scopeNum), "log.md")
	data, err := os.ReadFile(logFile)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading log: %w", err)
	}

	entries := parseLogEntries(string(data))
	if n >= len(entries) {
		return entries, nil
	}
	return entries[len(entries)-n:], nil
}

// buildLogEntry formats a single log entry. The returned string starts with a
// blank line so that, when appended to the existing file, the H2 header is
// separated from any preceding content.
func buildLogEntry(now time.Time, op, target, secondary, details string) string {
	var b strings.Builder
	b.WriteString("\n## [")
	b.WriteString(now.Format("2006-01-02 15:04"))
	b.WriteString("] ")
	b.WriteString(op)
	b.WriteString(" | ")
	b.WriteString(target)
	if secondary != "" {
		b.WriteString(" → ")
		b.WriteString(secondary)
	}
	b.WriteString("\n")
	if details != "" {
		b.WriteString("\n")
		b.WriteString(details)
		if !strings.HasSuffix(details, "\n") {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// parseLogEntries splits the log body into entries. Each entry begins at a
// line starting with "## [" and runs until the next such line (or EOF).
// Trailing whitespace on each entry is trimmed.
func parseLogEntries(body string) []string {
	lines := strings.Split(body, "\n")
	var entries []string
	var current []string
	flush := func() {
		if len(current) == 0 {
			return
		}
		entry := strings.TrimRight(strings.Join(current, "\n"), "\n ")
		if entry != "" {
			entries = append(entries, entry)
		}
		current = nil
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "## [") {
			flush()
			current = []string{line}
			continue
		}
		if len(current) > 0 {
			current = append(current, line)
		}
	}
	flush()
	return entries
}

// scopeMgmtAreaFolder returns the folder name for a scope's Management area,
// e.g. "S01.00-09 Management for S01".
func scopeMgmtAreaFolder(scopeNum int) string {
	return fmt.Sprintf("S%02d.00-09 Management for S%02d", scopeNum, scopeNum)
}

// scopeLogIDFolder returns the folder name for a scope's .07 Log ID, e.g.
// "S01.07 Log for S01.00-09".
func scopeLogIDFolder(scopeNum int) string {
	return fmt.Sprintf("S%02d.%02d Log for S%02d.00-09", scopeNum, logSlot, scopeNum)
}

// ensureLogFolder creates the Management area, the .07 Log ID folder, and
// the JDex file if they are missing. Returns the absolute path to the .07
// Log ID folder.
func ensureLogFolder(v *Vault, s *Scope, scopeNum int) (string, error) {
	mgmtAreaName := scopeMgmtAreaFolder(scopeNum)
	mgmtAreaPath := filepath.Join(s.Path, mgmtAreaName)

	if _, err := os.Stat(mgmtAreaPath); os.IsNotExist(err) {
		if err := os.MkdirAll(mgmtAreaPath, 0o755); err != nil {
			return "", fmt.Errorf("creating management area: %w", err)
		}
		// Reflect in in-memory tree so subsequent calls see it.
		s.Areas = append(s.Areas, Area{
			ScopeNumber: scopeNum,
			RangeStart:  0,
			RangeEnd:    9,
			Name:        fmt.Sprintf("Management for S%02d", scopeNum),
			Path:        mgmtAreaPath,
		})
	}

	logIDName := scopeLogIDFolder(scopeNum)
	logIDPath := filepath.Join(mgmtAreaPath, logIDName)
	if err := os.MkdirAll(logIDPath, 0o755); err != nil {
		return "", fmt.Errorf("creating log ID folder: %w", err)
	}

	jdexPath := filepath.Join(logIDPath, logIDName+".md")
	if _, err := os.Stat(jdexPath); os.IsNotExist(err) {
		jdex := fmt.Sprintf(`---
aliases:
  - %s
location: Obsidian
tags:
  - jdex
  - index
---
# %s

Append-only chronological record of vault mutations and skill activity for scope S%02d. Maintained by `+"`mimic`"+`. See `+"`log.md`"+` in this folder for entries.
`, logIDName, logIDName, scopeNum)
		if err := os.WriteFile(jdexPath, []byte(jdex), 0o644); err != nil {
			return "", fmt.Errorf("writing JDex: %w", err)
		}
	}

	return logIDPath, nil
}
