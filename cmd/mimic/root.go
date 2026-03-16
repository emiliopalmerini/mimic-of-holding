package main

import (
	"os"
	"path/filepath"

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
