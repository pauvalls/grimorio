package piper

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// Process represents a running OS process (abstracted for testing).
type Process interface {
	Pid() int
	Signal(sig os.Signal) error
	Wait() (*os.ProcessState, error)
}

// osProcess wraps *os.Process to satisfy the Process interface.
type osProcess struct {
	p *os.Process
}

func (op *osProcess) Pid() int                { return op.p.Pid }
func (op *osProcess) Signal(sig os.Signal) error { return op.p.Signal(sig) }
func (op *osProcess) Wait() (*os.ProcessState, error) { return op.p.Wait() }

// LifecycleManager manages the Piper server process lifecycle.
type LifecycleManager struct {
	config Config

	// Injectable dependencies for testing
	lookPath   func(string) (string, error)
	startCmd   func(name string, arg ...string) *exec.Cmd
	newProcess func(cmd *exec.Cmd) (Process, error)

	mu       sync.RWMutex
	process  Process
	cmd      *exec.Cmd
	running  bool
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// NewLifecycleManager creates a new LifecycleManager with the given config.
func NewLifecycleManager(config Config) *LifecycleManager {
	return &LifecycleManager{
		config:   config,
		lookPath: exec.LookPath,
		startCmd: exec.Command,
		newProcess: func(cmd *exec.Cmd) (Process, error) {
			if err := cmd.Start(); err != nil {
				return nil, err
			}
			return &osProcess{p: cmd.Process}, nil
		},
		stopCh: make(chan struct{}),
	}
}

// IsInstalled returns true if the `piper` binary is found in PATH.
func (lm *LifecycleManager) IsInstalled() bool {
	_, err := lm.lookPath("piper")
	return err == nil
}

// Start launches the Piper server process and begins healthcheck monitoring.
func (lm *LifecycleManager) Start(ctx context.Context) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.running {
		return nil
	}

	if !lm.IsInstalled() {
		return errors.New("piper: binary not found in PATH")
	}

	if err := lm.startProcessLocked(ctx); err != nil {
		return err
	}

	lm.running = true
	lm.stopCh = make(chan struct{})
	lm.wg.Add(1)
	go lm.healthcheckLoop()

	return nil
}

func (lm *LifecycleManager) startProcessLocked(ctx context.Context) error {
	piperPath, err := lm.lookPath("piper")
	if err != nil {
		return fmt.Errorf("piper: lookup failed: %w", err)
	}

	args := []string{
		"--model", lm.config.ModelPath,
		"--port", fmt.Sprintf("%d", lm.config.Port),
		"--host", lm.config.Host,
	}
	if lm.config.ConfigPath != "" {
		args = append(args, "--config", lm.config.ConfigPath)
	}

	lm.cmd = lm.startCmd(piperPath, args...)
	lm.cmd.Stdout = os.Stdout
	lm.cmd.Stderr = os.Stderr

	proc, err := lm.newProcess(lm.cmd)
	if err != nil {
		return fmt.Errorf("piper: failed to start process: %w", err)
	}

	lm.process = proc
	return nil
}

// Stop performs a graceful shutdown of the Piper server.
func (lm *LifecycleManager) Stop(ctx context.Context) error {
	lm.mu.Lock()
	if !lm.running {
		lm.mu.Unlock()
		return nil
	}
	lm.running = false
	close(lm.stopCh)
	proc := lm.process
	lm.mu.Unlock()

	if proc == nil {
		lm.wg.Wait()
		return nil
	}

	// Send SIGTERM
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		// Process may already be gone
		lm.wg.Wait()
		return nil
	}

	// Wait up to 5 seconds for graceful shutdown
	done := make(chan struct{})
	go func() {
		lm.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		// Fall through to force kill
	case <-time.After(5 * time.Second):
		// Timeout, force kill
	}

	// Force kill
	_ = proc.Signal(syscall.SIGKILL)
	lm.wg.Wait()
	return nil
}

// IsRunning returns true if the Piper server is currently running.
func (lm *LifecycleManager) IsRunning() bool {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.running
}

// Restart stops and starts the Piper server with auto-retry backoff.
func (lm *LifecycleManager) Restart(ctx context.Context) error {
	if err := lm.Stop(ctx); err != nil {
		return err
	}

	backoffs := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}
	maxAttempts := lm.config.MaxRestarts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	if maxAttempts > len(backoffs) {
		maxAttempts = len(backoffs)
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if err := lm.Start(ctx); err == nil {
			return nil
		}
		if attempt < maxAttempts-1 {
			select {
			case <-time.After(backoffs[attempt]):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return errors.New("piper: restart failed after all attempts")
}

func (lm *LifecycleManager) healthcheckLoop() {
	defer lm.wg.Done()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	client := NewClient(lm.config.Host, lm.config.Port)
	failed := 0
	maxFails := 3

	for {
		select {
		case <-lm.stopCh:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), lm.config.HealthcheckTimeout)
			err := client.HealthCheck(ctx)
			cancel()

			if err != nil {
				failed++
				if failed >= maxFails {
					// Attempt auto-restart
					lm.mu.Lock()
					wasRunning := lm.running
					lm.mu.Unlock()
					if wasRunning {
						_ = lm.Restart(ctx)
					}
					failed = 0
				}
			} else {
				failed = 0
			}
		}
	}
}
