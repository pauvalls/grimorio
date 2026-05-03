package image

import (
	"fmt"
	"os"
	"path/filepath"
)

type Provider interface {
	Generate(prompt string) ([]byte, error)
	IsConfigured() bool
	Name() string
}

type Config struct {
	Provider   string `json:"image_provider"`
	DalleKey   string `json:"dalle_api_key,omitempty"`
	DalleModel string `json:"dalle_model,omitempty"`
	Width      int    `json:"image_width"`
	Height     int    `json:"image_height"`
	Seed       int    `json:"image_seed"`
}

func DefaultConfig() Config {
	return Config{
		Provider:   "pollinations",
		DalleModel: "dall-e-3",
		Width:      1024,
		Height:     1024,
		Seed:       -1,
	}
}

// NewProvider creates a single provider by name
func NewProvider(cfg Config) (Provider, error) {
	switch cfg.Provider {
	case "dalle":
		return NewDalleProvider(cfg.DalleKey, cfg.DalleModel)
	case "raphael":
		return NewRaphaelProvider(), nil
	case "pollinations":
		return NewPollinationsProvider(cfg.Width, cfg.Height, cfg.Seed), nil
	default:
		return NewPollinationsProvider(cfg.Width, cfg.Height, cfg.Seed), nil
	}
}

// NewProviderChain creates a chain of fallback providers
// Tries primary first, then falls back to alternatives
func NewProviderChain(cfg Config) []Provider {
	var providers []Provider

	// Primary provider
	primary, err := NewProvider(cfg)
	if err == nil {
		providers = append(providers, primary)
	}

	// Always add free fallbacks regardless of primary
	// Raphael (free, unlimited)
	providers = append(providers, NewRaphaelProvider())
	// Pollinations (free, default)
	providers = append(providers, NewPollinationsProvider(cfg.Width, cfg.Height, cfg.Seed))

	return providers
}

// GenerateWithFallback tries multiple providers until one succeeds
func GenerateWithFallback(providers []Provider, prompt string) ([]byte, string, error) {
	var lastErr error
	for _, p := range providers {
		imgData, err := p.Generate(prompt)
		if err == nil {
			return imgData, p.Name(), nil
		}
		lastErr = err
	}
	return nil, "", fmt.Errorf("all providers failed; last error: %w", lastErr)
}

func GenerateAndSave(p Provider, prompt, outputPath string) error {
	imgData, err := p.Generate(prompt)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return os.WriteFile(outputPath, imgData, 0644)
}

// GenerateAndSaveWithFallback tries multiple providers and saves the first successful result
func GenerateAndSaveWithFallback(providers []Provider, prompt, outputPath string) (string, error) {
	imgData, providerName, err := GenerateWithFallback(providers, prompt)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, imgData, 0644); err != nil {
		return "", err
	}

	return providerName, nil
}
