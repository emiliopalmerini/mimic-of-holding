package main

import (
	"fmt"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <ref> <new-name>",
		Short: "Rename a JD item",
		Long:  "Change the human-readable name of any JD item (scope, area, category, ID).",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			result, err := vault.Rename(v, args[0], args[1])
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "Renamed %s: %q → %q\n", result.Ref, result.OldName, result.NewName)
			fmt.Fprintf(w, "Path: %s\n", result.NewPath)
			if result.LinksUpdated > 0 {
				fmt.Fprintf(w, "Updated %d wiki links\n", result.LinksUpdated)
			}
			return nil
		},
	}
}
