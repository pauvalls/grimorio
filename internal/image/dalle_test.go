package image

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDalleProvider_Generate_B64JSON(t *testing.T) {
	t.Parallel()

	wantData := []byte("fake-image-bytes")
	b64Data := base64.StdEncoding.EncodeToString(wantData)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("want POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/images/generations") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := dalleImageResponse{
			Data: []struct {
				URL           string `json:"url,omitempty"`
				B64JSON       string `json:"b64_json,omitempty"`
				RevisedPrompt string `json:"revised_prompt,omitempty"`
			}{
				{B64JSON: b64Data},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	d := &DalleProvider{
		APIKey:  "test-key",
		Model:   "dall-e-3",
		Size:    "1024x1024",
		BaseURL: ts.URL,
		client:  &http.Client{},
	}

	got, err := d.Generate("a red dragon")
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if string(got) != string(wantData) {
		t.Errorf("got %q, want %q", got, wantData)
	}
}

func TestDalleProvider_Generate_URL(t *testing.T) {
	t.Parallel()

	wantData := []byte("downloaded-image-bytes")

	// Secondary server that serves the actual image
	imageTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(wantData)
	}))
	defer imageTS.Close()

	// Primary server that returns the image URL
	apiTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := dalleImageResponse{
			Data: []struct {
				URL           string `json:"url,omitempty"`
				B64JSON       string `json:"b64_json,omitempty"`
				RevisedPrompt string `json:"revised_prompt,omitempty"`
			}{
				{URL: imageTS.URL + "/image.png"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer apiTS.Close()

	d := &DalleProvider{
		APIKey:  "test-key",
		Model:   "dall-e-3",
		Size:    "1024x1024",
		BaseURL: apiTS.URL,
		client:  &http.Client{},
	}

	got, err := d.Generate("a blue wizard")
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if string(got) != string(wantData) {
		t.Errorf("got %q, want %q", got, wantData)
	}
}

func TestDalleProvider_Generate_APIErrors(t *testing.T) {
	t.Parallel()

	codes := []int{http.StatusUnauthorized, http.StatusTooManyRequests, http.StatusInternalServerError}
	for _, code := range codes {
		t.Run(fmt.Sprintf("status_%d", code), func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(code)
				_, _ = w.Write([]byte(`{"error":"something went wrong"}`))
			}))
			defer ts.Close()

			d := &DalleProvider{
				APIKey:  "test-key",
				Model:   "dall-e-3",
				Size:    "1024x1024",
				BaseURL: ts.URL,
				client:  &http.Client{},
			}

			_, err := d.Generate("prompt")
			if err == nil {
				t.Fatal("expected error for non-200 status")
			}
			if !strings.Contains(err.Error(), fmt.Sprintf("%d", code)) {
				t.Errorf("error should contain status code %d: %v", code, err)
			}
		})
	}
}

func TestDalleProvider_Generate_EmptyData(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := dalleImageResponse{Data: []struct {
			URL           string `json:"url,omitempty"`
			B64JSON       string `json:"b64_json,omitempty"`
			RevisedPrompt string `json:"revised_prompt,omitempty"`
		}{}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	d := &DalleProvider{
		APIKey:  "test-key",
		Model:   "dall-e-3",
		Size:    "1024x1024",
		BaseURL: ts.URL,
		client:  &http.Client{},
	}

	_, err := d.Generate("prompt")
	if err == nil {
		t.Fatal("expected error for empty data")
	}
	if !strings.Contains(err.Error(), "no image returned") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDalleProvider_Generate_InvalidJSON(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{not valid json`))
	}))
	defer ts.Close()

	d := &DalleProvider{
		APIKey:  "test-key",
		Model:   "dall-e-3",
		Size:    "1024x1024",
		BaseURL: ts.URL,
		client:  &http.Client{},
	}

	_, err := d.Generate("prompt")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDalleProvider_Generate_BadBase64(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := dalleImageResponse{
			Data: []struct {
				URL           string `json:"url,omitempty"`
				B64JSON       string `json:"b64_json,omitempty"`
				RevisedPrompt string `json:"revised_prompt,omitempty"`
			}{
				{B64JSON: "!!!not-valid-base64!!!"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	d := &DalleProvider{
		APIKey:  "test-key",
		Model:   "dall-e-3",
		Size:    "1024x1024",
		BaseURL: ts.URL,
		client:  &http.Client{},
	}

	_, err := d.Generate("prompt")
	if err == nil {
		t.Fatal("expected error for bad base64")
	}
	if !strings.Contains(err.Error(), "base64") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDownloadImage_NetworkFailure(t *testing.T) {
	t.Parallel()

	// Use a server that immediately closes
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts.Close() // close immediately to force network failure

	_, err := downloadImage(ts.URL + "/image.png")
	if err == nil {
		t.Fatal("expected error for closed server")
	}
}

func TestDownloadImage_NotFound(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer ts.Close()

	data, err := downloadImage(ts.URL + "/missing.png")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if data != nil {
		t.Errorf("expected nil data, got %v", data)
	}
}

func TestDownloadImage_Success(t *testing.T) {
	t.Parallel()

	want := []byte("image-data")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(want)
	}))
	defer ts.Close()

	got, err := downloadImage(ts.URL + "/image.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("got %q, want %q", got, want)
	}
}
