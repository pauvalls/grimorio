# TTS (Text-to-Speech) Architecture Plan for Grimorio-DM

## Executive Summary

This document proposes a generic, provider-pattern TTS architecture for Grimorio-DM that integrates seamlessly with the existing codebase conventions. The design follows the same patterns established by the image generation subsystem (`internal/image/`), ensuring consistency and maintainability.

**Budget Target**: €20/month  
**Usage Pattern**: 5-10 sessions/month, ~3 hours each  
**Initial Provider**: ElevenLabs  
**Future Providers**: Google Cloud TTS, Amazon Polly, Azure TTS, Coqui TTS (self-hosted)

---

## 1. Current Codebase Patterns Analysis

### 1.1 Provider Pattern (Image Subsystem)
The existing image generation subsystem (`internal/image/`) provides the perfect template:

```
internal/image/
├── provider.go          # Provider interface + factory
├── dalle.go             # DALL-E provider implementation
├── pollinations.go      # Pollinations provider (free)
├── raphael.go           # Raphael provider (free)
└── provider_test.go     # Provider tests
```

**Key Patterns**:
- `Provider` interface with `Generate()`, `IsConfigured()`, `Name()` methods
- `Config` struct with provider selection and provider-specific settings
- `NewProvider(cfg Config)` factory function with switch statement
- `NewProviderChain(cfg Config)` for fallback chains
- Provider injected into `AssetService` via constructor

### 1.2 Configuration Pattern
Config is centralized in `internal/config/config.go`:
- JSON config file at `~/.config/grimorio/config.json`
- Embedded sub-configs (e.g., `image.Config`)
- Environment variable fallbacks (e.g., `OPENAI_API_KEY`)
- Default values in `DefaultConfig()`

### 1.3 Service Layer Pattern
Services in `internal/services/`:
- Constructor accepts repositories and config
- Business logic orchestration
- Rate limiting (per-campaign limiters in `AssetService`)
- Error wrapping with context

### 1.4 MCP Handler Pattern
Handlers in `internal/mcp/handlers/`:
- Struct wrapping a service pointer
- Constructor `NewXxxHandlers(service *services.XxxService)`
- Methods return `server.ToolHandlerFunc`
- Tool registration in `internal/mcp/server.go`

### 1.5 Domain Types
`NPCContext` already has a `DialogueVoice string` field (line 151 of `dm_context.go`), providing the perfect hook for voice ID assignment.

---

## 2. Proposed Architecture

### 2.1 Directory Structure

```
internal/
├── tts/
│   ├── provider.go              # TTSProvider interface + factory
│   ├── elevenlabs.go            # ElevenLabs implementation
│   ├── config.go                # TTSConfig struct
│   ├── cache.go                 # Disk-based audio cache
│   └── provider_test.go         # Interface tests
├── config/
│   └── config.go                # Add TTSConfig embedding
├── services/
│   └── tts_service.go           # Business logic + caching
├── mcp/handlers/
│   └── tts.go                   # MCP tool handlers
└── domain/
    └── tts.go                   # TTS domain types
```

### 2.2 Domain Types (`internal/domain/tts.go`)

```go
package domain

import "time"

// TTSRequest represents a request to synthesize speech
type TTSRequest struct {
    Text      string            `json:"text"`
    VoiceID   string            `json:"voice_id"`
    ModelID   string            `json:"model_id,omitempty"`
    Settings  TTSVoiceSettings  `json:"settings,omitempty"`
    Metadata  map[string]string `json:"metadata,omitempty"` // npc_id, campaign_id, etc.
}

// TTSVoiceSettings controls voice synthesis parameters
type TTSVoiceSettings struct {
    Stability       float64 `json:"stability"`        // 0.0 - 1.0
    SimilarityBoost float64 `json:"similarity_boost"` // 0.0 - 1.0
    Style           float64 `json:"style,omitempty"`  // 0.0 - 1.0 (ElevenLabs only)
    SpeakerBoost    bool    `json:"speaker_boost,omitempty"`
    Speed           float64 `json:"speed,omitempty"`  // 0.5 - 2.0 (generic)
}

// TTSResponse represents the result of a synthesis
type TTSResponse struct {
    AudioData    []byte    `json:"-"`               // Raw audio bytes
    AudioPath    string    `json:"audio_path"`      // Path to cached file
    ContentType  string    `json:"content_type"`    // audio/mpeg, audio/wav, etc.
    Duration     float64   `json:"duration_secs"`   // Estimated duration
    Provider     string    `json:"provider"`        // Provider name
    VoiceID      string    `json:"voice_id"`        // Voice used
    CharacterCount int     `json:"character_count"` // Text length
    Cached       bool      `json:"cached"`          // Was served from cache?
    GeneratedAt  time.Time `json:"generated_at"`
}

// TTSVoiceInfo represents available voice metadata
type TTSVoiceInfo struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Provider    string   `json:"provider"`
    Gender      string   `json:"gender,omitempty"`
    Age         string   `json:"age,omitempty"`
    Accent      string   `json:"accent,omitempty"`
    Description string   `json:"description,omitempty"`
    PreviewURL  string   `json:"preview_url,omitempty"`
    Language    string   `json:"language"` // ISO 639-1
    // Provider-specific fields
    Labels      map[string]string `json:"labels,omitempty"`
}

// NPCVoiceMapping links an NPC to a TTS voice
type NPCVoiceMapping struct {
    CampaignID string `json:"campaign_id"`
    NPCID      string `json:"npc_id"`
    NPCName    string `json:"npc_name"`
    VoiceID    string `json:"voice_id"`
    VoiceName  string `json:"voice_name,omitempty"`
    // Optional per-NPC voice tuning
    Settings   TTSVoiceSettings `json:"settings,omitempty"`
}
```

### 2.3 TTS Provider Interface (`internal/tts/provider.go`)

