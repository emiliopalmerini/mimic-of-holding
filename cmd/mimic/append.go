package main

import (
	"fmt"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newAppendCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "append <id> <filename> <content>",
		Short: "Append content to a file inside a JD ID",
		Long:  "Append content to a file inside a JD ID folder. Creates the file if it doesn't exist.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			path, err := vault.AppendFile(v, args[0], args[1], args[2])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Appended to %s\n", path)
			return nil
		},
	}
}
