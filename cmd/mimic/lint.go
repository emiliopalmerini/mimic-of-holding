package main

import (
	"fmt"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

func newLintCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lint [scope]",
		Short: "Run deterministic structural checks against the vault",
		Long: `Lint the vault for missing JDex files, broken frontmatter, name mismatches,
numbering gaps, broken wiki links, and orphan IDs.

Without arguments, lints the entire vault. With a scope (S01, S02, S03), only
issues whose source ref is in that scope are reported, but the link / target
index spans the whole vault so cross-scope references resolve correctly.

Exit code: 0 if clean, 1 if any issues, 2 on input/runtime error.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := parseVault()
			if err != nil {
				return err
			}
			scope := ""
			if len(args) == 1 {
				scope = args[0]
			}
			result, err := vault.Lint(v, scope)
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), vault.FormatLintReport(result))
			if scope != "" {
				autoLog(cmd.ErrOrStderr(), v, scope, "lint", scope, "",
					vault.SummarizeLintIssues(result.Issues))
			}
			if len(result.Issues) > 0 {
				cmd.SilenceUsage = true
				cmd.SilenceErrors = true
				return errLintFindings
			}
			return nil
		},
	}
}

// errLintFindings is returned by `mimic lint` when issues are found, so that
// the process exits with a non-zero code without printing a Cobra error banner.
var errLintFindings = errLintSentinel{}

type errLintSentinel struct{}

func (errLintSentinel) Error() string { return "lint: issues found" }