```go
package tts

import (
    "context"
    "fmt"
    "github.com/pauvalls/grimorio/internal/domain"
)

// Provider defines the generic TTS interface
type Provider interface {
    // Synthesize converts text to speech audio
    Synthesize(ctx context.Context, req domain.TTSRequest) (*domain.TTSResponse, error)
    
    // ListVoices returns available voices for this provider
    ListVoices(ctx context.Context, language string) ([]domain.TTSVoiceInfo, error)
    
    // GetVoice returns details for a specific voice
    GetVoice(ctx context.Context, voiceID string) (*domain.TTSVoiceInfo, error)
    
    // IsConfigured returns true if the provider has valid credentials
    IsConfigured() bool
    
    // Name returns the provider identifier
    Name() string
    
    // SupportsStreaming returns true if the provider supports streaming audio
    SupportsStreaming() bool
    
    // MaxCharacters returns the maximum text length per request
    MaxCharacters() int
}

// Config holds TTS configuration
type Config struct {
    Provider        string  `json:"tts_provider"`
    ElevenLabsKey   string  `json:"elevenlabs_api_key,omitempty"`
    ElevenLabsModel string  `json:"elevenlabs_model,omitempty"`
    CacheDir        string  `json:"tts_cache_dir,omitempty"`
    CacheEnabled    bool    `json:"tts_cache_enabled"`
    DefaultVoiceID  string  `json:"tts_default_voice_id,omitempty"`
    DefaultLanguage string  `json:"tts_default_language,omitempty"`
    MaxCacheSizeMB  int     `json:"tts_max_cache_size_mb,omitempty"`
}

// DefaultConfig returns default TTS configuration
func DefaultConfig() Config {
    return Config{
        Provider:        "elevenlabs",
        ElevenLabsModel: "eleven_multilingual_v2",
        CacheEnabled:    true,
        DefaultLanguage: "es", // Default to Spanish for Rioplatense
        MaxCacheSizeMB:  500,  // 500MB default cache
    }
}

// NewProvider creates a TTS provider by name
func NewProvider(cfg Config) (Provider, error) {
    switch cfg.Provider {
    case "elevenlabs":
        return NewElevenLabsProvider(cfg.ElevenLabsKey, cfg.ElevenLabsModel)
    case "google":
        return nil, fmt.Errorf("Google TTS provider not yet implemented")
    case "amazon":
        return nil, fmt.Errorf("Amazon Polly provider not yet implemented")
    case "azure":
        return nil, fmt.Errorf("Azure TTS provider not yet implemented")
    case "coqui":
        return nil, fmt.Errorf("Coqui TTS provider not yet implemented")
    default:
        return nil, fmt.Errorf("unknown TTS provider: %s", cfg.Provider)
    }
}
```

### 2.4 ElevenLabs Provider (`internal/tts/elevenlabs.go`)

```go
package tts

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "time"
    
    "github.com/pauvalls/grimorio/internal/domain"
)

const (
    elevenLabsBaseURL = "https://api.elevenlabs.io/v1"
    elevenLabsTimeout = 30 * time.Second
    maxTextLength     = 5000 // ElevenLabs limit
)

// ElevenLabsProvider implements the TTS Provider interface
type ElevenLabsProvider struct {
    apiKey string
    model  string
    client *http.Client
}

// NewElevenLabsProvider creates a new ElevenLabs provider
func NewElevenLabsProvider(apiKey, model string) (*ElevenLabsProvider, error) {
    if apiKey == "" {
        apiKey = os.Getenv("ELEVENLABS_API_KEY")
    }
    if apiKey == "" {
        return nil, fmt.Errorf("ELEVENLABS_API_KEY not set")
    }
    if model == "" {
        model = "eleven_multilingual_v2"
    }
    return &ElevenLabsProvider{
        apiKey: apiKey,
        model:  model,
        client: &http.Client{Timeout: elevenLabsTimeout},
    }, nil
}

func (e *ElevenLabsProvider) Name() string { return "elevenlabs" }

func (e *ElevenLabsProvider) IsConfigured() bool { return e.apiKey != "" }

func (e *ElevenLabsProvider) SupportsStreaming() bool { return true }

func (e *ElevenLabsProvider) MaxCharacters() int { return maxTextLength }

// elevenLabsRequest matches the ElevenLabs API schema
type elevenLabsRequest struct {
    Text          string                 `json:"text"`
    ModelID       string                 `json:"model_id"`
    VoiceSettings map[string]interface{} `json:"voice_settings,omitempty"`
}

// elevenLabsVoice matches the ElevenLabs voice schema
type elevenLabsVoice struct {
    VoiceID     string            `json:"voice_id"`
    Name        string            `json:"name"`
    Category    string            `json:"category"`
    Labels      map[string]string `json:"labels"`
    PreviewURL  string            `json:"preview_url"`
}

func (e *ElevenLabsProvider) Synthesize(ctx context.Context, req domain.TTSRequest) (*domain.TTSResponse, error) {
    if len(req.Text) > maxTextLength {
        return nil, fmt.Errorf("text exceeds maximum length of %d characters", maxTextLength)
    }
    
    apiReq := elevenLabsRequest{
        Text:    req.Text,
        ModelID: e.model,
    }
    
    // Map generic settings to ElevenLabs-specific settings
    if req.Settings.Stability > 0 || req.Settings.SimilarityBoost > 0 {
        apiReq.VoiceSettings = make(map[string]interface{})
        if req.Settings.Stability > 0 {
            apiReq.VoiceSettings["stability"] = req.Settings.Stability
        }
        if req.Settings.SimilarityBoost > 0 {
            apiReq.VoiceSettings["similarity_boost"] = req.Settings.SimilarityBoost
        }
        if req.Settings.Style > 0 {
            apiReq.VoiceSettings["style"] = req.Settings.Style
        }
        if req.Settings.SpeakerBoost {
            apiReq.VoiceSettings["use_speaker_boost"] = true
        }
    }
    
    jsonData, err := json.Marshal(apiReq)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }
    
    url := fmt.Sprintf("%s/text-to-speech/%s", elevenLabsBaseURL, req.VoiceID)
    if e.model != "" {
        url += fmt.Sprintf("?model_id=%s", e.model)
    }
    
    httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("xi-api-key", e.apiKey)
    
    resp, err := e.client.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("ElevenLabs API error %d: %s", resp.StatusCode, string(body))
    }
    
    audioData, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read audio data: %w", err)
    }
    
    // Estimate duration: ~150 chars per second for Spanish at normal speed
    duration := float64(len(req.Text)) / 150.0
    
    return &domain.TTSResponse{
        AudioData:      audioData,
        ContentType:    "audio/mpeg",
        Duration:       duration,
        Provider:       e.Name(),
        VoiceID:        req.VoiceID,
        CharacterCount: len(req.Text),
        GeneratedAt:    time.Now(),
    }, nil
}

func (e *ElevenLabsProvider) ListVoices(ctx context.Context, language string) ([]domain.TTSVoiceInfo, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", 
        fmt.Sprintf("%s/voices", elevenLabsBaseURL), nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("xi-api-key", e.apiKey)
    
    resp, err := e.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to list voices: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("ElevenLabs API error %d: %s", resp.StatusCode, string(body))
    }
    
    var result struct {
        Voices []elevenLabsVoice `json:"voices"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode voices: %w", err)
    }
    
    var voices []domain.TTSVoiceInfo
    for _, v := range result.Voices {
        voice := domain.TTSVoiceInfo{
            ID:         v.VoiceID,
            Name:       v.Name,
            Provider:   e.Name(),
            PreviewURL: v.PreviewURL,
            Language:   "multilingual", // ElevenLabs v2 is multilingual
            Labels:     v.Labels,
        }
        if gender, ok := v.Labels["gender"]; ok {
            voice.Gender = gender
        }
        if age, ok := v.Labels["age"]; ok {
            voice.Age = age
        }
        if accent, ok := v.Labels["accent"]; ok {
            voice.Accent = accent
        }
        if language, ok := v.Labels["language"]; ok {
            voice.Language = language
        }
        voices = append(voices, voice)
    }
    
    return voices, nil
}

