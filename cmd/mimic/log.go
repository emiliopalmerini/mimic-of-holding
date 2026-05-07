package main

import (
	"fmt"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newLogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Manage the per-scope activity log",
	}
	cmd.AddCommand(newLogAppendCmd())
	cmd.AddCommand(newLogTailCmd())
	return cmd
}

func newLogAppendCmd() *cobra.Command {
	var secondary, details string
	cmd := &cobra.Command{
		Use:   "append <scope> <op> <target>",
		Short: "Append an entry to a scope's activity log",
		Long: `Append an entry to a scope's activity log file.

Scope is one of S01, S02, S03. Op must be one of: create, archive, move,
move-file, rename, rename-file, frontmatter, ingest, lint.`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			return vault.Log(v, args[0], args[1], args[2], secondary, details)
		},
	}
	cmd.Flags().StringVar(&secondary, "secondary", "", "secondary target (renders as '→ <secondary>' in the header)")
	cmd.Flags().StringVar(&details, "details", "", "details body appended below the header")
	return cmd
}

func newLogTailCmd() *cobra.Command {
	var n int
	cmd := &cobra.Command{
		Use:   "tail <scope>",
		Short: "Show the last N entries of a scope's activity log",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			entries, err := vault.LogTail(v, args[0], n)
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			for i, e := range entries {
				if i > 0 {
					fmt.Fprintln(w)
				}
				fmt.Fprintln(w, e)
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&n, "n", "n", 10, "number of entries to return")
	return cmd
}
