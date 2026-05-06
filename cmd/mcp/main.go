package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	mcpgo "github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/config"
	mcpserver "github.com/pauvalls/grimorio/internal/mcp"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println("grimorio-mcp v1.0.0")
		os.Exit(0)
	}

	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".config", "grimorio", "config.json")

	// Ensure config dir exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "creating config dir: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loading config: %v\n", err)
		os.Exit(1)
	}

	// If no config exists, create default
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := cfg.Save(configPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving default config: %v\n", err)
		}
	}

	// Start MCP stdio server
	mcpServer := mcpserver.NewServer(cfg)
	stdioServer := mcpgo.NewStdioServer(mcpServer)
	if err := stdioServer.Listen(context.Background(), os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
