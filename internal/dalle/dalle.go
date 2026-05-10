package dalle

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Client struct {
	APIKey string
	Model  string
	Size   string
}

type ImageRequest struct {
	Prompt string `json:"prompt"`
	N      int    `json:"n"`
	Size   string `json:"size"`
	Model  string `json:"model,omitempty"`
}

type ImageResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL           string `json:"url,omitempty"`
		B64JSON       string `json:"b64_json,omitempty"`
		RevisedPrompt string `json:"revised_prompt,omitempty"`
	} `json:"data"`
}

func New(apiKey string) *Client {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	return &Client{
		APIKey: apiKey,
		Model:  "dall-e-3",
		Size:   "1024x1024",
	}
}

func (c *Client) IsConfigured() bool {
	return c.APIKey != ""
}

func (c *Client) Generate(prompt string) (string, error) {
	if !c.IsConfigured() {
		return "", fmt.Errorf("OPENAI_API_KEY not set — DALL-E generation requires an API key")
	}

	reqBody := ImageRequest{
		Prompt: prompt,
		N:      1,
		Size:   c.Size,
		Model:  c.Model,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/images/generations", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("DALL-E API error %d: %s", resp.StatusCode, string(body))
	}

	var imageResp ImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&imageResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(imageResp.Data) == 0 {
		return "", fmt.Errorf("no image returned from DALL-E")
	}

	data := imageResp.Data[0]

	if data.B64JSON != "" {
		decoded, err := base64.StdEncoding.DecodeString(data.B64JSON)
		if err != nil {
			return "", fmt.Errorf("failed to decode base64 image: %w", err)
		}
		return string(decoded), nil
	}

	if data.URL != "" {
		imgResp, err := http.Get(data.URL)
		if err != nil {
			return "", fmt.Errorf("failed to download image: %w", err)
		}
		defer func() { _ = imgResp.Body.Close() }()

		imgData, err := io.ReadAll(imgResp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read image: %w", err)
		}
		return string(imgData), nil
	}

	return "", fmt.Errorf("no image data in response")
}

func (c *Client) GenerateAndSave(prompt, outputPath string) error {
	imgData, err := c.Generate(prompt)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return os.WriteFile(outputPath, []byte(imgData), 0644)
}

func (c *Client) GenerateAndSaveBase64(prompt, outputPath string) (string, error) {
	imgData, err := c.Generate(prompt)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(imgData), 0644); err != nil {
		return "", fmt.Errorf("failed to write image: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(imgData))
	return fmt.Sprintf("data:image/png;base64,%s", encoded), nil
}
