package main

import (
	"fmt"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newReadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "read <id>",
		Short: "Read a JDex entry",
		Long:  "Display the JDex entry and file listing for a JD ID (e.g., S01.11.11).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			result, err := vault.Read(v, args[0])
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "# %s %s\n", result.Ref, result.Name)
			fmt.Fprintf(w, "Path: %s\n\n", result.Path)
			if result.JDex != "" {
				fmt.Fprintf(w, "--- JDex ---\n%s\n", result.JDex)
			} else {
				fmt.Fprintln(w, "(no JDex file)")
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
