package main

import (
	"fmt"
	"strings"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newReadCmd() *cobra.Command {
	var deep bool
	cmd := &cobra.Command{
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
			var result *vault.ReadResult
			if deep {
				result, err = vault.ReadDeep(v, args[0], file)
			} else {
				result, err = vault.Read(v, args[0], file)
			}
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			printReadResult(w, result, 0)
			return nil
		},
	}
	cmd.Flags().BoolVar(&deep, "deep", false, "Recursively include all descendant content")
	return cmd
}

func printReadResult(w interface{ Write([]byte) (int, error) }, result *vault.ReadResult, indent int) {
	prefix := ""
	for range indent {
		prefix += "  "
	}
	fmt.Fprintf(w, "%s# %s %s\n", prefix, result.Ref, result.Name)
	if result.Content != "" {
		fmt.Fprintf(w, "%s%s\n", prefix, result.Content)
	}
	if len(result.Children) > 0 && len(result.DeepChildren) == 0 {
		for _, c := range result.Children {
			fmt.Fprintf(w, "%s  %s\n", prefix, c)
		}
	}
	if len(result.Files) > 0 {
		fmt.Fprintf(w, "%sFiles: %s\n", prefix, strings.Join(result.Files, ", "))
	}
	for _, child := range result.DeepChildren {
		printReadResult(w, &child, indent+1)
	}
}
