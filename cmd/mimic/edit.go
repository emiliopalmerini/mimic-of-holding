package main

import (
	"fmt"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit <id> <filename> <old_string> <new_string>",
		Short: "Search-and-replace edit a file inside a JD ID",
		Long:  "Replace the first (and only) occurrence of old_string with new_string in a file within a JD ID folder.",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			path, err := vault.EditFile(v, args[0], args[1], args[2], args[3])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Edited %s\n", path)
			return nil
		},
	}
}
