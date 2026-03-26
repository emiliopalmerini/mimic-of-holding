package main

import (
	"fmt"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newRenameFileCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rename-file <id-ref> <old-filename> <new-filename>",
		Short: "Rename a file inside a JD ID folder",
		Long:  "Rename a file and update wiki links across the vault. If the file is the JDex file, the folder is also renamed.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			result, err := vault.RenameFile(v, args[0], args[1], args[2])
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "Renamed file: %q → %q\n", args[1], args[2])
			fmt.Fprintf(w, "Path: %s\n", result.NewPath)
			if result.LinksUpdated > 0 {
				fmt.Fprintf(w, "Updated %d wiki links\n", result.LinksUpdated)
			}
			if result.HeadingUpdated {
				fmt.Fprintln(w, "Heading updated: yes")
			}
			if result.FolderRenamed {
				fmt.Fprintln(w, "Folder renamed: yes")
			}
			return nil
		},
	}
}
