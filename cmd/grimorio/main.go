package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	mcpgo "github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/config"
	"github.com/pauvalls/grimorio/internal/game"
	mcpserver "github.com/pauvalls/grimorio/internal/mcp"
	"github.com/pauvalls/grimorio/internal/repository"
	gameserver "github.com/pauvalls/grimorio/internal/server"
	"github.com/pauvalls/grimorio/internal/websocket"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version":
			fmt.Println("grimorio v1.0.0")
			os.Exit(0)
		case "serve":
			if err := runServe(); err != nil {
				fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	// Default: MCP stdio server
	if err := runMCP(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func runMCP() error {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".config", "grimorio", "config.json")

	// Ensure config dir exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
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
	return stdioServer.Listen(context.Background(), os.Stdin, os.Stdout)
}

func runServe() error {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".config", "grimorio", "config.json")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Create repositories
	campaignRepo := repository.NewFilesystemCampaignRepository(cfg.OutputDir)
	charRepo := repository.NewFilesystemCharacterRepository(cfg.OutputDir)
	questRepo := repository.NewFilesystemQuestRepository(cfg.OutputDir)
	
	// Create session repository (SQLite)
	sessionsDir := filepath.Join(cfg.OutputDir, ".sessions")
	os.MkdirAll(sessionsDir, 0755)
	sessionRepo, err := repository.NewSQLiteSessionRepository(filepath.Join(sessionsDir, "game.db"))
	if err != nil {
		return fmt.Errorf("creating session repository: %w", err)
	}

	// Create LLM client
	var llmClient game.LLMClient
	if cfg.GameEngine.APIKey != "" {
		llmClient, err = game.NewLLMClient(cfg.GameEngine)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create LLM client: %v\n", err)
		}
	}

	// Create game engine
	engine := game.NewEngine(sessionRepo, campaignRepo, charRepo, questRepo, llmClient)

	// Create WebSocket hub
	hub := websocket.NewHub(engine)
	go hub.Run()

	// Start HTTP server
	return gameserver.Start(cfg, engine, hub)
}
