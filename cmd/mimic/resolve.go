package main

import (
	"fmt"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newResolveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resolve <ref> [file]",
		Short: "Resolve a JD reference to a filesystem path",
		Long:  "Translate a JD reference (scope, area, category, or ID) to its absolute filesystem path. Optionally append a filename for ID references.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			file := ""
			if len(args) == 2 {
				file = args[1]
			}
			path, err := vault.Resolve(v, args[0], file)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
}
