package main

import (
	"fmt"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newInboxCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "inbox [scope]",
		Short: "List inbox items",
		Long:  "List files in inbox folders across the vault. Optional scope filter (S01).",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			scopeFilter := ""
			if len(args) == 1 {
				scopeFilter = args[0]
			}
			items, err := vault.Inbox(v, scopeFilter)
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			if len(items) == 0 {
				fmt.Fprintln(w, "All inboxes are empty.")
				return nil
			}
			currentRef := ""
			for _, item := range items {
				if item.InboxRef != currentRef {
					if currentRef != "" {
						fmt.Fprintln(w)
					}
					fmt.Fprintf(w, "%s (%s)\n", item.InboxRef, item.InboxName)
					currentRef = item.InboxRef
				}
				fmt.Fprintf(w, "  %s\n", item.File)
			}
			return nil
		},
	}
}
