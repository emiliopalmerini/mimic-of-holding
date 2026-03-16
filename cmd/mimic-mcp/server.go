package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func newServer(vaultRoot string) *server.MCPServer {
	s := server.NewMCPServer(
		"mimic-of-holding",
		"0.1.0",
	)

	s.AddTool(
		mcp.NewTool("browse",
			mcp.WithDescription("Display the vault tree. Optional filter: scope (S01), area (S01.10-19), or category (S01.11)."),
			mcp.WithString("filter", mcp.Description("Optional filter")),
		),
		browseHandler(vaultRoot),
	)

	s.AddTool(
		mcp.NewTool("search",
			mcp.WithDescription("Search the vault by JD reference (S01.11), name (Entertainment), or file content."),
			mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
			mcp.WithBoolean("content", mcp.Description("If true, search inside file content instead of names")),
			mcp.WithString("scope", mcp.Description("Optional scope filter (e.g., S01)")),
		),
		searchHandler(vaultRoot),
	)

	s.AddTool(
		mcp.NewTool("read",
			mcp.WithDescription("Read any JD level (scope, area, category, ID) or a specific file within an ID."),
			mcp.WithString("ref", mcp.Required(), mcp.Description("JD reference (S01, S01.10-19, S01.11, S01.11.11)")),
			mcp.WithString("file", mcp.Description("Optional filename to read within an ID")),
			mcp.WithBoolean("deep", mcp.Description("If true, recursively include all descendant content")),
		),
		readHandler(vaultRoot),
	)

	s.AddTool(
		mcp.NewTool("create",
			mcp.WithDescription("Create a new JD ID in the given category."),
			mcp.WithString("category", mcp.Required(), mcp.Description("Category reference (e.g., S01.11)")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Name for the new ID")),
		),
		createHandler(vaultRoot),
	)

	s.AddTool(
		mcp.NewTool("archive",
			mcp.WithDescription("Archive a JD ID or category to its parent's archive folder."),
			mcp.WithString("ref", mcp.Required(), mcp.Description("JD reference to archive")),
		),
		archiveHandler(vaultRoot),
	)

	s.AddTool(
		mcp.NewTool("write",
			mcp.WithDescription("Create or overwrite a file inside a JD ID folder."),
			mcp.WithString("ref", mcp.Required(), mcp.Description("JD ID reference (e.g., S01.11.11)")),
			mcp.WithString("file", mcp.Required(), mcp.Description("Filename to write")),
			mcp.WithString("content", mcp.Required(), mcp.Description("File content")),
		),
		writeHandler(vaultRoot),
	)

	s.AddTool(
		mcp.NewTool("append",
			mcp.WithDescription("Append content to a file inside a JD ID folder. Creates the file if it doesn't exist."),
			mcp.WithString("ref", mcp.Required(), mcp.Description("JD ID reference (e.g., S01.11.11)")),
			mcp.WithString("file", mcp.Required(), mcp.Description("Filename to append to")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Content to append")),
		),
		appendHandler(vaultRoot),
	)

	s.AddTool(
		mcp.NewTool("inbox",
			mcp.WithDescription("List files in inbox folders across the vault."),
			mcp.WithString("scope", mcp.Description("Optional scope filter (e.g., S01)")),
		),
		inboxHandler(vaultRoot),
	)

	return s
}

func parseVaultForMCP(root string) (*vault.Vault, error) {
	return vault.ParseVault(root)
}

func browseHandler(vaultRoot string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		v, err := parseVaultForMCP(vaultRoot)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		filter := request.GetString("filter", "")
		out, err := vault.Browse(v, filter)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(out), nil
	}
}

func searchHandler(vaultRoot string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		v, err := parseVaultForMCP(vaultRoot)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		query := request.GetString("query", "")
		if query == "" {
			return mcp.NewToolResultError("query is required"), nil
		}
		opts := vault.SearchOpts{
			Content: request.GetBool("content", false),
			Scope:   request.GetString("scope", ""),
		}
		results, err := vault.Search(v, query, opts)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if len(results) == 0 {
			return mcp.NewToolResultText("No results found."), nil
		}
		var b strings.Builder
		for _, r := range results {
			fmt.Fprintf(&b, "[%s] %s  %s\n", r.Type, r.Ref, r.Name)
			if r.Breadcrumb != "" {
				fmt.Fprintf(&b, "  %s\n", r.Breadcrumb)
			}
			if r.MatchLine != "" {
				fmt.Fprintf(&b, "  > %s\n", r.MatchLine)
			}
		}
		return mcp.NewToolResultText(b.String()), nil
	}
}

