package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
	"github.com/spf13/cobra"
)

var vaultPath string

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mimic",
		Short: "CLI for the Bag of Holding Obsidian vault",
	}
	cmd.PersistentFlags().StringVar(&vaultPath, "vault", "", "path to vault (default ~/Documents/bag_of_holding)")

	cmd.AddCommand(newBrowseCmd())
	cmd.AddCommand(newSearchCmd())
	cmd.AddCommand(newReadCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newArchiveCmd())
	cmd.AddCommand(newInboxCmd())
	cmd.AddCommand(newRenameCmd())
	cmd.AddCommand(newFrontmatterCmd())
	cmd.AddCommand(newMoveCmd())
	cmd.AddCommand(newMoveFileCmd())
	cmd.AddCommand(newRenameFileCmd())
	cmd.AddCommand(newTemplatesCmd())
	cmd.AddCommand(newResolveCmd())
	cmd.AddCommand(newLogCmd())

	return cmd
}

func resolveVaultPath() string {
	if vaultPath != "" {
		return vaultPath
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Documents", "bag_of_holding")
}

func parseVault() (*vault.Vault, error) {
	return vault.ParseVault(resolveVaultPath())
}

func parseVarFlags(flags []string) map[string]string {
	if len(flags) == 0 {
		return nil
	}
	vars := make(map[string]string, len(flags))
	for _, f := range flags {
		k, v, _ := strings.Cut(f, "=")
		if k != "" {
			vars[k] = v
		}
	}
	return vars
}
