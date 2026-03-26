package main

import (
	"context"
	"fmt"
	"strconv"
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
			mcp.WithDescription("Search the vault by JD reference (S01.11), name (Entertainment), file content, backlinks, or tags."),
			mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
			mcp.WithBoolean("content", mcp.Description("If true, search inside file content instead of names")),
			mcp.WithBoolean("meta", mcp.Description("If true, query is key:value format for YAML frontmatter search")),
			mcp.WithBoolean("backlinks", mcp.Description("If true, query is a JD ID ref; returns notes that link to it")),
			mcp.WithBoolean("tags", mcp.Description("If true, list all tags (empty/whitespace query) or find notes by tag (query = tag name)")),
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
			mcp.WithDescription("Create a new JD ID in the given category. Use 'templates' tool first to discover available templates."),
			mcp.WithString("category", mcp.Required(), mcp.Description("Category reference (e.g., S01.11)")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Name for the new ID")),
			mcp.WithString("template", mcp.Description("Optional template name for the JDex file")),
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
			mcp.WithDescription("Create or overwrite a file inside a JD ID folder. Use 'templates' tool first to discover available templates."),
			mcp.WithString("ref", mcp.Required(), mcp.Description("JD ID reference (e.g., S01.11.11)")),
			mcp.WithString("file", mcp.Required(), mcp.Description("Filename to write")),
			mcp.WithString("content", mcp.Description("File content (optional when template is provided)")),
			mcp.WithString("template", mcp.Description("Optional template name (used when content is empty)")),
		),
		writeHandler(vaultRoot),
	)

	s.AddTool(
		mcp.NewTool("edit",
			mcp.WithDescription("Search-and-replace edit of a file inside a JD ID folder. old_string must appear exactly once."),
			mcp.WithString("ref", mcp.Required(), mcp.Description("JD ID reference (e.g., S01.11.11)")),
			mcp.WithString("file", mcp.Required(), mcp.Description("Filename to edit")),
			mcp.WithString("old_string", mcp.Required(), mcp.Description("Exact text to find (must be unique in file)")),
			mcp.WithString("new_string", mcp.Required(), mcp.Description("Replacement text")),
		),
		editHandler(vaultRoot),
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
		mcp.NewTool("rename",
			mcp.WithDescription("Rename a JD item (scope, area, category, or ID). Updates wiki links across the vault."),
			mcp.WithString("ref", mcp.Required(), mcp.Description("JD reference to rename")),
			mcp.WithString("name", mcp.Required(), mcp.Description("New human-readable name")),
		),
		renameHandler(vaultRoot),
	)

	s.AddTool(
		mcp.NewTool("move",
			mcp.WithDescription("Move a JD item to a different parent. Move an ID to a category, or a category to an area. Updates wiki links."),
			mcp.WithString("ref", mcp.Required(), mcp.Description("JD reference to move")),
			mcp.WithString("to", mcp.Required(), mcp.Description("Target parent reference")),
		),
		moveHandler(vaultRoot),
	)

	s.AddTool(
		mcp.NewTool("move_file",
			mcp.WithDescription("Move a file from one JD ID to another."),
			mcp.WithString("from", mcp.Required(), mcp.Description("Source JD ID reference")),
			mcp.WithString("file", mcp.Required(), mcp.Description("Filename to move")),
			mcp.WithString("to", mcp.Required(), mcp.Description("Target JD ID reference")),
		),
		moveFileHandler(vaultRoot),
	)

	s.AddTool(
		mcp.NewTool("rename_file",
			mcp.WithDescription("Rename a file inside a JD ID folder. Updates wiki links across the vault. If the file is the JDex file, the folder is also renamed."),
			mcp.WithString("ref", mcp.Required(), mcp.Description("JD ID reference (e.g., S01.11.11)")),
			mcp.WithString("old_name", mcp.Required(), mcp.Description("Current filename")),
			mcp.WithString("new_name", mcp.Required(), mcp.Description("Desired new filename")),
		),
		renameFileHandler(vaultRoot),
	)

	s.AddTool(
		mcp.NewTool("templates",
			mcp.WithDescription("List available templates for a category. Templates are discovered from .03 IDs in the category, area, and scope hierarchy."),
			mcp.WithString("category", mcp.Required(), mcp.Description("Category reference (e.g., S01.11)")),
		),
		templatesHandler(vaultRoot),
	)

	s.AddTool(
		mcp.NewTool("frontmatter",
			mcp.WithDescription("Edit YAML frontmatter fields. Actions: 'set' (scalar), 'add' (append to list), 'remove' (remove from list)."),
			mcp.WithString("ref", mcp.Required(), mcp.Description("JD ID reference (e.g., S01.11.11)")),
			mcp.WithString("file", mcp.Required(), mcp.Description("Filename within the ID folder")),
			mcp.WithString("action", mcp.Required(), mcp.Description("Action: set, add, or remove")),
			mcp.WithString("key", mcp.Required(), mcp.Description("Frontmatter field name")),
			mcp.WithString("value", mcp.Required(), mcp.Description("Value to set, add, or remove")),
		),
		frontmatterHandler(vaultRoot),
	)

	s.AddTool(
		mcp.NewTool("stats",
			mcp.WithDescription("Show vault statistics: totals, empty categories, orphan notes (no inbound links), and largest categories."),
		),
		statsHandler(vaultRoot),
	)

	s.AddTool(
		mcp.NewTool("recent",
			mcp.WithDescription("List the most recently modified files in the vault."),
			mcp.WithNumber("limit", mcp.Description("Max results to return (default 10)")),
			mcp.WithString("scope", mcp.Description("Optional scope filter (e.g., S01)")),
		),
		recentHandler(vaultRoot),
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
			Content:   request.GetBool("content", false),
			Meta:      request.GetBool("meta", false),
			Backlinks: request.GetBool("backlinks", false),
			Tags:      request.GetBool("tags", false),
			Scope:     request.GetString("scope", ""),
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
		template := request.GetString("template", "")
		result, err := vault.Create(v, category, name, template)
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
		template := request.GetString("template", "")
		path, err := vault.WriteFile(v, ref, file, content, template)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Written %s", path)), nil
	}
}

