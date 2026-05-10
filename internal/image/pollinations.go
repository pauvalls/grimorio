package image

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type PollinationsProvider struct {
	BaseURL string
	Width   int
	Height  int
	Seed    int
	Model   string
	client  *http.Client
}

const pollinationsTimeout = 5 * time.Minute

func NewPollinationsProvider(width, height, seed int) *PollinationsProvider {
	if width <= 0 {
		width = 1024
	}
	if height <= 0 {
		height = 1024
	}
	return &PollinationsProvider{
		BaseURL: "https://image.pollinations.ai/prompt",
		Width:   width,
		Height:  height,
		Seed:    seed,
		Model:   "flux",
		client:  &http.Client{Timeout: pollinationsTimeout},
	}
}

func (p *PollinationsProvider) Name() string {
	return "pollinations"
}

func (p *PollinationsProvider) IsConfigured() bool {
	return true
}

func (p *PollinationsProvider) Generate(prompt string) ([]byte, error) {
	encodedPrompt := url.QueryEscape(prompt)

	apiURL := fmt.Sprintf("%s/%s?width=%d&height=%d&nologo=true",
		p.BaseURL, encodedPrompt, p.Width, p.Height)

	if p.Seed >= 0 {
		apiURL += fmt.Sprintf("&seed=%d", p.Seed)
	}

	if p.Model != "" {
		apiURL += fmt.Sprintf("&model=%s", p.Model)
	}

	resp, err := p.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("pollinations request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("pollinations API error %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}
