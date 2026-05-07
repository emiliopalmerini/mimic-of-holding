package main

import (
	"fmt"
	"strings"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	var top int
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search the vault",
		Long: `Search by JD reference (S01.11), name (Entertainment), or content.

Content search forms:
  ?<query>   substring match per line (every matching line, in vault order)
  ??<query>  BM25-ranked (top files by relevance, with snippet)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			query := args[0]
			opts := vault.SearchOpts{}

			// ?? prefix → BM25 ranked content search.
			// ?  prefix → substring content search.
			if after, found := strings.CutPrefix(query, "??"); found {
				query = after
				opts.Content = true
				opts.Ranked = true
				opts.Top = top
			} else if after, found := strings.CutPrefix(query, "?"); found {
				query = after
				opts.Content = true
			}

			results, err := vault.Search(v, query, opts)
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			if len(results) == 0 {
				fmt.Fprintln(w, "No results found.")
				return nil
			}
			for _, r := range results {
				if opts.Ranked {
					fmt.Fprintf(w, "[%s] %s  %s  (score %.2f)\n", r.Type, r.Ref, r.Name, r.Score)
					if r.Snippet != "" {
						fmt.Fprintf(w, "  > %s\n", r.Snippet)
					}
					continue
				}
				fmt.Fprintf(w, "[%s] %s  %s\n", r.Type, r.Ref, r.Name)
				if r.MatchLine != "" {
					fmt.Fprintf(w, "  > %s\n", r.MatchLine)
				}
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&top, "top", 10, "max results in ?? (BM25) mode")
	return cmd
}
