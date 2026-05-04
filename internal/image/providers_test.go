package image

import (
	"os"
	"testing"
)

func TestNewPollinationsProvider(t *testing.T) {
	p := NewPollinationsProvider(512, 512, 42)
	if p == nil {
		t.Fatal("NewPollinationsProvider() returned nil")
	}
	if p.Name() != "pollinations" {
		t.Errorf("Name() = %s, want pollinations", p.Name())
	}
	if !p.IsConfigured() {
		t.Error("IsConfigured() should return true")
	}
	if p.Width != 512 {
		t.Errorf("Width = %d, want 512", p.Width)
	}
	if p.Height != 512 {
		t.Errorf("Height = %d, want 512", p.Height)
	}
	if p.Seed != 42 {
		t.Errorf("Seed = %d, want 42", p.Seed)
	}
}

func TestNewPollinationsProvider_Defaults(t *testing.T) {
	p := NewPollinationsProvider(0, 0, -1)
	if p.Width != 1024 {
		t.Errorf("Default Width = %d, want 1024", p.Width)
	}
	if p.Height != 1024 {
		t.Errorf("Default Height = %d, want 1024", p.Height)
	}
}

func TestNewRaphaelProvider(t *testing.T) {
	r := NewRaphaelProvider()
	if r == nil {
		t.Fatal("NewRaphaelProvider() returned nil")
	}
	if r.Name() != "raphael" {
		t.Errorf("Name() = %s, want raphael", r.Name())
	}
	if !r.IsConfigured() {
		t.Error("IsConfigured() should return true")
	}
	if r.BaseURL != "https://raphael.app" {
		t.Errorf("BaseURL = %s, want https://raphael.app", r.BaseURL)
	}
}

func TestNewDalleProvider(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	d, err := NewDalleProvider("", "")
	if err != nil {
		t.Fatalf("NewDalleProvider() error: %v", err)
	}
	if d == nil {
		t.Fatal("NewDalleProvider() returned nil")
	}
	if d.Name() != "dalle" {
		t.Errorf("Name() = %s, want dalle", d.Name())
	}
	if !d.IsConfigured() {
		t.Error("IsConfigured() should return true")
	}
	if d.Model != "dall-e-3" {
		t.Errorf("Default Model = %s, want dall-e-3", d.Model)
	}
}

func TestNewDalleProvider_MissingKey(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")

	_, err := NewDalleProvider("", "")
	if err == nil {
		t.Error("NewDalleProvider() should error without API key")
	}
}

func TestNewDalleProvider_CustomModel(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	d, _ := NewDalleProvider("test-key", "dall-e-2")
	if d.Model != "dall-e-2" {
		t.Errorf("Model = %s, want dall-e-2", d.Model)
	}
}
