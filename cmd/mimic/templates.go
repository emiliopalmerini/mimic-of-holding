package main

import (
	"fmt"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newTemplatesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "templates <category-ref>",
		Short: "List available templates for a category",
		Long:  "List templates from the category's .03 Templates ID, with hierarchical lookup through area and scope levels.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			templates, err := vault.ListTemplates(v, args[0])
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			if len(templates) == 0 {
				fmt.Fprintf(w, "No templates found for %s\n", args[0])
				return nil
			}
			fmt.Fprintf(w, "Templates for %s:\n", args[0])
			for _, t := range templates {
				fmt.Fprintf(w, "  %s (%s, %s)\n", t.Name, t.SourceRef, t.Source)
			}
			return nil
		},
	}
}
