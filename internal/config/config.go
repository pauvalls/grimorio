package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	OutputDir string `json:"output_dir"`
	PDFEngine string `json:"pdf_engine"`
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		OutputDir: filepath.Join(home, "campaigns"),
		PDFEngine: "wkhtmltopdf",
	}
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.OutputDir == "" {
		home, _ := os.UserHomeDir()
		cfg.OutputDir = filepath.Join(home, "campaigns")
	}
	if cfg.PDFEngine == "" {
		cfg.PDFEngine = "wkhtmltopdf"
	}
	return &cfg, nil
}

func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
