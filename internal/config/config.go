package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pauvalls/grimorio/internal/image"
)

// TTSConfig holds TTS-related configuration for the Piper integration.
type TTSConfig struct {
	// Enabled controls whether TTS is active.
	Enabled bool `json:"enabled" env:"GRIMORIO_TTS_ENABLED"`

	// Piper-specific settings.
	Piper PiperConfig `json:"piper"`

	// Chunker controls text segmentation behavior.
	Chunker ChunkerConfig `json:"chunker"`

	// Audio controls playback behavior.
	Audio AudioConfig `json:"audio"`
}

// PiperConfig parametrizes the local Piper TTS engine.
type PiperConfig struct {
	// ModelPath is the absolute path to the Piper ONNX model.
	ModelPath string `json:"model_path" env:"PIPER_MODEL_PATH"`

	// ConfigPath is the absolute path to the model JSON config.
	// If empty, the system attempts to auto-detect <model>.onnx.json.
	ConfigPath string `json:"config_path" env:"PIPER_CONFIG_PATH"`

	// Port is the HTTP port for the Piper local server.
	Port int `json:"port" env:"PIPER_PORT"`

	// Host is the bind address for the Piper local server.
	Host string `json:"host" env:"PIPER_HOST"`

	// LengthScale controls speech speed (<1 faster, >1 slower).
	LengthScale float64 `json:"length_scale" env:"PIPER_LENGTH_SCALE"`

	// Volume controls playback volume (0.0–1.0).
	Volume float64 `json:"volume" env:"PIPER_VOLUME"`

	// CacheDir is an optional directory for caching generated WAV files.
	CacheDir string `json:"cache_dir" env:"PIPER_CACHE_DIR"`

	// MaxRestarts is the maximum number of auto-restart attempts.
	MaxRestarts int `json:"max_restarts" env:"PIPER_MAX_RESTARTS"`

	// HealthcheckTimeout is the timeout for Piper health checks.
	HealthcheckTimeout time.Duration `json:"healthcheck_timeout" env:"PIPER_HEALTHCHECK_TIMEOUT"`
}

// ChunkerConfig controls how DM text is split into TTS chunks.
type ChunkerConfig struct {
	// MaxChunkSize is the maximum characters per chunk.
	MaxChunkSize int `json:"max_chunk_size" env:"CHUNKER_MAX_SIZE"`
}

// AudioConfig controls audio playback settings.
type AudioConfig struct {
	// Player is the preferred audio player: "auto", "aplay", "paplay", "ffplay".
	Player string `json:"player" env:"AUDIO_PLAYER"`

	// Device is an optional audio device string (e.g. ALSA plughw).
	Device string `json:"device" env:"AUDIO_DEVICE"`

	// PreloadBuffer is how many chunks ahead to synthesize.
	PreloadBuffer int `json:"preload_buffer" env:"AUDIO_PRELOAD_BUFFER"`
}

type Config struct {
	OutputDir       string `json:"output_dir"`
	PDFEngine       string `json:"pdf_engine"`
	CompilerVersion int    `json:"compiler_version"`
	image.Config
	TTS TTSConfig `json:"tts"`
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		OutputDir:       filepath.Join(home, "campaigns"),
		PDFEngine:       "", // auto-detect: prefers chromium/chrome, falls back to wkhtmltopdf
		CompilerVersion: 2,
		Config:          image.DefaultConfig(),
		TTS:             DefaultTTSConfig(),
	}
}

