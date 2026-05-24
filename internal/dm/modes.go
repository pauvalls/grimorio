// Package dm provides Dungeon Master mode management, text chunking,
// voice registry, and TTS client interfaces for the Grimorio campaign system.
package dm

// DMMode represents the output mode for the Dungeon Master.
type DMMode string

const (
	// ModeWritten outputs DM responses as pure text only.
	ModeWritten DMMode = "written"
	// ModeTTS outputs DM responses as synthesized speech via TTS.
	ModeTTS DMMode = "tts"
)

// TTSConfig holds configuration for the TTS client and server connection.
type TTSConfig struct {
	ServerURL     string `json:"server_url"`
	Enabled       bool   `json:"enabled"`
	PreloadNext   bool   `json:"preload_next"`
	ShowSubtitles bool   `json:"show_subtitles"`
}

// DefaultTTSConfig returns a TTSConfig with sensible defaults.
func DefaultTTSConfig() *TTSConfig {
	return &TTSConfig{
		ServerURL:     "ws://localhost:8765/tts",
		Enabled:       false,
		PreloadNext:   true,
		ShowSubtitles: true,
	}
}