func editHandler(vaultRoot string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		v, err := parseVaultForMCP(vaultRoot)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		ref := request.GetString("ref", "")
		file := request.GetString("file", "")
		oldString := request.GetString("old_string", "")
		newString := request.GetString("new_string", "")
		if ref == "" || file == "" || oldString == "" {
			return mcp.NewToolResultError("ref, file, and old_string are required"), nil
		}
		path, err := vault.EditFile(v, ref, file, oldString, newString)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Edited %s", path)), nil
	}
}

func renameHandler(vaultRoot string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		v, err := parseVaultForMCP(vaultRoot)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		ref := request.GetString("ref", "")
		name := request.GetString("name", "")
		if ref == "" || name == "" {
			return mcp.NewToolResultError("ref and name are required"), nil
		}
		result, err := vault.Rename(v, ref, name)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		text := fmt.Sprintf("Renamed %s: %q → %q\nPath: %s", result.Ref, result.OldName, result.NewName, result.NewPath)
		if result.LinksUpdated > 0 {
			text += fmt.Sprintf("\nUpdated %d wiki links", result.LinksUpdated)
		}
		return mcp.NewToolResultText(text), nil
	}
}

func renameFileHandler(vaultRoot string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		v, err := parseVaultForMCP(vaultRoot)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		ref := request.GetString("ref", "")
		oldName := request.GetString("old_name", "")
		newName := request.GetString("new_name", "")
		if ref == "" || oldName == "" || newName == "" {
			return mcp.NewToolResultError("ref, old_name, and new_name are required"), nil
		}
		result, err := vault.RenameFile(v, ref, oldName, newName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		text := fmt.Sprintf("Renamed file: %q → %q\nPath: %s", oldName, newName, result.NewPath)
		if result.LinksUpdated > 0 {
			text += fmt.Sprintf("\nUpdated %d wiki links", result.LinksUpdated)
		}
		if result.HeadingUpdated {
			text += "\nHeading updated: yes"
		}
		if result.FolderRenamed {
			text += "\nFolder renamed: yes"
		}
		return mcp.NewToolResultText(text), nil
	}
}

func templatesHandler(vaultRoot string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		v, err := parseVaultForMCP(vaultRoot)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		category := request.GetString("category", "")
		if category == "" {
			return mcp.NewToolResultError("category is required"), nil
		}
		templates, err := vault.ListTemplates(v, category)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if len(templates) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No templates found for %s", category)), nil
		}
		var lines []string
		for _, t := range templates {
			lines = append(lines, fmt.Sprintf("%s (%s, %s)", t.Name, t.SourceRef, t.Source))
		}
		return mcp.NewToolResultText(fmt.Sprintf("Templates for %s:\n%s", category, strings.Join(lines, "\n"))), nil
	}
}

