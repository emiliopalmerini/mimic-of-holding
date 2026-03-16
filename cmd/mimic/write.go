package main

import (
	"fmt"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newWriteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "write <id> <filename> <content>",
		Short: "Write a file inside a JD ID",
		Long:  "Create or overwrite a file inside a JD ID folder (e.g., mimic write S01.12.11 Recipe.md '# Recipe').",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			path, err := vault.WriteFile(v, args[0], args[1], args[2])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Written %s\n", path)
			return nil
		},
	}
}
