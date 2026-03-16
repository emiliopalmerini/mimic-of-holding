package main

import (
	"fmt"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search the vault",
		Long:  "Search by JD reference (S01.11), name (Entertainment), or content (?pasta recipe).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			results, err := vault.Search(v, args[0])
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			if len(results) == 0 {
				fmt.Fprintln(w, "No results found.")
				return nil
			}
			for _, r := range results {
				fmt.Fprintf(w, "[%s] %s  %s\n", r.Type, r.Ref, r.Name)
				if r.MatchLine != "" {
					fmt.Fprintf(w, "  > %s\n", r.MatchLine)
				}
			}
			return nil
		},
	}
}
