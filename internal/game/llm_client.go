package game

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// LLMClient defines the interface for LLM providers
type LLMClient interface {
	Chat(ctx context.Context, messages []Message) (string, error)
	ChatStream(ctx context.Context, messages []Message) (<-chan string, error)
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMConfig holds configuration for LLM providers
type LLMConfig struct {
	Provider string `json:"provider"` // openai, anthropic, kimi, opencode
	APIKey   string `json:"api_key"`
	BaseURL  string `json:"base_url"`
	Model    string `json:"model"`
}

// NewLLMClient creates a new LLM client based on configuration
func NewLLMClient(cfg LLMConfig) (LLMClient, error) {
	switch cfg.Provider {
	case "openai":
		return NewOpenAIClient(cfg), nil
	case "anthropic":
		return NewAnthropicClient(cfg), nil
	case "kimi", "moonshot":
		return NewKimiClient(cfg), nil
	case "opencode":
		return NewOpenCodeClient(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.Provider)
	}
}

// OpenAIClient implements LLMClient for OpenAI API
type OpenAIClient struct {
	config LLMConfig
	client *http.Client
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(cfg LLMConfig) *OpenAIClient {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-4"
	}
	return &OpenAIClient{
		config: cfg,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// Chat sends a chat request to OpenAI
func (c *OpenAIClient) Chat(ctx context.Context, messages []Message) (string, error) {
	requestBody := map[string]interface{}{
		"model":    c.config.Model,
		"messages": messages,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return result.Choices[0].Message.Content, nil
}

// ChatStream sends a streaming chat request to OpenAI
func (c *OpenAIClient) ChatStream(ctx context.Context, messages []Message) (<-chan string, error) {
	// TODO: Implement streaming
	return nil, fmt.Errorf("streaming not yet implemented")
}

// AnthropicClient implements LLMClient for Anthropic API
type AnthropicClient struct {
	config LLMConfig
	client *http.Client
}

// NewAnthropicClient creates a new Anthropic client
func NewAnthropicClient(cfg LLMConfig) *AnthropicClient {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.anthropic.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "claude-3-sonnet-20240229"
	}
	return &AnthropicClient{
		config: cfg,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// Chat sends a chat request to Anthropic
func (c *AnthropicClient) Chat(ctx context.Context, messages []Message) (string, error) {
	// Convert messages to Anthropic format
	var systemMessage string
	var conversationMessages []map[string]string
	
	for _, msg := range messages {
		if msg.Role == "system" {
			systemMessage = msg.Content
		} else {
			conversationMessages = append(conversationMessages, map[string]string{
				"role":    msg.Role,
				"content": msg.Content,
			})
		}
	}

	requestBody := map[string]interface{}{
		"model":    c.config.Model,
		"messages": conversationMessages,
		"max_tokens": 4096,
	}
	
	if systemMessage != "" {
		requestBody["system"] = systemMessage
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return result.Content[0].Text, nil
}

// ChatStream sends a streaming chat request to Anthropic
func (c *AnthropicClient) ChatStream(ctx context.Context, messages []Message) (<-chan string, error) {
	return nil, fmt.Errorf("streaming not yet implemented")
}

// KimiClient implements LLMClient for Moonshot AI (Kimi) API
type KimiClient struct {
	config LLMConfig
	client *http.Client
}

// NewKimiClient creates a new Kimi client
func NewKimiClient(cfg LLMConfig) *KimiClient {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.moonshot.cn/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "moonshot-v1-8k"
	}
	return &KimiClient{
		config: cfg,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// Chat sends a chat request to Kimi API
func (c *KimiClient) Chat(ctx context.Context, messages []Message) (string, error) {
	requestBody := map[string]interface{}{
		"model":    c.config.Model,
		"messages": messages,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Kimi API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from Kimi API")
	}

	return result.Choices[0].Message.Content, nil
}

// ChatStream sends a streaming chat request to Kimi API
func (c *KimiClient) ChatStream(ctx context.Context, messages []Message) (<-chan string, error) {
	return nil, fmt.Errorf("streaming not yet implemented")
}

// OpenCodeClient implements LLMClient for OpenCode API
type OpenCodeClient struct {
	config LLMConfig
	client *http.Client
}

// NewOpenCodeClient creates a new OpenCode client
func NewOpenCodeClient(cfg LLMConfig) *OpenCodeClient {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.opencode.ai/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "opencode-default"
	}
	return &OpenCodeClient{
		config: cfg,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// Chat sends a chat request to OpenCode API
func (c *OpenCodeClient) Chat(ctx context.Context, messages []Message) (string, error) {
	requestBody := map[string]interface{}{
		"model":    c.config.Model,
		"messages": messages,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenCode API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenCode API")
	}

	return result.Choices[0].Message.Content, nil
}

// ChatStream sends a streaming chat request to OpenCode API
func (c *OpenCodeClient) ChatStream(ctx context.Context, messages []Message) (<-chan string, error) {
	return nil, fmt.Errorf("streaming not yet implemented")
}

// DefaultLLMConfig returns a default LLM configuration
func DefaultLLMConfig() LLMConfig {
	return LLMConfig{
		Provider: "openai",
		Model:    "gpt-4",
	}
}

// LoadLLMConfigFromEnv loads LLM configuration from environment variables
func LoadLLMConfigFromEnv() LLMConfig {
	cfg := DefaultLLMConfig()
	
	// Check for Kimi API key
	if key := os.Getenv("KIMI_API_KEY"); key != "" {
		cfg.Provider = "kimi"
		cfg.APIKey = key
		if model := os.Getenv("KIMI_MODEL"); model != "" {
			cfg.Model = model
		}
		return cfg
	}
	
	// Check for OpenCode API key
	if key := os.Getenv("OPENCODE_API_KEY"); key != "" {
		cfg.Provider = "opencode"
		cfg.APIKey = key
		if model := os.Getenv("OPENCODE_MODEL"); model != "" {
			cfg.Model = model
		}
		return cfg
	}
	
	// Check for OpenAI API key
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		cfg.Provider = "openai"
		cfg.APIKey = key
		if model := os.Getenv("OPENAI_MODEL"); model != "" {
			cfg.Model = model
		}
		return cfg
	}
	
	// Check for Anthropic API key
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		cfg.Provider = "anthropic"
		cfg.APIKey = key
		if model := os.Getenv("ANTHROPIC_MODEL"); model != "" {
			cfg.Model = model
		}
		return cfg
	}
	
	return cfg
}
