package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/config"
	mcpserver "github.com/pauvalls/grimorio/internal/mcp"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("grimorio %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// Parse flags
	compilerVersion := 0
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--compiler-version" && i+1 < len(os.Args) {
			if v, err := strconv.Atoi(os.Args[i+1]); err == nil && (v == 1 || v == 2) {
				compilerVersion = v
			}
			i++
		}
	}

	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".config", "grimorio", "config.json")

	// Ensure config dir exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config dir: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Override compiler version from flag
	if compilerVersion != 0 {
		cfg.CompilerVersion = compilerVersion
	}

	// If no config exists, create default
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := cfg.Save(configPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving default config: %v\n", err)
		}
	}

	// Start MCP stdio server
	mcpServer := mcpserver.NewServer(cfg)
	stdioServer := server.NewStdioServer(mcpServer)
	if err := stdioServer.Listen(context.Background(), os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
