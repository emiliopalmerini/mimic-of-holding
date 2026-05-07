package main

import (
	"fmt"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	var template string
	var varFlags []string
	var categoryName, areaName string

	cmd := &cobra.Command{
		Use:   "create <category> <name>",
		Short: "Create a new JD ID",
		Long:  "Create a new ID in the given category (e.g., mimic create S01.11 Cinema). Auto-creates area and category if they don't exist.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			customVars := parseVarFlags(varFlags)
			opts := vault.CreateOpts{
				CategoryName: categoryName,
				AreaName:     areaName,
			}
			result, err := vault.Create(v, args[0], args[1], template, customVars, opts)
			if err != nil {
				return err
			}
			autoLog(cmd.ErrOrStderr(), v, scopeFromRef(result.Ref), "create",
				fmt.Sprintf("%s %s", result.Ref, result.Name), "", "Created JDex.")
			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "Created %s %s\n", result.Ref, result.Name)
			fmt.Fprintf(w, "Path: %s\n", result.Path)
			return nil
		},
	}
	cmd.Flags().StringVar(&template, "template", "", "template name to use for the JDex file")
	cmd.Flags().StringArrayVar(&varFlags, "var", nil, "custom template variable as key=value (repeatable)")
	cmd.Flags().StringVar(&categoryName, "category-name", "", "name for auto-created category (if it doesn't exist)")
	cmd.Flags().StringVar(&areaName, "area-name", "", "name for auto-created area (if it doesn't exist)")
	return cmd
}
