package vault

import (
	"strings"
	"testing"
)

func TestLog_RejectsInvalidScope(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	cases := []string{"", "foo", "S0", "S001", "S01.10"}
	for _, c := range cases {
		if err := Log(v, c, "create", "X", "", ""); err == nil {
			t.Errorf("expected error for scope %q, got nil", c)
		}
	}
}

func TestLog_RejectsUnknownScope(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	if err := Log(v, "S99", "create", "X", "", ""); err == nil {
		t.Error("expected error for unknown scope S99")
	}
}

func TestLog_RejectsNonCanonicalOp(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	cases := []string{"", "foo", "Create", "DELETE"}
	for _, c := range cases {
		if err := Log(v, "S01", c, "X", "", ""); err == nil {
			t.Errorf("expected error for op %q", c)
		}
	}
}

func TestLog_AcceptsAllCanonicalOps(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	ops := []string{"create", "archive", "move", "move-file", "rename", "rename-file", "frontmatter", "ingest", "lint"}
	for _, op := range ops {
		if err := Log(v, "S01", op, "target", "", ""); err != nil {
			t.Errorf("op %q rejected: %v", op, err)
		}
	}
}

func TestLog_RejectsEmptyTarget(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	if err := Log(v, "S01", "create", "", "", ""); err == nil {
		t.Error("expected error for empty target")
	}
}

func TestLogTail_EmptyWhenLogMissing(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	entries, err := LogTail(v, "S01", 5)
	if err != nil {
		t.Fatalf("LogTail: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestLogTail_NWithFewerEntries(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	if err := Log(v, "S01", "create", "only", "", ""); err != nil {
		t.Fatalf("Log: %v", err)
	}
	entries, err := LogTail(v, "S01", 10)
	if err != nil {
		t.Fatalf("LogTail: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if !strings.Contains(entries[0], "create | only") {
		t.Errorf("entry content unexpected: %s", entries[0])
	}
}

func TestLogTail_ZeroOrNegativeN(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	if err := Log(v, "S01", "create", "x", "", ""); err != nil {
		t.Fatalf("Log: %v", err)
	}
	for _, n := range []int{0, -1, -100} {
		entries, err := LogTail(v, "S01", n)
		if err != nil {
			t.Fatalf("LogTail(n=%d): %v", n, err)
		}
		if len(entries) != 0 {
			t.Errorf("expected 0 entries for n=%d, got %d", n, len(entries))
		}
	}
}

func TestLog_HeaderWithoutSecondary(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	if err := Log(v, "S01", "create", "S01.11.99 Foo", "", ""); err != nil {
		t.Fatalf("Log: %v", err)
	}
	entries, err := LogTail(v, "S01", 1)
	if err != nil {
		t.Fatalf("LogTail: %v", err)
	}
	if strings.Contains(entries[0], "→") {
		t.Errorf("header should not contain arrow when secondary empty: %s", entries[0])
	}
}

func TestLog_DetailsAppendedAsBody(t *testing.T) {
	root := copyFixtureVault(t)
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault: %v", err)
	}
	if err := Log(v, "S01", "ingest", "Article", "", "Touched: A, B, C."); err != nil {
		t.Fatalf("Log: %v", err)
	}
	entries, err := LogTail(v, "S01", 1)
	if err != nil {
		t.Fatalf("LogTail: %v", err)
	}
	if !strings.Contains(entries[0], "Touched: A, B, C.") {
		t.Errorf("details not appended: %s", entries[0])
	}
}
