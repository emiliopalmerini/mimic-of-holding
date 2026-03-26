package main

import (
	"fmt"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newWriteCmd() *cobra.Command {
	var template string

	cmd := &cobra.Command{
		Use:   "write <id> <filename> [content]",
		Short: "Write a file inside a JD ID",
		Long:  "Create or overwrite a file inside a JD ID folder. Content is optional when --template is provided.",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			content := ""
			if len(args) == 3 {
				content = args[2]
			}
			path, err := vault.WriteFile(v, args[0], args[1], content, template)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Written %s\n", path)
			return nil
		},
	}
	cmd.Flags().StringVar(&template, "template", "", "template name to use for file content (used when content is empty)")
	return cmd
}
