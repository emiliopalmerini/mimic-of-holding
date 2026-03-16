package main

import (
	"fmt"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <category> <name>",
		Short: "Create a new JD ID",
		Long:  "Create a new ID in the given category (e.g., mimic create S01.11 Cinema).",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			result, err := vault.Create(v, args[0], args[1])
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "Created %s %s\n", result.Ref, result.Name)
			fmt.Fprintf(w, "Path: %s\n", result.Path)
			return nil
		},
	}
}