func (e *ElevenLabsProvider) GetVoice(ctx context.Context, voiceID string) (*domain.TTSVoiceInfo, error) {
    voices, err := e.ListVoices(ctx, "")
    if err != nil {
        return nil, err
    }
    for _, v := range voices {
        if v.ID == voiceID {
            return &v, nil
        }
    }
    return nil, fmt.Errorf("voice not found: %s", voiceID)
}
```

### 2.5 TTS Service (`internal/services/tts_service.go`)

```go
package services

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "time"
    
    "github.com/pauvalls/grimorio/internal/cache"
    "github.com/pauvalls/grimorio/internal/domain"
    "github.com/pauvalls/grimorio/internal/tts"
)

// TTSService handles text-to-speech generation with caching
type TTSService struct {
    provider     tts.Provider
    cacheDir     string
    cacheEnabled bool
    maxCacheSize int64 // bytes
    
    // In-memory cache for hot audio data (small responses)
    memCache *cache.LRUCache[string, []byte]
    
    // Per-campaign voice mappings: campaignID -> npcID -> NPCVoiceMapping
    voiceMappings map[string]map[string]domain.NPCVoiceMapping
    mappingsMu    sync.RWMutex
}

// NewTTSService creates a new TTS service
func NewTTSService(provider tts.Provider, cacheDir string, cacheEnabled bool, maxCacheSizeMB int) *TTSService {
    if maxCacheSizeMB <= 0 {
        maxCacheSizeMB = 500
    }
    return &TTSService{
        provider:      provider,
        cacheDir:      cacheDir,
        cacheEnabled:  cacheEnabled,
        maxCacheSize:  int64(maxCacheSizeMB) * 1024 * 1024,
        memCache:      cache.NewLRU[string, []byte](100), // Cache 100 recent audio files in memory
        voiceMappings: make(map[string]map[string]domain.NPCVoiceMapping),
    }
}

// Synthesize generates speech for text, using cache if available
func (s *TTSService) Synthesize(ctx context.Context, req domain.TTSRequest) (*domain.TTSResponse, error) {
    if s.provider == nil || !s.provider.IsConfigured() {
        return nil, fmt.Errorf("TTS provider not configured")
    }
    
    // Generate cache key from text + voice + settings
    cacheKey := s.generateCacheKey(req)
    cachePath := filepath.Join(s.cacheDir, cacheKey+".mp3")
    
    // Check in-memory cache first (fastest)
    if s.cacheEnabled {
        if audioData, ok := s.memCache.Get(cacheKey); ok {
            return &domain.TTSResponse{
                AudioData:      audioData,
                AudioPath:      cachePath,
                ContentType:    "audio/mpeg",
                Provider:       s.provider.Name(),
                VoiceID:        req.VoiceID,
                CharacterCount: len(req.Text),
                Cached:         true,
                GeneratedAt:    time.Now(),
            }, nil
        }
        
        // Check disk cache
        if audioData, err := os.ReadFile(cachePath); err == nil {
            // Populate in-memory cache
            s.memCache.Put(cacheKey, audioData)
            return &domain.TTSResponse{
                AudioData:      audioData,
                AudioPath:      cachePath,
                ContentType:    "audio/mpeg",
                Provider:       s.provider.Name(),
                VoiceID:        req.VoiceID,
                CharacterCount: len(req.Text),
                Cached:         true,
                GeneratedAt:    time.Now(),
            }, nil
        }
    }
    
    // Generate fresh audio
    resp, err := s.provider.Synthesize(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("synthesis failed: %w", err)
    }
    
    // Cache the result
    if s.cacheEnabled && resp.AudioData != nil {
        // Ensure cache directory exists
        if err := os.MkdirAll(s.cacheDir, 0755); err == nil {
            // Write to disk cache
            if err := os.WriteFile(cachePath, resp.AudioData, 0644); err == nil {
                resp.AudioPath = cachePath
                // Also cache in memory
                s.memCache.Put(cacheKey, resp.AudioData)
                
                // Enforce max cache size (simple LRU eviction)
                s.enforceCacheSize()
            }
        }
    }
    
    return resp, nil
}

// SynthesizeNPC generates speech for an NPC's dialogue
func (s *TTSService) SynthesizeNPC(ctx context.Context, campaignID, npcID, npcName, text string) (*domain.TTSResponse, error) {
    // Get voice mapping for this NPC
    mapping := s.GetVoiceMapping(campaignID, npcID)
    if mapping == nil {
        // No mapping exists - this is an error condition that should be handled
        // by the caller assigning a voice first
        return nil, fmt.Errorf("no voice assigned to NPC %s in campaign %s", npcName, campaignID)
    }
    
    req := domain.TTSRequest{
        Text:     text,
        VoiceID:  mapping.VoiceID,
        Settings: mapping.Settings,
        Metadata: map[string]string{
            "campaign_id": campaignID,
            "npc_id":      npcID,
            "npc_name":    npcName,
        },
    }
    
    return s.Synthesize(ctx, req)
}

// ListAvailableVoices returns all available voices from the provider
func (s *TTSService) ListAvailableVoices(ctx context.Context, language string) ([]domain.TTSVoiceInfo, error) {
    if s.provider == nil {
        return nil, fmt.Errorf("TTS provider not configured")
    }
    return s.provider.ListVoices(ctx, language)
}

