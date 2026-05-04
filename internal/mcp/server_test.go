package mcp

import (
	"testing"

	"github.com/pauvalls/grimorio/internal/config"
)

func TestNewServer(t *testing.T) {
	cfg := config.DefaultConfig()
	server := NewServer(cfg)
	if server == nil {
		t.Fatal("NewServer() returned nil")
	}
}
