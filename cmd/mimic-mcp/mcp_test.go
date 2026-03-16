package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func testdataVault(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test file path")
	}
	return filepath.Join(filepath.Dir(filename), "..", "..", "testdata", "vault")
}

func callTool(t *testing.T, s interface{ HandleMessage(context.Context, json.RawMessage) mcp.JSONRPCMessage }, tool string, args map[string]any) *mcp.CallToolResult {
	t.Helper()

	params := map[string]any{
		"name":      tool,
		"arguments": args,
	}
	paramsJSON, _ := json.Marshal(params)
	reqJSON, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params":  json.RawMessage(paramsJSON),
	})

	resp := s.HandleMessage(context.Background(), json.RawMessage(reqJSON))

	// Extract result from response
	respJSON, _ := json.Marshal(resp)
	var parsed struct {
		Result *mcp.CallToolResult `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respJSON, &parsed); err != nil {
		t.Fatalf("unmarshalling response: %v\nraw: %s", err, respJSON)
	}
	if parsed.Error != nil {
		t.Fatalf("MCP error: %s", parsed.Error.Message)
	}
	if parsed.Result == nil {
		t.Fatalf("nil result, raw response: %s", respJSON)
	}
	return parsed.Result
}

func resultText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			return tc.Text
		}
	}
	raw, _ := json.Marshal(result.Content)
	t.Fatalf("no text content in result: %s", raw)
	return ""
}

func copyTestdataVault(t *testing.T) string {
	t.Helper()
	src := testdataVault(t)
	dst := t.TempDir()
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatalf("copying fixture: %v", err)
	}
	return dst
}

// --- Browse ---

func TestMCP_Browse(t *testing.T) {
	s := newServer(testdataVault(t))
	result := callTool(t, s, "browse", map[string]any{})
	text := resultText(t, result)
	if !strings.Contains(text, "S01 Me") {
		t.Errorf("expected S01 Me in output:\n%s", text)
	}
}

func TestMCP_BrowseFilter(t *testing.T) {
	s := newServer(testdataVault(t))
	result := callTool(t, s, "browse", map[string]any{"filter": "S01"})
	text := resultText(t, result)
	if !strings.Contains(text, "S01 Me") {
		t.Error("expected S01 in output")
	}
	if strings.Contains(text, "S02") {
		t.Error("S02 should not appear when filtering by S01")
	}
}

// --- Search ---

func TestMCP_SearchByName(t *testing.T) {
	s := newServer(testdataVault(t))
	result := callTool(t, s, "search", map[string]any{"query": "Entertainment"})
	text := resultText(t, result)
	if !strings.Contains(text, "Entertainment") {
		t.Errorf("expected Entertainment in output:\n%s", text)
	}
}

func TestMCP_SearchByRef(t *testing.T) {
	s := newServer(testdataVault(t))
	result := callTool(t, s, "search", map[string]any{"query": "S01.11.11"})
	text := resultText(t, result)
	if !strings.Contains(text, "Theatre, 2025 Season") {
		t.Errorf("expected Theatre in output:\n%s", text)
	}
}

func TestMCP_SearchByContent(t *testing.T) {
	s := newServer(testdataVault(t))
	result := callTool(t, s, "search", map[string]any{"query": "Italian", "content": true})
	text := resultText(t, result)
	if !strings.Contains(text, "Italian") {
		t.Errorf("expected content match in output:\n%s", text)
	}
}

func TestMCP_SearchMissingQuery(t *testing.T) {
	s := newServer(testdataVault(t))
	result := callTool(t, s, "search", map[string]any{})
	if !result.IsError {
		t.Fatal("expected error for missing query")
	}
}

// --- Read ---

func TestMCP_Read(t *testing.T) {
	s := newServer(testdataVault(t))
	result := callTool(t, s, "read", map[string]any{"ref": "S01.11.11"})
	text := resultText(t, result)
	if !strings.Contains(text, "Theatre, 2025 Season") {
		t.Errorf("expected JDex content in output:\n%s", text)
	}
}

func TestMCP_ReadScope(t *testing.T) {
	s := newServer(testdataVault(t))
	result := callTool(t, s, "read", map[string]any{"ref": "S01"})
	text := resultText(t, result)
	if !strings.Contains(text, "Lifestyle") {
		t.Errorf("expected area in scope read:\n%s", text)
	}
}

func TestMCP_ReadFile(t *testing.T) {
	s := newServer(testdataVault(t))
	result := callTool(t, s, "read", map[string]any{"ref": "S01.11.11", "file": "notes.md"})
	text := resultText(t, result)
	if !strings.Contains(text, "theatre season") {
		t.Errorf("expected file content:\n%s", text)
	}
}

func TestMCP_ReadInvalidRef(t *testing.T) {
	s := newServer(testdataVault(t))
	result := callTool(t, s, "read", map[string]any{"ref": "xyz"})
	if !result.IsError {
		t.Fatal("expected error for invalid ref")
	}
}

// --- Create ---

func TestMCP_Create(t *testing.T) {
	root := copyTestdataVault(t)
	s := newServer(root)
	result := callTool(t, s, "create", map[string]any{"category": "S01.12", "name": "Pasta"})
	text := resultText(t, result)
	if !strings.Contains(text, "S01.12.11") {
		t.Errorf("expected new ref in output:\n%s", text)
	}
}

// --- Archive ---

func TestMCP_Archive(t *testing.T) {
	root := copyTestdataVault(t)
	s := newServer(root)
	result := callTool(t, s, "archive", map[string]any{"ref": "S01.11.11"})
	text := resultText(t, result)
	if !strings.Contains(text, "Archived") {
		t.Errorf("expected archived confirmation in output:\n%s", text)
	}
}

// --- Inbox ---

func TestMCP_Inbox(t *testing.T) {
	s := newServer(testdataVault(t))
	result := callTool(t, s, "inbox", map[string]any{})
	text := resultText(t, result)
	if !strings.Contains(text, "new-show-idea.md") {
		t.Errorf("expected inbox item in output:\n%s", text)
	}
}

func TestMCP_InboxScopeFilter(t *testing.T) {
	s := newServer(testdataVault(t))
	result := callTool(t, s, "inbox", map[string]any{"scope": "S01"})
	text := resultText(t, result)
	if strings.Contains(text, "episode-pitch.md") {
		t.Error("S02 item should not appear when filtering by S01")
	}
}
