package main

import (
	"fmt"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newMoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "move <ref> <target>",
		Short: "Move a JD item to a different parent",
		Long:  "Move an ID to a different category, or a category to a different area.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			result, err := vault.Move(v, args[0], args[1])
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "Moved %s → %s\n", result.OldRef, result.NewRef)
			fmt.Fprintf(w, "Path: %s\n", result.NewPath)
			if result.LinksUpdated > 0 {
				fmt.Fprintf(w, "Updated %d wiki links\n", result.LinksUpdated)
			}
			return nil
		},
	}
}

func newMoveFileCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "move-file <from-id> <filename> <to-id>",
		Short: "Move a file between JD IDs",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			path, err := vault.MoveFile(v, args[0], args[1], args[2])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Moved to %s\n", path)
			return nil
		},
	}
}
