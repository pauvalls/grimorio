package image

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type DalleProvider struct {
	APIKey string
	Model  string
	Size   string
}

func NewDalleProvider(apiKey, model string) (*DalleProvider, error) {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set — DALL-E provider requires an API key")
	}
	if model == "" {
		model = "dall-e-3"
	}
	return &DalleProvider{
		APIKey: apiKey,
		Model:  model,
		Size:   "1024x1024",
	}, nil
}

func (d *DalleProvider) Name() string {
	return "dalle"
}

func (d *DalleProvider) IsConfigured() bool {
	return d.APIKey != ""
}

type dalleImageRequest struct {
	Prompt string `json:"prompt"`
	N      int    `json:"n"`
	Size   string `json:"size"`
	Model  string `json:"model,omitempty"`
}

type dalleImageResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL           string `json:"url,omitempty"`
		B64JSON       string `json:"b64_json,omitempty"`
		RevisedPrompt string `json:"revised_prompt,omitempty"`
	} `json:"data"`
}

func (d *DalleProvider) Generate(prompt string) ([]byte, error) {
	reqBody := dalleImageRequest{
		Prompt: prompt,
		N:      1,
		Size:   d.Size,
		Model:  d.Model,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/images/generations", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("DALL-E API error %d: %s", resp.StatusCode, string(body))
	}

	var imageResp dalleImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&imageResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(imageResp.Data) == 0 {
		return nil, fmt.Errorf("no image returned from DALL-E")
	}

	data := imageResp.Data[0]

	if data.B64JSON != "" {
		return base64.StdEncoding.DecodeString(data.B64JSON)
	}

	if data.URL != "" {
		return downloadImage(data.URL)
	}

	return nil, fmt.Errorf("no image data in response")
}

func downloadImage(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