func moveHandler(vaultRoot string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		v, err := parseVaultForMCP(vaultRoot)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		ref := request.GetString("ref", "")
		to := request.GetString("to", "")
		if ref == "" || to == "" {
			return mcp.NewToolResultError("ref and to are required"), nil
		}
		result, err := vault.Move(v, ref, to)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		text := fmt.Sprintf("Moved %s → %s\nPath: %s", result.OldRef, result.NewRef, result.NewPath)
		if result.LinksUpdated > 0 {
			text += fmt.Sprintf("\nUpdated %d wiki links", result.LinksUpdated)
		}
		return mcp.NewToolResultText(text), nil
	}
}

func moveFileHandler(vaultRoot string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		v, err := parseVaultForMCP(vaultRoot)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		from := request.GetString("from", "")
		file := request.GetString("file", "")
		to := request.GetString("to", "")
		if from == "" || file == "" || to == "" {
			return mcp.NewToolResultError("from, file, and to are required"), nil
		}
		path, err := vault.MoveFile(v, from, file, to)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Moved to %s", path)), nil
	}
}

func frontmatterHandler(vaultRoot string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		v, err := parseVaultForMCP(vaultRoot)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		ref := request.GetString("ref", "")
		file := request.GetString("file", "")
		action := request.GetString("action", "")
		key := request.GetString("key", "")
		value := request.GetString("value", "")
		if ref == "" || file == "" || action == "" || key == "" {
			return mcp.NewToolResultError("ref, file, action, and key are required"), nil
		}

		var path string
		switch action {
		case "set":
			path, err = vault.SetFrontmatter(v, ref, file, key, value)
		case "add":
			path, err = vault.AddToFrontmatterList(v, ref, file, key, value)
		case "remove":
			path, err = vault.RemoveFromFrontmatterList(v, ref, file, key, value)
		default:
			return mcp.NewToolResultError(fmt.Sprintf("unknown action %q (use set, add, or remove)", action)), nil
		}
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Updated frontmatter in %s", path)), nil
	}
}

func statsHandler(vaultRoot string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		v, err := parseVaultForMCP(vaultRoot)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		stats, err := vault.Stats(v)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		var b strings.Builder
		fmt.Fprintf(&b, "Scopes: %d  Areas: %d  Categories: %d  IDs: %d  Files: %d\n",
			stats.TotalScopes, stats.TotalAreas, stats.TotalCategories, stats.TotalIDs, stats.TotalFiles)
		if len(stats.EmptyCategories) > 0 {
			fmt.Fprintf(&b, "\nEmpty categories:\n")
			for _, ref := range stats.EmptyCategories {
				fmt.Fprintf(&b, "  %s\n", ref)
			}
		}
		if len(stats.OrphanIDs) > 0 {
			fmt.Fprintf(&b, "\nOrphan IDs (no inbound links):\n")
			for _, ref := range stats.OrphanIDs {
				fmt.Fprintf(&b, "  %s\n", ref)
			}
		}
		if len(stats.LargestCategories) > 0 {
			fmt.Fprintf(&b, "\nLargest categories:\n")
			for _, cs := range stats.LargestCategories {
				fmt.Fprintf(&b, "  %s %s (%d IDs)\n", cs.Ref, cs.Name, cs.Count)
			}
		}
		return mcp.NewToolResultText(b.String()), nil
	}
}

func recentHandler(vaultRoot string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		v, err := parseVaultForMCP(vaultRoot)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		limitStr := request.GetString("limit", "")
		limit := 10
		if limitStr != "" {
			if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
				limit = n
			}
		}
		scope := request.GetString("scope", "")
		results, err := vault.Recent(v, limit, scope)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if len(results) == 0 {
			return mcp.NewToolResultText("No recent files found."), nil
		}
		var b strings.Builder
		for _, r := range results {
			fmt.Fprintf(&b, "[%s] %s  %s\n", r.Ref, r.File, r.ModTime.Format("2006-01-02 15:04"))
			fmt.Fprintf(&b, "  %s\n", r.Breadcrumb)
		}
		return mcp.NewToolResultText(b.String()), nil
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
			fmt.Fprintf(&b, "  %s", item.File)
			if item.Preview != "" {
				fmt.Fprintf(&b, "  — %s", item.Preview)
			}
			fmt.Fprintln(&b)
		}
		return mcp.NewToolResultText(b.String()), nil
	}
}