// AssignVoiceToNPC assigns a voice to an NPC
func (s *TTSService) AssignVoiceToNPC(campaignID string, mapping domain.NPCVoiceMapping) error {
    s.mappingsMu.Lock()
    defer s.mappingsMu.Unlock()
    
    if s.voiceMappings[campaignID] == nil {
        s.voiceMappings[campaignID] = make(map[string]domain.NPCVoiceMapping)
    }
    s.voiceMappings[campaignID][mapping.NPCID] = mapping
    
    // Persist to disk (campaign-specific voice mapping file)
    return s.saveVoiceMappings(campaignID)
}

// GetVoiceMapping retrieves voice mapping for an NPC
func (s *TTSService) GetVoiceMapping(campaignID, npcID string) *domain.NPCVoiceMapping {
    s.mappingsMu.RLock()
    defer s.mappingsMu.RUnlock()
    
    if campaignMap, ok := s.voiceMappings[campaignID]; ok {
        if mapping, ok := campaignMap[npcID]; ok {
            return &mapping
        }
    }
    return nil
}

// GetNPCVoices returns all voice mappings for a campaign
func (s *TTSService) GetNPCVoices(campaignID string) []domain.NPCVoiceMapping {
    s.mappingsMu.RLock()
    defer s.mappingsMu.RUnlock()
    
    var mappings []domain.NPCVoiceMapping
    if campaignMap, ok := s.voiceMappings[campaignID]; ok {
        for _, m := range campaignMap {
            mappings = append(mappings, m)
        }
    }
    return mappings
}

// generateCacheKey creates a deterministic cache key from request
func (s *TTSService) generateCacheKey(req domain.TTSRequest) string {
    // Hash of: text + voice_id + model + stability + similarity_boost
    h := sha256.New()
    h.Write([]byte(req.Text))
    h.Write([]byte(req.VoiceID))
    h.Write([]byte(req.ModelID))
    fmt.Fprintf(h, "%.2f_%.2f", req.Settings.Stability, req.Settings.SimilarityBoost)
    return hex.EncodeToString(h.Sum(nil))[:32] // First 32 chars of hex
}

// enforceCacheSize removes oldest files if cache exceeds max size
func (s *TTSService) enforceCacheSize() {
    // Implementation: list cache dir, sort by mtime, remove oldest until under limit
    // (omitted for brevity - standard file system operations)
}

// saveVoiceMappings persists voice mappings to disk
func (s *TTSService) saveVoiceMappings(campaignID string) error {
    // Implementation: save as JSON in campaign dir
    // Path: <baseDir>/<campaignID>/tts/voice-mappings.json
    return nil
}

// loadVoiceMappings loads voice mappings from disk
func (s *TTSService) loadVoiceMappings(campaignID string) error {
    // Implementation: load from JSON file
    return nil
}
```

### 2.6 MCP Handlers (`internal/mcp/handlers/tts.go`)

```go
package handlers

import (
    "context"
    "fmt"
    
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
    "github.com/pauvalls/grimorio/internal/domain"
    "github.com/pauvalls/grimorio/internal/services"
)

// TTSHandlers handles TTS-related MCP tools
type TTSHandlers struct {
    service *services.TTSService
}

// NewTTSHandlers creates new TTS handlers
func NewTTSHandlers(service *services.TTSService) *TTSHandlers {
    return &TTSHandlers{service: service}
}

// HandleListVoices returns available TTS voices
func (h *TTSHandlers) HandleListVoices() server.ToolHandlerFunc {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        args, ok := request.Params.Arguments.(map[string]any)
        if !ok {
            return mcp.NewToolResultError("invalid arguments"), nil
        }
        
        language := getStringArg(args, "language")
        if language == "" {
            language = "es" // Default to Spanish
        }
        
        voices, err := h.service.ListAvailableVoices(ctx, language)
        if err != nil {
            return mcp.NewToolResultError(err.Error()), nil
        }
        
        // Format as readable text
        var result string
        result = fmt.Sprintf("Available TTS voices (%d found):\n\n", len(voices))
        for _, v := range voices {
            result += fmt.Sprintf("- %s (ID: %s)", v.Name, v.ID)
            if v.Gender != "" {
                result += fmt.Sprintf(", %s", v.Gender)
            }
            if v.Accent != "" {
                result += fmt.Sprintf(", %s accent", v.Accent)
            }
            if v.Description != "" {
                result += fmt.Sprintf(" - %s", v.Description)
            }
            result += "\n"
        }
        
        return mcp.NewToolResultText(result), nil
    }
}

// HandleSynthesize generates speech from text
func (h *TTSHandlers) HandleSynthesize() server.ToolHandlerFunc {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        args, ok := request.Params.Arguments.(map[string]any)
        if !ok {
            return mcp.NewToolResultError("invalid arguments"), nil
        }
        
        text := getStringArg(args, "text")
        voiceID := getStringArg(args, "voice_id")
        
        if text == "" {
            return mcp.NewToolResultError("text is required"), nil
        }
        if voiceID == "" {
            return mcp.NewToolResultError("voice_id is required"), nil
        }
        
        req := domain.TTSRequest{
            Text:    text,
            VoiceID: voiceID,
        }
        
        resp, err := h.service.Synthesize(ctx, req)
        if err != nil {
            return mcp.NewToolResultError(err.Error()), nil
        }
        
        var result string
        if resp.Cached {
            result = fmt.Sprintf("Audio served from cache: %s\n", resp.AudioPath)
        } else {
            result = fmt.Sprintf("Audio generated: %s\n", resp.AudioPath)
        }
        result += fmt.Sprintf("Duration: %.1fs | Characters: %d | Provider: %s | Voice: %s",
            resp.Duration, resp.CharacterCount, resp.Provider, resp.VoiceID)
        
        return mcp.NewToolResultText(result), nil
    }
}

// HandleAssignNPCVoice assigns a voice to an NPC
func (h *TTSHandlers) HandleAssignNPCVoice() server.ToolHandlerFunc {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        args, ok := request.Params.Arguments.(map[string]any)
        if !ok {
            return mcp.NewToolResultError("invalid arguments"), nil
        }
        
        campaign := getStringArg(args, "campaign")
        npcID := getStringArg(args, "npc_id")
        npcName := getStringArg(args, "npc_name")
        voiceID := getStringArg(args, "voice_id")
        
        if campaign == "" || npcID == "" || voiceID == "" {
            return mcp.NewToolResultError("campaign, npc_id, and voice_id are required"), nil
        }
        
        mapping := domain.NPCVoiceMapping{
            CampaignID: campaign,
            NPCID:      npcID,
            NPCName:    npcName,
            VoiceID:    voiceID,
        }
        
        if err := h.service.AssignVoiceToNPC(campaign, mapping); err != nil {
            return mcp.NewToolResultError(err.Error()), nil
        }
        
        return mcp.NewToolResultText(
            fmt.Sprintf("Voice %s assigned to NPC %s in campaign %s", voiceID, npcName, campaign)),
        nil
    }
}

