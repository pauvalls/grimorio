// Package piper provides the local Piper TTS engine integration.
package piper

import "time"

// Config parametrizes the local Piper TTS engine.
type Config struct {
	// ModelPath is the absolute path to the Piper ONNX model.
	ModelPath string `json:"model_path"`

	// ConfigPath is the absolute path to the model JSON config.
	// If empty, the system attempts to auto-detect <model>.onnx.json.
	ConfigPath string `json:"config_path"`

	// Port is the HTTP port for the Piper local server.
	Port int `json:"port"`

	// Host is the bind address for the Piper local server.
	Host string `json:"host"`

	// LengthScale controls speech speed (<1 faster, >1 slower).
	LengthScale float64 `json:"length_scale"`

	// Volume controls playback volume (0.0–1.0).
	Volume float64 `json:"volume"`

	// CacheDir is an optional directory for caching generated WAV files.
	CacheDir string `json:"cache_dir"`

	// MaxRestarts is the maximum number of auto-restart attempts.
	MaxRestarts int `json:"max_restarts"`

	// HealthcheckTimeout is the timeout for Piper health checks.
	HealthcheckTimeout time.Duration `json:"healthcheck_timeout"`

	// Player is the preferred audio player: "auto", "aplay", "paplay", "ffplay".
	Player string `json:"player"`

	// Device is an optional audio device string (e.g. ALSA plughw).
	Device string `json:"device"`

	// PreloadBuffer is how many chunks ahead to synthesize.
	PreloadBuffer int `json:"preload_buffer"`

	// MaxChunkSize is the maximum characters per TTS chunk.
	MaxChunkSize int `json:"max_chunk_size"`
}

// DefaultConfig returns a Config with sensible defaults for Piper TTS.
func DefaultConfig() Config {
	return Config{
		ModelPath:          "",
		ConfigPath:         "",
		Port:               5000,
		Host:               "127.0.0.1",
		LengthScale:        1.0,
		Volume:             0.8,
		CacheDir:           "",
		MaxRestarts:        3,
		HealthcheckTimeout: 2 * time.Second,
		Player:             "auto",
		Device:             "",
		PreloadBuffer:      1,
		MaxChunkSize:       150,
	}
}
