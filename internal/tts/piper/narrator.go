package piper

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/pauvalls/grimorio/internal/dm"
)

// NarratorPipeline orquesta filter → chunk → synthesize → play con precarga.
type NarratorPipeline interface {
	Narrate(ctx context.Context, text string) error
	Stop() error
}

// Narrator implements NarratorPipeline.
type Narrator struct {
	filter TextFilter
	engine TTSEngine
	player AudioPlayer
	maxLen int

	mu      sync.RWMutex
	running bool
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// NewNarrator crea un nuevo Narrator con las dependencias dadas.
func NewNarrator(filter TextFilter, engine TTSEngine, player AudioPlayer, maxLen int) *Narrator {
	return &Narrator{
		filter: filter,
		engine: engine,
		player: player,
		maxLen: maxLen,
	}
}

// Narrate filtra, chunkea, sintetiza y reproduce el texto dado.
func (n *Narrator) Narrate(ctx context.Context, text string) error {
	// 1. Filter
	filtered := n.filter.Filter(text)
	filtered = strings.TrimSpace(filtered)
	if filtered == "" {
		return nil
	}

	// 2. Chunk
	chunks := dm.SplitIntoChunks(filtered, n.maxLen)
	if len(chunks) == 0 {
		return nil
	}

	// Stop any previous narration
	_ = n.Stop()

	narrateCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	n.mu.Lock()
	n.cancel = cancel
	n.running = true
	n.mu.Unlock()

	// Channels for the pipeline
	textCh := make(chan string, len(chunks))
	for _, c := range chunks {
		textCh <- c.Text
	}
	close(textCh)

	// Buffered channel for precarga de 1 chunk
	audioCh := make(chan io.ReadCloser, 1)

	// Synthesizer worker
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		defer close(audioCh)

		for txt := range textCh {
			select {
			case <-narrateCtx.Done():
				return
			default:
			}

			wav, err := n.engine.Synthesize(narrateCtx, txt)
			if err != nil {
				fmt.Fprintf(os.Stderr, "narrator: synthesize error: %v\n", err)
				continue
			}

			select {
			case audioCh <- wav:
			case <-narrateCtx.Done():
				wav.Close()
				return
			}
		}
	}()

	// Player worker
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()

		for wav := range audioCh {
			select {
			case <-narrateCtx.Done():
				wav.Close()
				return
			default:
			}

			if err := n.player.Enqueue(wav); err != nil {
				fmt.Fprintf(os.Stderr, "narrator: enqueue error: %v\n", err)
				wav.Close()
				continue
			}
		}

		n.mu.Lock()
		n.running = false
		n.mu.Unlock()
	}()

	n.wg.Wait()
	return nil
}

// Stop cancels the current narration and waits for workers to finish.
func (n *Narrator) Stop() error {
	n.mu.Lock()
	if n.cancel != nil {
		n.cancel()
	}
	n.mu.Unlock()

	n.wg.Wait()

	// Also stop the audio player
	if n.player != nil {
		_ = n.player.Stop()
	}

	return nil
}

// IsRunning returns true if a narration is in progress.
func (n *Narrator) IsRunning() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.running
}