// HandleSynthesizeNPC generates speech for an NPC
func (h *TTSHandlers) HandleSynthesizeNPC() server.ToolHandlerFunc {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        args, ok := request.Params.Arguments.(map[string]any)
        if !ok {
            return mcp.NewToolResultError("invalid arguments"), nil
        }
        
        campaign := getStringArg(args, "campaign")
        npcID := getStringArg(args, "npc_id")
        npcName := getStringArg(args, "npc_name")
        text := getStringArg(args, "text")
        
        if campaign == "" || npcID == "" || text == "" {
            return mcp.NewToolResultError("campaign, npc_id, and text are required"), nil
        }
        
        resp, err := h.service.SynthesizeNPC(ctx, campaign, npcID, npcName, text)
        if err != nil {
            return mcp.NewToolResultError(err.Error()), nil
        }
        
        var result string
        if resp.Cached {
            result = fmt.Sprintf("NPC audio served from cache: %s\n", resp.AudioPath)
        } else {
            result = fmt.Sprintf("NPC audio generated: %s\n", resp.AudioPath)
        }
        result += fmt.Sprintf("NPC: %s | Duration: %.1fs | Characters: %d", 
            npcName, resp.Duration, resp.CharacterCount)
        
        return mcp.NewToolResultText(result), nil
    }
}

// HandleListNPCVoices lists voice assignments for a campaign
func (h *TTSHandlers) HandleListNPCVoices() server.ToolHandlerFunc {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        args, ok := request.Params.Arguments.(map[string]any)
        if !ok {
            return mcp.NewToolResultError("invalid arguments"), nil
        }
        
        campaign := getStringArg(args, "campaign")
        if campaign == "" {
            return mcp.NewToolResultError("campaign is required"), nil
        }
        
        mappings := h.service.GetNPCVoices(campaign)
        if len(mappings) == 0 {
            return mcp.NewToolResultText("No voice assignments found for this campaign"), nil
        }
        
        var result string
        result = fmt.Sprintf("NPC Voice Assignments for %s:\n\n", campaign)
        for _, m := range mappings {
            result += fmt.Sprintf("- %s (ID: %s): Voice %s\n", m.NPCName, m.NPCID, m.VoiceID)
        }
        
        return mcp.NewToolResultText(result), nil
    }
}
```

### 2.7 MCP Tool Registration (Additions to `internal/mcp/server.go`)

```go
// After existing service initialization:

// TTS service initialization
ttsProvider, err := tts.NewProvider(cfg.TTSConfig)
if err != nil {
    log.Printf("WARNING: TTS provider initialization failed: %v. TTS tools will be unavailable.", err)
}
ttsCacheDir := filepath.Join(cfg.OutputDir, ".tts-cache")
ttsService := services.NewTTSService(ttsProvider, ttsCacheDir, cfg.TTSConfig.CacheEnabled, cfg.TTSConfig.MaxCacheSizeMB)
ttsHandlers := handlers.NewTTSHandlers(ttsService)

// In tool registration section:

// TTS tools
s.AddTool(mcp.NewTool("list_tts_voices",
    mcp.WithDescription("List available TTS voices from the configured provider"),
    mcp.WithString("language", mcp.Description("Language filter (e.g., 'es', 'en')")),
), ttsHandlers.HandleListVoices())

s.AddTool(mcp.NewTool("synthesize_speech",
    mcp.WithDescription("Convert text to speech audio"),
    mcp.WithString("text", mcp.Required(), mcp.Description("Text to synthesize")),
    mcp.WithString("voice_id", mcp.Required(), mcp.Description("Voice ID from list_tts_voices")),
), ttsHandlers.HandleSynthesize())

s.AddTool(mcp.NewTool("assign_npc_voice",
    mcp.WithDescription("Assign a TTS voice to an NPC for consistent dialogue"),
    mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name")),
    mcp.WithString("npc_id", mcp.Required(), mcp.Description("NPC identifier")),
    mcp.WithString("npc_name", mcp.Description("NPC display name")),
    mcp.WithString("voice_id", mcp.Required(), mcp.Description("Voice ID from list_tts_voices")),
), ttsHandlers.HandleAssignNPCVoice())

s.AddTool(mcp.NewTool("synthesize_npc_dialogue",
    mcp.WithDescription("Generate speech for an NPC using their assigned voice"),
    mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name")),
    mcp.WithString("npc_id", mcp.Required(), mcp.Description("NPC identifier")),
    mcp.WithString("npc_name", mcp.Description("NPC display name")),
    mcp.WithString("text", mcp.Required(), mcp.Description("Dialogue text")),
), ttsHandlers.HandleSynthesizeNPC())

s.AddTool(mcp.NewTool("list_npc_voices",
    mcp.WithDescription("List all NPC voice assignments in a campaign"),
    mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name")),
), ttsHandlers.HandleListNPCVoices())
```

### 2.8 Configuration Integration

**Updates to `internal/config/config.go`**:

```go
type Config struct {
    OutputDir       string `json:"output_dir"`
    PDFEngine       string `json:"pdf_engine"`
    CompilerVersion int    `json:"compiler_version"`
    image.Config
    tts.Config      // Embedded TTS config
}

