package main

import (
	"fmt"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newBrowseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "browse [filter]",
		Short: "Display the vault tree",
		Long:  "Display the JD hierarchy. Optional filter: scope (S01), area (S01.10-19), or category (S01.11).",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			filter := ""
			if len(args) == 1 {
				filter = args[0]
			}
			out, err := vault.Browse(v, filter)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), out)
			return nil
		},
	}
}
