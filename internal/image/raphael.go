package image

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RaphaelProvider generates images using Raphael AI (raphael.app)
// Free, unlimited, no API key required
type RaphaelProvider struct {
	BaseURL string
	Client  *http.Client
	Model   string
	Aspect  string
}

// RaphaelRequest represents the API request payload
type raphaelRequest struct {
	Prompt         string  `json:"prompt"`
	Aspect         string  `json:"aspect"`
	IsSafeContent  bool    `json:"isSafeContent"`
	AutoTranslate  bool    `json:"autoTranslate"`
	ModelID        string  `json:"model_id"`
	NumberOfImages int     `json:"number_of_images"`
	HighQuality    bool    `json:"highQuality"`
	FastMode       bool    `json:"fastMode"`
	TurnstileToken *string `json:"turnstileToken"`
}

// RaphaelResponse represents the API response
type raphaelResponse struct {
	URL           string `json:"url"`
	Seed          int64  `json:"seed"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	IsHighQuality bool   `json:"isHighQuality"`
}

func NewRaphaelProvider() *RaphaelProvider {
	return &RaphaelProvider{
		BaseURL: "https://raphael.app",
		Client:  &http.Client{Timeout: 120 * time.Second},
		Model:   "raphael-basic",
		Aspect:  "1:1",
	}
}

func (r *RaphaelProvider) Name() string {
	return "raphael"
}

func (r *RaphaelProvider) IsConfigured() bool {
	return true // No API key needed
}

func (r *RaphaelProvider) Generate(prompt string) ([]byte, error) {
	reqBody := raphaelRequest{
		Prompt:         prompt,
		Aspect:         r.Aspect,
		IsSafeContent:  true,
		AutoTranslate:  true,
		ModelID:        r.Model,
		NumberOfImages: 1,
		HighQuality:    false,
		FastMode:       true, // Use fast mode for better reliability
		TurnstileToken: nil,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", r.BaseURL+"/api/generate-image", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://raphael.app/")
	req.Header.Set("Accept", "application/json")

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("raphael API error %d: %s", resp.StatusCode, string(body))
	}

	var imgResp raphaelResponse
	if err := json.Unmarshal(body, &imgResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w (body: %s)", err, string(body))
	}

	if imgResp.URL == "" {
		return nil, fmt.Errorf("no image URL in response")
	}

	// Download the actual image
	imageURL := imgResp.URL
	if !bytes.HasPrefix([]byte(imageURL), []byte("http")) {
		imageURL = r.BaseURL + imageURL
	}

	return r.downloadImage(imageURL)
}

func (r *RaphaelProvider) downloadImage(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://raphael.app/")

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("image download failed %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}