func DefaultConfig() *Config {
    home, _ := os.UserHomeDir()
    return &Config{
        OutputDir:       filepath.Join(home, "campaigns"),
        PDFEngine:       "wkhtmltopdf",
        CompilerVersion: 2,
        Config:          image.DefaultConfig(),
        TTSConfig:       tts.DefaultConfig(), // Add this
    }
}
```

**Config file example** (`~/.config/grimorio/config.json`):

```json
{
  "output_dir": "/home/user/campaigns",
  "pdf_engine": "wkhtmltopdf",
  "compiler_version": 2,
  "image_provider": "pollinations",
  "image_width": 1024,
  "image_height": 1024,
  "tts_provider": "elevenlabs",
  "elevenlabs_api_key": "sk_xxxxxxxxxxxxxxxx",
  "elevenlabs_model": "eleven_multilingual_v2",
  "tts_cache_enabled": true,
  "tts_cache_dir": "/home/user/.cache/grimorio/tts",
  "tts_default_voice_id": "21m00Tcm4TlvDq8ikWAM",
  "tts_default_language": "es",
  "tts_max_cache_size_mb": 500
}
```

---

## 3. Caching Strategy

### 3.1 Two-Tier Cache

```
┌─────────────────────────────────────────────────────────────┐
│                     TTS Cache Architecture                   │
├─────────────────────────────────────────────────────────────┤
│  Tier 1: In-Memory (LRU)                                     │
│  - Capacity: 100 entries                                     │
│  - Content: Raw audio bytes ([]byte)                        │
│  - Key: SHA256(text+voice_id+settings)[:32]                │
│  - Lifetime: Process lifetime                               │
│  - Use case: Repeated NPC dialogue during session           │
├─────────────────────────────────────────────────────────────┤
│  Tier 2: Disk Cache                                          │
│  - Location: ~/.cache/grimorio/tts/ or <outputDir>/.tts-cache│
│  - Capacity: Configurable (default 500MB)                   │
│  - Content: MP3 files                                        │
│  - Key: Same SHA256 as memory cache                         │
│  - Lifetime: Persistent across sessions                     │
│  - Use case: Campaign-specific audio (prologue, lore)       │
├─────────────────────────────────────────────────────────────┤
│  Tier 3: Provider API                                        │
│  - Cost: € per character                                    │
│  - Only hit when cache misses                               │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 Cache Key Generation

```go
func generateCacheKey(req domain.TTSRequest) string {
    h := sha256.New()
    h.Write([]byte(req.Text))
    h.Write([]byte(req.VoiceID))
    h.Write([]byte(req.ModelID))
    fmt.Fprintf(h, "%.2f_%.2f", req.Settings.Stability, req.Settings.SimilarityBoost)
    return hex.EncodeToString(h.Sum(nil))[:32]
}
```

### 3.3 Cache Eviction Policy

- **Memory cache**: Pure LRU (handled by `cache.LRUCache`)
- **Disk cache**: Size-based eviction with mtime sorting
  - Check total size on each write
  - If over limit: delete oldest files (by modification time) until under 80% of limit
  - Files are small (~50-200KB each for typical NPC dialogue)

### 3.4 Cache Hit Scenarios

**High hit rate scenarios**:
1. NPC repeated greetings/phrases ("Bienvenidos a mi taberna")
2. Prologue narration (read once per session)
3. Recurring location descriptions
4. Combat barks ("¡Por la gloria!")
5. Common item/spell descriptions

---

## 4. Cost Analysis: ElevenLabs

### 4.1 Pricing Model

ElevenLabs pricing (as of 2024):
- **Starter Plan**: $5/month (€4.60) - 30,000 characters/month
- **Creator Plan**: $22/month (€20.24) - 100,000 characters/month
- **Pro Plan**: $99/month - 500,000 characters/month

Pay-as-you-go overages:
- Starter: $0.18 per 1,000 characters
- Creator: $0.10 per 1,000 characters

### 4.2 Usage Estimates

**Assumptions**:
- 5-10 sessions per month
- ~3 hours per session
- Spanish language (~150 chars/second of audio at normal speed)

**Per-session breakdown**:

| Content Type | Chars/Session | Sessions | Monthly Chars | Notes |
|-------------|---------------|----------|---------------|-------|
| Prologue narration | 2,000 | 5-10 | 10,000-20,000 | Read once/session |
| NPC dialogue | 5,000 | 5-10 | 25,000-50,000 | ~30 min of dialogue |
| Combat barks | 1,500 | 5-10 | 7,500-15,000 | Short phrases, cached |
| Location descriptions | 2,000 | 5-10 | 10,000-20,000 | Some cached |
| **Total (max)** | | | **~105,000** | |
| **Total (avg)** | | | **~70,000** | |

### 4.3 Cost Projections

| Plan | Monthly Cost | Characters | Buffer | Verdict |
|------|-------------|------------|--------|---------|
| Starter | €4.60 | 30,000 | ❌ Too small | Would need overages |
| Creator | €20.24 | 100,000 | ✅ 30% buffer | **Recommended** |
| Pro | €91.00 | 500,000 | ✅ Huge buffer | Overkill |

**Recommendation**: Creator Plan at €20.24/month
- 100,000 characters provides a ~30% buffer over estimated maximum usage
- Includes API access, custom voices, and higher quality models
- Fits within the €20/month budget

### 4.4 Cost Optimization Strategies

1. **Aggressive caching**: Cache all generated audio
2. **Dialogue chunking**: Break long monologues into reusable segments
3. **Voice reuse**: Assign same voice to minor NPCs
4. **Text preprocessing**: Remove markdown, normalize whitespace before hashing
5. **Preview mode**: Generate low-quality previews during prep, full quality during session
6. **Batch generation**: Pre-generate common phrases at session start

---

## 5. NPCContext Integration

### 5.1 Voice ID Assignment Flow

```
Campaign Creation
    ↓
NPCs Generated (markdown)
    ↓
DM reviews NPCs
    ↓
list_tts_voices → select voices
    ↓
assign_npc_voice for each major NPC
    ↓
Voice mappings saved to <campaign>/tts/voice-mappings.json
    ↓
NPCContext.DMContext includes voice_id
    ↓
During session: synthesize_npc_dialogue uses assigned voice
```

### 5.2 Voice Mapping Persistence

File: `<campaign_dir>/tts/voice-mappings.json`

```json
{
  "campaign_id": "shadows-of-baldur",
  "mappings": {
    "elara-moonwhisper": {
      "npc_id": "elara-moonwhisper",
      "npc_name": "Elara Moonwhisper",
      "voice_id": "XB0fDUnXU5powFXDhCwa",
      "voice_name": "Charlotte",
      "settings": {
        "stability": 0.5,
        "similarity_boost": 0.75
      }
    },
    "grukk-ironfist": {
      "npc_id": "grukk-ironfist",
      "npc_name": "Grukk Ironfist",
      "voice_id": "TxGEqnHWrfWFTfGW9XjX",
      "voice_name": "Josh",
      "settings": {
        "stability": 0.3,
        "similarity_boost": 0.5
      }
    }
  }
}
```

### 5.3 Integration with NPCContext

