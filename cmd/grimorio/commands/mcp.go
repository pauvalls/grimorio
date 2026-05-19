package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pauvalls/grimorio/internal/config"
	mcpserver "github.com/pauvalls/grimorio/internal/mcp"
	"github.com/urfave/cli/v2"
	"github.com/mark3labs/mcp-go/server"
)

// runMCPServer loads config and starts the MCP stdio server.
// It is used as the default action (backward compat) and by the "mcp" subcommand.
func RunMCPServer(cCtx *cli.Context) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	configPath := filepath.Join(home, ".config", "grimorio", "config.json")

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("error creating config dir: %w", err)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	// Override compiler version from flag
	if cv := cCtx.Int("compiler-version"); cv == 1 || cv == 2 {
		cfg.CompilerVersion = cv
	}

	// Save default config if none existed
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if saveErr := cfg.Save(configPath); saveErr != nil {
			return fmt.Errorf("error saving default config: %w", saveErr)
		}
	}

	mcpSrv := mcpserver.NewServer(cfg)
	stdioServer := server.NewStdioServer(mcpSrv)
	if err := stdioServer.Listen(context.Background(), os.Stdin, os.Stdout); err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}