func readHandler(vaultRoot string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		v, err := parseVaultForMCP(vaultRoot)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		ref := request.GetString("ref", "")
		if ref == "" {
			return mcp.NewToolResultError("ref is required"), nil
		}
		file := request.GetString("file", "")
		deep := request.GetBool("deep", false)
		var result *vault.ReadResult
		if deep {
			result, err = vault.ReadDeep(v, ref, file)
		} else {
			result, err = vault.Read(v, ref, file)
		}
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		var b strings.Builder
		renderReadResult(&b, result, 0)
		return mcp.NewToolResultText(b.String()), nil
	}
}

func createHandler(vaultRoot string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		v, err := parseVaultForMCP(vaultRoot)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		category := request.GetString("category", "")
		name := request.GetString("name", "")
		if category == "" || name == "" {
			return mcp.NewToolResultError("category and name are required"), nil
		}
		result, err := vault.Create(v, category, name)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Created %s %s\nPath: %s", result.Ref, result.Name, result.Path)), nil
	}
}

func archiveHandler(vaultRoot string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		v, err := parseVaultForMCP(vaultRoot)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		ref := request.GetString("ref", "")
		if ref == "" {
			return mcp.NewToolResultError("ref is required"), nil
		}
		result, err := vault.Archive(v, ref)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Archived %s\nNew path: %s", result.Ref, result.NewPath)), nil
	}
}

func renderReadResult(b *strings.Builder, result *vault.ReadResult, indent int) {
	prefix := strings.Repeat("  ", indent)
	fmt.Fprintf(b, "%s# %s %s\n", prefix, result.Ref, result.Name)
	if result.Content != "" {
		fmt.Fprintf(b, "%s%s\n", prefix, result.Content)
	}
	if len(result.Children) > 0 && len(result.DeepChildren) == 0 {
		for _, c := range result.Children {
			fmt.Fprintf(b, "%s  %s\n", prefix, c)
		}
	}
	if len(result.Files) > 0 {
		fmt.Fprintf(b, "%sFiles: %s\n", prefix, strings.Join(result.Files, ", "))
	}
	for _, child := range result.DeepChildren {
		renderReadResult(b, &child, indent+1)
	}
}

func appendHandler(vaultRoot string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		v, err := parseVaultForMCP(vaultRoot)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		ref := request.GetString("ref", "")
		file := request.GetString("file", "")
		content := request.GetString("content", "")
		if ref == "" || file == "" {
			return mcp.NewToolResultError("ref and file are required"), nil
		}
		path, err := vault.AppendFile(v, ref, file, content)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Appended to %s", path)), nil
	}
}

func writeHandler(vaultRoot string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		v, err := parseVaultForMCP(vaultRoot)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		ref := request.GetString("ref", "")
		file := request.GetString("file", "")
		content := request.GetString("content", "")
		if ref == "" || file == "" {
			return mcp.NewToolResultError("ref and file are required"), nil
		}
		path, err := vault.WriteFile(v, ref, file, content)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Written %s", path)), nil
	}
}

func inboxHandler(vaultRoot string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		v, err := parseVaultForMCP(vaultRoot)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		scope := request.GetString("scope", "")
		items, err := vault.Inbox(v, scope)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if len(items) == 0 {
			return mcp.NewToolResultText("All inboxes are empty."), nil
		}
		var b strings.Builder
		currentRef := ""
		for _, item := range items {
			if item.InboxRef != currentRef {
				if currentRef != "" {
					fmt.Fprintln(&b)
				}
				fmt.Fprintf(&b, "%s (%s)\n", item.InboxRef, item.InboxName)
				currentRef = item.InboxRef
			}
			fmt.Fprintf(&b, "  %s\n", item.File)
		}
		return mcp.NewToolResultText(b.String()), nil
	}
}