Update `internal/domain/dm_context.go`:

```go
type NPCContext struct {
    Name          string   `json:"name"`
    Description   string   `json:"description"`
    Motivation    string   `json:"motivation"`
    Secret        string   `json:"secret,omitempty"`
    Faction       string   `json:"faction,omitempty"`
    DialogueVoice string   `json:"dialogue_voice"`
    VoiceID       string   `json:"voice_id,omitempty"`      // NEW: TTS voice ID
    VoiceSettings TTSVoiceSettings `json:"voice_settings,omitempty"` // NEW: Per-NPC tuning
    Personality   []string `json:"personality_traits"`
    Stats         NPCStats `json:"stats,omitempty"`
    Tactics       string   `json:"tactics,omitempty"`
}
```

---

## 6. Adding Future Providers

### 6.1 Provider Interface Compliance

To add a new provider, implement the `tts.Provider` interface:

```go
type Provider interface {
    Synthesize(ctx context.Context, req domain.TTSRequest) (*domain.TTSResponse, error)
    ListVoices(ctx context.Context, language string) ([]domain.TTSVoiceInfo, error)
    GetVoice(ctx context.Context, voiceID string) (*domain.TTSVoiceInfo, error)
    IsConfigured() bool
    Name() string
    SupportsStreaming() bool
    MaxCharacters() int
}
```

### 6.2 Example: Google Cloud TTS Provider

```go
// internal/tts/google.go

type GoogleProvider struct {
    client *texttospeech.Client
    projectID string
}

func NewGoogleProvider(credentialsPath string) (*GoogleProvider, error) {
    ctx := context.Background()
    client, err := texttospeech.NewClient(ctx, option.WithCredentialsFile(credentialsPath))
    if err != nil {
        return nil, err
    }
    return &GoogleProvider{client: client}, nil
}

func (g *GoogleProvider) Name() string { return "google" }
func (g *GoogleProvider) IsConfigured() bool { return g.client != nil }
func (g *GoogleProvider) SupportsStreaming() bool { return false }
func (g *GoogleProvider) MaxCharacters() int { return 5000 }

func (g *GoogleProvider) Synthesize(ctx context.Context, req domain.TTSRequest) (*domain.TTSResponse, error) {
    // Map domain.TTSRequest to texttospeech.SynthesizeSpeechRequest
    // Call g.client.SynthesizeSpeech
    // Map response to domain.TTSResponse
}

func (g *GoogleProvider) ListVoices(ctx context.Context, language string) ([]domain.TTSVoiceInfo, error) {
    // Call g.client.ListVoices
    // Map to []domain.TTSVoiceInfo
}
```

### 6.3 Provider Registration

Update `internal/tts/provider.go`:

```go
func NewProvider(cfg Config) (Provider, error) {
    switch cfg.Provider {
    case "elevenlabs":
        return NewElevenLabsProvider(cfg.ElevenLabsKey, cfg.ElevenLabsModel)
    case "google":
        return NewGoogleProvider(cfg.GoogleCredentialsPath)
    case "amazon":
        return NewAmazonProvider(cfg.AwsRegion, cfg.AwsAccessKey, cfg.AwsSecretKey)
    case "azure":
        return NewAzureProvider(cfg.AzureKey, cfg.AzureRegion)
    case "coqui":
        return NewCoquiProvider(cfg.CoquiEndpoint)
    default:
        return nil, fmt.Errorf("unknown TTS provider: %s", cfg.Provider)
    }
}
```

### 6.4 Provider Comparison Matrix

| Provider | Cost | Quality | Latency | Streaming | Languages | Self-hosted |
|----------|------|---------|---------|-----------|-----------|-------------|
| **ElevenLabs** | €20/100k chars | Excellent | Medium | Yes | Multilingual | No |
| **Google Cloud** | $4/1M chars | Good | Low | Yes | 40+ | No |
| **Amazon Polly** | $4/1M chars | Good | Low | Yes | 30+ | No |
| **Azure TTS** | $16/1M chars | Excellent | Low | Yes | 140+ | No |
| **Coqui TTS** | Free | Medium | High | No | Limited | Yes |
| **Mimic 3** | Free | Medium | Medium | No | Limited | Yes |

### 6.5 Provider Configuration Extensions

```go
type Config struct {
    Provider             string `json:"tts_provider"`
    
    // ElevenLabs
    ElevenLabsKey        string `json:"elevenlabs_api_key,omitempty"`
    ElevenLabsModel      string `json:"elevenlabs_model,omitempty"`
    
    // Google Cloud
    GoogleCredentialsPath string `json:"google_credentials_path,omitempty"`
    GoogleProjectID      string `json:"google_project_id,omitempty"`
    
    // Amazon Polly
    AwsRegion            string `json:"aws_region,omitempty"`
    AwsAccessKey         string `json:"aws_access_key,omitempty"`
    AwsSecretKey         string `json:"aws_secret_key,omitempty"`
    
    // Azure
    AzureKey             string `json:"azure_key,omitempty"`
    AzureRegion          string `json:"azure_region,omitempty"`
    
    // Coqui / Self-hosted
    CoquiEndpoint        string `json:"coqui_endpoint,omitempty"`
    
    // Common
    CacheDir             string `json:"tts_cache_dir,omitempty"`
    CacheEnabled         bool   `json:"tts_cache_enabled"`
    DefaultVoiceID       string `json:"tts_default_voice_id,omitempty"`
    DefaultLanguage      string `json:"tts_default_language,omitempty"`
    MaxCacheSizeMB       int    `json:"tts_max_cache_size_mb,omitempty"`
}
```

---

## 7. Testing Strategy

### 7.1 Provider Tests

```go
// internal/tts/elevenlabs_test.go

func TestElevenLabsProvider_IsConfigured(t *testing.T) {
    tests := []struct {
        name   string
        apiKey string
        want   bool
    }{
        {"configured", "sk_test_123", true},
        {"empty", "", false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p := &ElevenLabsProvider{apiKey: tt.apiKey}
            if got := p.IsConfigured(); got != tt.want {
                t.Errorf("IsConfigured() = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestElevenLabsProvider_Synthesize(t *testing.T) {
    // Use httptest to mock ElevenLabs API
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("fake_audio_data"))
    }))
    defer server.Close()
    
    p := &ElevenLabsProvider{
        apiKey: "test_key",
        model:  "eleven_multilingual_v2",
        client: &http.Client{},
    }
    // Override baseURL for testing
    
    req := domain.TTSRequest{
        Text:    "Hola mundo",
        VoiceID: "test_voice",
    }
    
    resp, err := p.Synthesize(context.Background(), req)
    if err != nil {
        t.Fatalf("Synthesize failed: %v", err)
    }
    if string(resp.AudioData) != "fake_audio_data" {
        t.Errorf("Unexpected audio data")
    }
}
```

