package dalle

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	client := New("test-key")
	if client == nil {
		t.Fatal("New() returned nil")
	}
	if client.APIKey != "test-key" {
		t.Errorf("expected APIKey 'test-key', got '%s'", client.APIKey)
	}
	if client.Model != "dall-e-3" {
		t.Errorf("expected Model 'dall-e-3', got '%s'", client.Model)
	}
	if client.Size != "1024x1024" {
		t.Errorf("expected Size '1024x1024', got '%s'", client.Size)
	}
}

func TestNew_FromEnv(t *testing.T) {
	_ = os.Setenv("OPENAI_API_KEY", "env-key")
	defer func() { _ = os.Unsetenv("OPENAI_API_KEY") }()

	client := New("")
	if client == nil {
		t.Fatal("New() returned nil")
	}
	if client.APIKey != "env-key" {
		t.Errorf("expected APIKey from env, got '%s'", client.APIKey)
	}
}

func TestIsConfigured(t *testing.T) {
	client := New("test-key")
	if !client.IsConfigured() {
		t.Error("IsConfigured() should return true when API key is set")
	}

	client2 := New("")
	if client2.IsConfigured() {
		t.Error("IsConfigured() should return false when API key is empty")
	}
}

func TestImageRequestStruct(t *testing.T) {
	req := ImageRequest{
		Prompt: "test prompt",
		N:      1,
		Size:   "1024x1024",
		Model:  "dall-e-3",
	}
	if req.Prompt != "test prompt" {
		t.Error("ImageRequest Prompt field not set correctly")
	}
}

func TestImageResponseStruct(t *testing.T) {
	resp := ImageResponse{
		Created: 1234567890,
	}
	if resp.Created != 1234567890 {
		t.Error("ImageResponse Created field not set correctly")
	}
}
