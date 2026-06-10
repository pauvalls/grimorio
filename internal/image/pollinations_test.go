package image

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestPollinationsProvider_Generate_QueryParams(t *testing.T) {
	t.Parallel()

	var capturedURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("image-bytes"))
	}))
	defer ts.Close()

	p := &PollinationsProvider{
		BaseURL: ts.URL + "/prompt",
		Width:   512,
		Height:  512,
		Seed:    42,
		Model:   "flux",
		client:  &http.Client{},
	}

	_, err := p.Generate("test prompt")
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	// Parse the captured URL to verify query params
	parsed, err := url.Parse(capturedURL)
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	q := parsed.Query()
	if q.Get("width") != "512" {
		t.Errorf("width = %q, want 512", q.Get("width"))
	}
	if q.Get("height") != "512" {
		t.Errorf("height = %q, want 512", q.Get("height"))
	}
	if q.Get("seed") != "42" {
		t.Errorf("seed = %q, want 42", q.Get("seed"))
	}
	if q.Get("model") != "flux" {
		t.Errorf("model = %q, want flux", q.Get("model"))
	}
	if q.Get("nologo") != "true" {
		t.Errorf("nologo = %q, want true", q.Get("nologo"))
	}
}

func TestPollinationsProvider_Generate_Success(t *testing.T) {
	t.Parallel()

	wantData := []byte("pollinations-image-data")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(wantData)
	}))
	defer ts.Close()

	p := &PollinationsProvider{
		BaseURL: ts.URL + "/prompt",
		Width:   1024,
		Height:  1024,
		Seed:    -1,
		client:  &http.Client{},
	}

	got, err := p.Generate("a castle")
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if string(got) != string(wantData) {
		t.Errorf("got %q, want %q", got, wantData)
	}
}

func TestPollinationsProvider_Generate_Non200(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"server error"}`))
	}))
	defer ts.Close()

	p := &PollinationsProvider{
		BaseURL: ts.URL + "/prompt",
		Width:   1024,
		Height:  1024,
		client:  &http.Client{},
	}

	_, err := p.Generate("prompt")
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should contain 500: %v", err)
	}
}

func TestPollinationsProvider_Generate_ClosedServer(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts.Close()

	p := &PollinationsProvider{
		BaseURL: ts.URL + "/prompt",
		Width:   1024,
		Height:  1024,
		client:  &http.Client{},
	}

	_, err := p.Generate("prompt")
	if err == nil {
		t.Fatal("expected error for closed server")
	}
}

func TestPollinationsProvider_Generate_NoSeed(t *testing.T) {
	t.Parallel()

	var capturedURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("image"))
	}))
	defer ts.Close()

	p := &PollinationsProvider{
		BaseURL: ts.URL + "/prompt",
		Width:   1024,
		Height:  1024,
		Seed:    -1,
		client:  &http.Client{},
	}

	_, err := p.Generate("prompt")
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	if strings.Contains(capturedURL, "seed=") {
		t.Error("URL should not contain seed when Seed < 0")
	}
}

func TestPollinationsProvider_Generate_NoModel(t *testing.T) {
	t.Parallel()

	var capturedURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("image"))
	}))
	defer ts.Close()

	p := &PollinationsProvider{
		BaseURL: ts.URL + "/prompt",
		Width:   1024,
		Height:  1024,
		Model:   "",
		client:  &http.Client{},
	}

	_, err := p.Generate("prompt")
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	if strings.Contains(capturedURL, "model=") {
		t.Error("URL should not contain model when Model is empty")
	}
}
