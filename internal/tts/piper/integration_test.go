package piper

import (
	"context"
	"io"
	"strings"
	"sync"
	"testing"
)

// TestIntegrationPipelineEndToEnd verifies the full pipeline:
// text → filter → chunk → synthesize → play, in correct order.
func TestIntegrationPipelineEndToEnd(t *testing.T) {
	filter := &DefaultTextFilter{}

	engine := &integrationEngine{}
	player := &integrationPlayer{}

	// maxLen 20 forces splitting into multiple chunks
	n := NewNarrator(filter, engine, player, 20)

	text := `El dragón rojo exhala fuego.
| Stat | Valor |
| HP   | 250   |
Un guardia se acerca.`

	if err := n.Narrate(context.Background(), text); err != nil {
		t.Fatalf("Narrate() error = %v", err)
	}

	// Verify filter removed table lines
	engine.mu.Lock()
	for _, call := range engine.calls {
		if strings.Contains(call, "|") {
			t.Errorf("engine should not receive table lines, got: %q", call)
		}
	}
	engine.mu.Unlock()

	// Verify playback order matches chunk order
	player.mu.Lock()
	if len(player.enqueued) < 2 {
		t.Fatalf("expected at least 2 enqueued chunks, got %d: %v", len(player.enqueued), player.enqueued)
	}

	// First chunk should be the narration before the table
	if !strings.Contains(player.enqueued[0], "dragón") {
		t.Errorf("expected first chunk to contain 'dragón', got %q", player.enqueued[0])
	}
	// Last chunk should be the narration after the table
	last := player.enqueued[len(player.enqueued)-1]
	if !strings.Contains(last, "guardia") {
		t.Errorf("expected last chunk to contain 'guardia', got %q", last)
	}
	player.mu.Unlock()
}

// TestIntegrationPipelineOrder verifies strict FIFO order through the pipeline.
func TestIntegrationPipelineOrder(t *testing.T) {
	filter := &mockFilter{}
	engine := &integrationEngine{}
	player := &integrationPlayer{}

	n := NewNarrator(filter, engine, player, 15)

	text := "One. Two. Three. Four."
	if err := n.Narrate(context.Background(), text); err != nil {
		t.Fatalf("Narrate() error = %v", err)
	}

	player.mu.Lock()
	if len(player.enqueued) != 4 {
		t.Fatalf("expected 4 chunks, got %d: %v", len(player.enqueued), player.enqueued)
	}
	for i, expected := range []string{"One.", "Two.", "Three.", "Four."} {
		if player.enqueued[i] != expected {
			t.Errorf("chunk[%d] = %q, want %q", i, player.enqueued[i], expected)
		}
	}
	player.mu.Unlock()
}

// TestIntegrationSkipsTables verifies that markdown tables are completely
// removed and do not produce audio chunks.
func TestIntegrationSkipsTables(t *testing.T) {
	filter := &DefaultTextFilter{}
	engine := &integrationEngine{}
	player := &integrationPlayer{}

	n := NewNarrator(filter, engine, player, 150)

	text := `| Nombre | HP |
|--------|----|
| Goblin | 7  |
El goblin ataca.`

	if err := n.Narrate(context.Background(), text); err != nil {
		t.Fatalf("Narrate() error = %v", err)
	}

	player.mu.Lock()
	if len(player.enqueued) != 1 {
		t.Fatalf("expected 1 chunk (only post-table text), got %d: %v", len(player.enqueued), player.enqueued)
	}
	if !strings.Contains(player.enqueued[0], "goblin ataca") {
		t.Errorf("expected chunk to contain 'goblin ataca', got %q", player.enqueued[0])
	}
	player.mu.Unlock()
}

// ---- Integration test helpers ----

type integrationEngine struct {
	mu    sync.Mutex
	calls []string
}

func (m *integrationEngine) Synthesize(ctx context.Context, text string) (io.ReadCloser, error) {
	m.mu.Lock()
	m.calls = append(m.calls, text)
	m.mu.Unlock()
	return io.NopCloser(strings.NewReader(text)), nil
}

func (m *integrationEngine) HealthCheck(ctx context.Context) error { return nil }
func (m *integrationEngine) Close() error                           { return nil }

type integrationPlayer struct {
	mu       sync.Mutex
	enqueued []string
}

func (m *integrationPlayer) Enqueue(wav io.Reader) error {
	data, _ := io.ReadAll(wav)
	m.mu.Lock()
	m.enqueued = append(m.enqueued, string(data))
	m.mu.Unlock()
	return nil
}

func (m *integrationPlayer) Skip() error     { return nil }
func (m *integrationPlayer) Stop() error     { return nil }
func (m *integrationPlayer) Pause() error    { return nil }
func (m *integrationPlayer) Resume() error   { return nil }
func (m *integrationPlayer) IsPlaying() bool { return false }
func (m *integrationPlayer) Close() error    { return nil }
