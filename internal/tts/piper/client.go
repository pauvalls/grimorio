package piper

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TTSEngine synthesizes plain text into WAV audio.
type TTSEngine interface {
	Synthesize(ctx context.Context, text string) (io.ReadCloser, error)
	HealthCheck(ctx context.Context) error
	Close() error
}

// Client is an HTTP client for the Piper local TTS server.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a Piper HTTP client targeting the given host:port.
func NewClient(host string, port int) *Client {
	return &Client{
		baseURL: fmt.Sprintf("http://%s:%d", host, port),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewClientWithHTTP allows injecting a custom http.Client (useful for testing).
func NewClientWithHTTP(baseURL string, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// Synthesize sends text to the Piper HTTP server and returns the WAV audio stream.
func (c *Client) Synthesize(ctx context.Context, text string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader([]byte(text)))
	if err != nil {
		return nil, fmt.Errorf("piper: create synthesize request: %w", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("piper: synthesize request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("piper: synthesize returned %d: %s", resp.StatusCode, string(body))
	}

	return resp.Body, nil
}

// HealthCheck verifies the Piper server is responsive.
func (c *Client) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL, nil)
	if err != nil {
		return fmt.Errorf("piper: create health request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("piper: health check failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("piper: health check returned %d", resp.StatusCode)
	}
	return nil
}

// Close is a no-op for the HTTP client but satisfies the TTSEngine interface.
func (c *Client) Close() error {
	return nil
}
