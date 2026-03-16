package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	vaultFlag := flag.String("vault", "", "path to vault (default ~/Documents/bag_of_holding)")
	flag.Parse()

	vaultRoot := *vaultFlag
	if vaultRoot == "" {
		home, _ := os.UserHomeDir()
		vaultRoot = filepath.Join(home, "Documents", "bag_of_holding")
	}

	s := newServer(vaultRoot)
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
