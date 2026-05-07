package main

import (
	"fmt"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newFrontmatterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "frontmatter <action> <ref> <file> <key> <value>",
		Short: "Edit YAML frontmatter fields",
		Long:  "Edit YAML frontmatter in a file within a JD ID folder.\nActions: set (scalar), add (append to list), remove (remove from list).",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			action, ref, file, key, value := args[0], args[1], args[2], args[3], args[4]
			var path string
			switch action {
			case "set":
				path, err = vault.SetFrontmatter(v, ref, file, key, value)
			case "add":
				path, err = vault.AddToFrontmatterList(v, ref, file, key, value)
			case "remove":
				path, err = vault.RemoveFromFrontmatterList(v, ref, file, key, value)
			default:
				return fmt.Errorf("unknown action %q (use set, add, or remove)", action)
			}
			if err != nil {
				return err
			}
			autoLog(cmd.ErrOrStderr(), v, scopeFromRef(ref), "frontmatter",
				fmt.Sprintf("%s/%s", ref, file), "",
				fmt.Sprintf("%s %s", action, key))
			fmt.Fprintf(cmd.OutOrStdout(), "Updated frontmatter in %s\n", path)
			return nil
		},
	}
	return cmd
}
