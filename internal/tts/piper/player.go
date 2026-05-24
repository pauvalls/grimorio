package piper

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

// AudioPlayer reproduces WAV audio chunks sequentially via a system audio player.
type AudioPlayer interface {
	Enqueue(wav io.Reader) error
	Skip() error
	Stop() error
	Pause() error
	Resume() error
	IsPlaying() bool
	Close() error
}

// Player implements AudioPlayer with a FIFO queue and system player detection.
type Player struct {
	player   string // "aplay", "paplay", "ffplay", or ""
	device   string // optional audio device
	preferred string // user-preferred player

	// Injectable for testing
	lookPath    func(string) (string, error)
	startCmd    func(name string, arg ...string) *exec.Cmd
	playChunkFn func(io.Reader) error

	mu        sync.RWMutex
	queue     []io.Reader
	playing   bool
	paused    bool
	stopCh    chan struct{}
	skipCh    chan struct{}
	wg        sync.WaitGroup
	closed    bool
}

// NewPlayer creates an AudioPlayer, auto-detecting the best available system player.
func NewPlayer(preferred, device string) *Player {
	return newPlayerWithDeps(preferred, device, exec.LookPath, exec.Command)
}

func newPlayerWithDeps(preferred, device string, lookPath func(string) (string, error), startCmd func(string, ...string) *exec.Cmd) *Player {
	p := &Player{
		preferred: preferred,
		device:    device,
		lookPath:  lookPath,
		startCmd:  startCmd,
		stopCh:    make(chan struct{}),
		skipCh:    make(chan struct{}),
	}
	p.player = p.detectPlayer(preferred)
	return p
}

func (p *Player) playChunkDelegate(wav io.Reader) error {
	if p.playChunkFn != nil {
		return p.playChunkFn(wav)
	}
	return p.playChunk(wav)
}

func (p *Player) detectPlayer(preferred string) string {
	if preferred != "" && preferred != "auto" {
		if _, err := p.lookPath(preferred); err == nil {
			return preferred
		}
	}

	candidates := []string{"aplay", "paplay", "ffplay"}
	for _, c := range candidates {
		if _, err := p.lookPath(c); err == nil {
			return c
		}
	}
	return ""
}

// Enqueue adds a WAV chunk to the playback queue.
func (p *Player) Enqueue(wav io.Reader) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return errors.New("player: closed")
	}
	if p.player == "" {
		return errors.New("player: no audio player available")
	}

	p.queue = append(p.queue, wav)
	if !p.playing {
		p.playing = true
		p.wg.Add(1)
		go p.playLoop()
	}
	return nil
}

// Skip interrupts the currently playing chunk and moves to the next.
func (p *Player) Skip() error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if !p.playing {
		return nil
	}
	select {
	case p.skipCh <- struct{}{}:
	default:
	}
	return nil
}

// Stop halts playback and clears the queue.
func (p *Player) Stop() error {
	p.mu.Lock()
	if !p.playing {
		p.mu.Unlock()
		return nil
	}

	p.queue = nil
	p.playing = false
	ch := p.stopCh
	p.stopCh = make(chan struct{}) // prepare for next use
	p.mu.Unlock()

	close(ch)
	p.wg.Wait()
	return nil
}

// Pause pauses playback after the current chunk finishes.
func (p *Player) Pause() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.paused = true
	return nil
}

// Resume resumes playback from pause.
func (p *Player) Resume() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.paused {
		return nil
	}
	p.paused = false
	if !p.playing && len(p.queue) > 0 {
		p.playing = true
		p.wg.Add(1)
		go p.playLoop()
	}
	return nil
}

// IsPlaying returns true if the player is actively reproducing audio.
func (p *Player) IsPlaying() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.playing
}

// Close stops playback and releases resources.
func (p *Player) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	p.queue = nil
	playing := p.playing
	var ch chan struct{}
	if playing {
		ch = p.stopCh
		p.stopCh = make(chan struct{})
	}
	p.mu.Unlock()

	if playing {
		close(ch)
		p.wg.Wait()
	}
	return nil
}

func (p *Player) playLoop() {
	defer p.wg.Done()

	for {
		p.mu.Lock()
		if p.paused || len(p.queue) == 0 {
			p.playing = false
			p.mu.Unlock()
			return
		}
		wav := p.queue[0]
		p.queue = p.queue[1:]
		p.mu.Unlock()

		if err := p.playChunkDelegate(wav); err != nil {
			// Log and continue to next chunk
			fmt.Fprintf(os.Stderr, "player: chunk playback error: %v\n", err)
		}
	}
}

func (p *Player) playChunk(wav io.Reader) error {
	args := p.playerArgs()
	cmd := p.startCmd(p.player, args...)
	cmd.Stdin = wav

	// Wire stop and skip signals
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		return err
	case <-p.stopCh:
		_ = cmd.Process.Signal(os.Interrupt)
		return errors.New("player: stopped")
	case <-p.skipCh:
		_ = cmd.Process.Signal(os.Interrupt)
		return errors.New("player: skipped")
	}
}

func (p *Player) playerArgs() []string {
	switch p.player {
	case "aplay":
		if p.device != "" {
			return []string{"-D", p.device, "-"}
		}
		return []string{"-"}
	case "paplay":
		if p.device != "" {
			return []string{"--device=" + p.device, "-"}
		}
		return []string{"-"}
	case "ffplay":
		return []string{"-nodisp", "-autoexit", "-"}
	default:
		return []string{"-"}
	}
}
