package main

import (
	"fmt"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newReadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "read <ref> [file]",
		Short: "Read any JD level or a file within an ID",
		Long:  "Read a scope (S01), area (S01.10-19), category (S01.11), ID (S01.11.11), or a specific file within an ID.",
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
			result, err := vault.Read(v, args[0], file)
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "# %s %s\n", result.Ref, result.Name)
			fmt.Fprintf(w, "Path: %s\n\n", result.Path)
			if result.Content != "" {
				fmt.Fprintln(w, result.Content)
			}
			if len(result.Children) > 0 {
				fmt.Fprintln(w, "--- Children ---")
				for _, c := range result.Children {
					fmt.Fprintf(w, "  %s\n", c)
				}
			}
			if len(result.Files) > 0 {
				fmt.Fprintln(w, "\n--- Files ---")
				for _, f := range result.Files {
					fmt.Fprintf(w, "  %s\n", f)
				}
			}
			return nil
		},
	}
}
