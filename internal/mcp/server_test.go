package mcp

import (
	"testing"

	"github.com/pauvalls/grimorio/internal/config"
)

func TestNewServer(t *testing.T) {
	cfg := config.DefaultConfig()
	server, shutdown := NewServer(cfg)
	if server == nil {
		t.Fatal("NewServer() returned nil")
	}
	if shutdown == nil {
		t.Fatal("NewServer() returned nil shutdown")
	}
	_ = shutdown()
}
