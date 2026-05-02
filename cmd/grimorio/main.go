package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/server"
	"github.com/paupena/grimorio/internal/config"
	mcpserver "github.com/paupena/grimorio/internal/mcp"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println("grimorio v1.0.0")
		os.Exit(0)
	}

	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".config", "grimorio", "config.json")
	
	// Ensure config dir exists
	os.MkdirAll(filepath.Dir(configPath), 0755)
	
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
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
	stdioServer := server.NewStdioServer(mcpServer)
	if err := stdioServer.Listen(context.Background(), os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