// DefaultTTSConfig returns a TTSConfig with sensible defaults.
func DefaultTTSConfig() TTSConfig {
	return TTSConfig{
		Enabled: false,
		Piper: PiperConfig{
			ModelPath:          "",
			ConfigPath:         "",
			Port:               5000,
			Host:               "127.0.0.1",
			LengthScale:        1.0,
			Volume:             0.8,
			CacheDir:           "",
			MaxRestarts:        3,
			HealthcheckTimeout: 2 * time.Second,
		},
		Chunker: ChunkerConfig{
			MaxChunkSize: 150,
		},
		Audio: AudioConfig{
			Player:        "auto",
			Device:        "",
			PreloadBuffer: 1,
		},
	}
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			applyEnvOverrides(cfg)
			return cfg, nil
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
	// PDFEngine empty means auto-detect at runtime (compiler.New handles this)
	// No fallback here to allow the compiler to pick the best available engine
	if cfg.CompilerVersion == 0 {
		cfg.CompilerVersion = 2
	}
	if cfg.Provider == "" {
		cfg.Provider = "pollinations"
	}
	if cfg.DalleKey == "" {
		cfg.DalleKey = os.Getenv("OPENAI_API_KEY")
	}
	if cfg.DalleModel == "" {
		cfg.DalleModel = "dall-e-3"
	}
	applyTTSDefaults(&cfg)
	applyEnvOverrides(&cfg)
	return &cfg, nil
}

// applyTTSDefaults fills in zero-valued TTS fields with defaults.
func applyTTSDefaults(cfg *Config) {
	defs := DefaultTTSConfig()
	if cfg.TTS.Piper.Port == 0 {
		cfg.TTS.Piper.Port = defs.Piper.Port
	}
	if cfg.TTS.Piper.Host == "" {
		cfg.TTS.Piper.Host = defs.Piper.Host
	}
	if cfg.TTS.Piper.LengthScale == 0 {
		cfg.TTS.Piper.LengthScale = defs.Piper.LengthScale
	}
	if cfg.TTS.Piper.Volume == 0 {
		cfg.TTS.Piper.Volume = defs.Piper.Volume
	}
	if cfg.TTS.Piper.MaxRestarts == 0 {
		cfg.TTS.Piper.MaxRestarts = defs.Piper.MaxRestarts
	}
	if cfg.TTS.Piper.HealthcheckTimeout == 0 {
		cfg.TTS.Piper.HealthcheckTimeout = defs.Piper.HealthcheckTimeout
	}
	if cfg.TTS.Chunker.MaxChunkSize == 0 {
		cfg.TTS.Chunker.MaxChunkSize = defs.Chunker.MaxChunkSize
	}
	if cfg.TTS.Audio.Player == "" {
		cfg.TTS.Audio.Player = defs.Audio.Player
	}
	if cfg.TTS.Audio.PreloadBuffer == 0 {
		cfg.TTS.Audio.PreloadBuffer = defs.Audio.PreloadBuffer
	}
}

// applyEnvOverrides reads environment variables and overrides config values.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("GRIMORIO_TTS_ENABLED"); v != "" {
		cfg.TTS.Enabled = v == "true" || v == "1"
	}
	if v := os.Getenv("PIPER_MODEL_PATH"); v != "" {
		cfg.TTS.Piper.ModelPath = v
	}
	if v := os.Getenv("PIPER_CONFIG_PATH"); v != "" {
		cfg.TTS.Piper.ConfigPath = v
	}
	if v := os.Getenv("PIPER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.TTS.Piper.Port = port
		}
	}
	if v := os.Getenv("PIPER_HOST"); v != "" {
		cfg.TTS.Piper.Host = v
	}
	if v := os.Getenv("PIPER_LENGTH_SCALE"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.TTS.Piper.LengthScale = f
		}
	}
	if v := os.Getenv("PIPER_VOLUME"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.TTS.Piper.Volume = f
		}
	}
	if v := os.Getenv("PIPER_CACHE_DIR"); v != "" {
		cfg.TTS.Piper.CacheDir = v
	}
	if v := os.Getenv("PIPER_MAX_RESTARTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.TTS.Piper.MaxRestarts = n
		}
	}
	if v := os.Getenv("CHUNKER_MAX_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.TTS.Chunker.MaxChunkSize = n
		}
	}
	if v := os.Getenv("AUDIO_PLAYER"); v != "" {
		cfg.TTS.Audio.Player = v
	}
	if v := os.Getenv("AUDIO_DEVICE"); v != "" {
		cfg.TTS.Audio.Device = v
	}
	if v := os.Getenv("AUDIO_PRELOAD_BUFFER"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.TTS.Audio.PreloadBuffer = n
		}
	}
}

func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
