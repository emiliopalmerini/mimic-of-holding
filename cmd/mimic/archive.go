package main

import (
	"fmt"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newArchiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "archive <ref>",
		Short: "Archive a JD item",
		Long:  "Archive an ID (S01.11.11) or category (S01.11) to its parent's archive folder.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			result, err := vault.Archive(v, args[0])
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "Archived %s\n", result.Ref)
			fmt.Fprintf(w, "New path: %s\n", result.NewPath)
			return nil
		},
	}
}
