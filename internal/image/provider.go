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

func NewProvider(cfg Config) (Provider, error) {
	switch cfg.Provider {
	case "dalle":
		return NewDalleProvider(cfg.DalleKey, cfg.DalleModel)
	case "pollinations":
		return NewPollinationsProvider(cfg.Width, cfg.Height, cfg.Seed), nil
	default:
		return NewPollinationsProvider(cfg.Width, cfg.Height, cfg.Seed), nil
	}
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