### 7.2 Service Tests

```go
// internal/services/tts_service_test.go

func TestTTSService_CacheHit(t *testing.T) {
    // Create mock provider
    mockProvider := &mockTTSProvider{}
    svc := NewTTSService(mockProvider, t.TempDir(), true, 100)
    
    req := domain.TTSRequest{Text: "Hello", VoiceID: "voice1"}
    
    // First call - cache miss
    resp1, err := svc.Synthesize(context.Background(), req)
    if err != nil {
        t.Fatal(err)
    }
    if resp1.Cached {
        t.Error("First call should be cache miss")
    }
    
    // Second call - cache hit
    resp2, err := svc.Synthesize(context.Background(), req)
    if err != nil {
        t.Fatal(err)
    }
    if !resp2.Cached {
        t.Error("Second call should be cache hit")
    }
}
```

### 7.3 Handler Tests

Follow the same pattern as existing handler tests (`internal/mcp/handlers/handlers_test.go`):

```go
func TestTTSHandlers_HandleListVoices(t *testing.T) {
    mockService := &mockTTSService{}
    handlers := NewTTSHandlers(mockService)
    
    req := mcp.CallToolRequest{
        Params: struct {
            Arguments map[string]any `json:"arguments,omitempty"`
            Meta      *struct {
                ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
            } `json:"_meta,omitempty"`
        }{
            Arguments: map[string]any{
                "language": "es",
            },
        },
    }
    
    result, err := handlers.HandleListVoices()(context.Background(), req)
    if err != nil {
        t.Fatal(err)
    }
    // Assert result
}
```

---

## 8. Session Flow Integration

### 8.1 Pre-Session (Prep Phase)

```
DM: generate_session_prep
    ↓
AI: Returns NPCs that will appear
    ↓
DM: list_tts_voices (filter by language)
    ↓
AI: Returns voice options
    ↓
DM: assign_npc_voice for each major NPC
    ↓
System: Saves mappings to campaign/tts/voice-mappings.json
    ↓
DM: Pre-generate common phrases (optional)
    synthesize_npc_dialogue for greetings, barks, etc.
```

### 8.2 During Session

```
Player interacts with NPC
    ↓
AI generates NPC response text
    ↓
DM (or AI): synthesize_npc_dialogue
    - Checks cache first
    - If miss: calls ElevenLabs API
    - Returns audio file path
    ↓
Audio plays to players
    ↓
(Next interaction: cache hit, instant playback)
```

### 8.3 Post-Session

```
Session ends
    ↓
Cache persists to disk for next session
    ↓
Voice mappings remain for campaign continuity
    ↓
New NPCs can be assigned voices before next session
```

---

## 9. Error Handling & Edge Cases

### 9.1 Provider Not Configured

```go
if !ttsProvider.IsConfigured() {
    // TTS tools return graceful error
    return mcp.NewToolResultError(
        "TTS not configured. Set ELEVENLABS_API_KEY or run 'grimorio config'"
    ), nil
}
```

### 9.2 API Rate Limiting

```go
// ElevenLabs rate limits:
// - Starter: 10 requests/second
// - Creator: 20 requests/second
// - Pro: 30 requests/second

// Solution: Per-campaign rate limiter (same pattern as AssetService)
var (
    ttsLimiters   = make(map[string]*rate.Limiter)
    ttsLimitersMu sync.RWMutex
)

func getTTSLimiter(campaign string) *rate.Limiter {
    // 5 requests per second with burst of 2
    return rate.NewLimiter(rate.Every(200*time.Millisecond), 2)
}
```

### 9.3 Text Too Long

```go
func splitText(text string, maxLen int) []string {
    // Split by sentences, respecting maxLen
    // Return chunks that can be synthesized separately
}
```

### 9.4 Cache Corruption

```go
// On read error, delete corrupt file and regenerate
if err != nil {
    os.Remove(cachePath)
    return nil, fmt.Errorf("cache read failed, regenerating: %w", err)
}
```

---

## 10. Implementation Order

### Phase 1: Foundation (Week 1)
1. Create `internal/domain/tts.go` - domain types
2. Create `internal/tts/provider.go` - interface and factory
3. Create `internal/tts/config.go` - TTS configuration
4. Update `internal/config/config.go` - embed TTS config
5. Write provider interface tests

### Phase 2: ElevenLabs Provider (Week 1)
1. Create `internal/tts/elevenlabs.go` - provider implementation
2. Write ElevenLabs provider tests with httptest mocks
3. Manual integration test with real API key

### Phase 3: Service Layer (Week 2)
1. Create `internal/services/tts_service.go`
2. Implement two-tier caching (memory + disk)
3. Implement NPC voice mapping management
4. Write comprehensive service tests

### Phase 4: MCP Integration (Week 2)
1. Create `internal/mcp/handlers/tts.go`
2. Register tools in `internal/mcp/server.go`
3. Write handler tests
4. Update `NPCContext` to include `VoiceID` and `VoiceSettings`

### Phase 5: Testing & Polish (Week 3)
1. End-to-end testing with real ElevenLabs account
2. Cost validation with actual usage
3. Performance benchmarking (cache hit rates)
4. Documentation update

### Phase 6: Future Providers (Future)
1. Google Cloud TTS provider
2. Amazon Polly provider
3. Self-hosted Coqui TTS provider (for zero-cost option)

---

## 11. Summary

This architecture provides:

- **Generic Interface**: `tts.Provider` allows swapping providers without changing business logic
- **ElevenLabs First**: Concrete implementation with full ElevenLabs v2 API support
- **Configurable**: JSON config + env vars, consistent with existing patterns
- **Budget-Conscious**: Creator plan at €20/month fits usage with 30% buffer
- **Cached**: Two-tier caching minimizes API calls and cost
- **NPC-Aware**: Voice assignments per NPC for consistent characterization
- **Extensible**: Adding new providers requires only implementing the interface + factory case
- **Testable**: Interface-based design enables full mocking

The design deliberately mirrors the existing image generation subsystem (`internal/image/`) to maintain codebase consistency and reduce cognitive load for future maintainers.
