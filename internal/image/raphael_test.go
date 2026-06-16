package image

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRaphaelProvider_Generate_RequestBody(t *testing.T) {
	t.Parallel()

	var capturedBody []byte
	var serverURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/api/generate-image") {
			if r.Method != http.MethodPost {
				t.Errorf("want POST, got %s", r.Method)
			}
			capturedBody, _ = io.ReadAll(r.Body)
			resp := raphaelResponse{URL: serverURL + "/image.png"}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		if r.URL.Path == "/image.png" {
			_, _ = w.Write([]byte("image-bytes"))
			return
		}
		t.Errorf("unexpected path: %s", r.URL.Path)
	}))
	defer ts.Close()
	serverURL = ts.URL

	r := &RaphaelProvider{
		BaseURL: ts.URL,
		Client:  &http.Client{},
		Model:   "raphael-basic",
		Aspect:  "1:1",
	}

	_, err := r.Generate("dragon portrait")
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(capturedBody, &body); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	if body["prompt"] != "dragon portrait" {
		t.Errorf("prompt = %v, want 'dragon portrait'", body["prompt"])
	}
	if body["aspect"] != "1:1" {
		t.Errorf("aspect = %v, want '1:1'", body["aspect"])
	}
	if body["model_id"] != "raphael-basic" {
		t.Errorf("model_id = %v, want 'raphael-basic'", body["model_id"])
	}
	if body["fastMode"] != true {
		t.Errorf("fastMode = %v, want true", body["fastMode"])
	}
}

func TestRaphaelProvider_Generate_RelativeURL(t *testing.T) {
	t.Parallel()

	wantData := []byte("raphael-image-bytes")

	// Track which endpoint was hit
	var imageHit bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/api/generate-image") {
			resp := raphaelResponse{URL: "/generated/image.png"}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		if r.URL.Path == "/generated/image.png" {
			imageHit = true
			_, _ = w.Write(wantData)
			return
		}
		t.Errorf("unexpected path: %s", r.URL.Path)
	}))
	defer ts.Close()

	r := &RaphaelProvider{
		BaseURL: ts.URL,
		Client:  &http.Client{},
		Model:   "raphael-basic",
		Aspect:  "1:1",
	}

	got, err := r.Generate("prompt")
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if !imageHit {
		t.Error("image endpoint was not hit")
	}
	if string(got) != string(wantData) {
		t.Errorf("got %q, want %q", got, wantData)
	}
}

func TestRaphaelProvider_Generate_AbsoluteURL(t *testing.T) {
	t.Parallel()

	wantData := []byte("absolute-image-bytes")

	// Image server on a different "host"
	imageTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(wantData)
	}))
	defer imageTS.Close()

	// API server that returns absolute URL
	apiTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := raphaelResponse{URL: imageTS.URL + "/image.png"}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer apiTS.Close()

	r := &RaphaelProvider{
		BaseURL: apiTS.URL,
		Client:  &http.Client{},
		Model:   "raphael-basic",
		Aspect:  "1:1",
	}

	got, err := r.Generate("prompt")
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if string(got) != string(wantData) {
		t.Errorf("got %q, want %q", got, wantData)
	}
}

func TestRaphaelProvider_Generate_APIErrors(t *testing.T) {
	t.Parallel()

	codes := []int{http.StatusBadRequest, http.StatusInternalServerError}
	for _, code := range codes {
		t.Run(http.StatusText(code), func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(code)
				_, _ = w.Write([]byte(`{"error":"fail"}`))
			}))
			defer ts.Close()

			r := &RaphaelProvider{
				BaseURL: ts.URL,
				Client:  &http.Client{},
				Model:   "raphael-basic",
				Aspect:  "1:1",
			}

			_, err := r.Generate("prompt")
			if err == nil {
				t.Fatal("expected error for non-200 status")
			}
			if !strings.Contains(err.Error(), string(rune('0'+code/100))) {
				// Just check it contains some error info
				t.Logf("error: %v", err)
			}
		})
	}
}

func TestRaphaelProvider_Generate_InvalidJSON(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{invalid`))
	}))
	defer ts.Close()

	r := &RaphaelProvider{
		BaseURL: ts.URL,
		Client:  &http.Client{},
		Model:   "raphael-basic",
		Aspect:  "1:1",
	}

	_, err := r.Generate("prompt")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRaphaelProvider_Generate_MissingURL(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := raphaelResponse{URL: ""}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	r := &RaphaelProvider{
		BaseURL: ts.URL,
		Client:  &http.Client{},
		Model:   "raphael-basic",
		Aspect:  "1:1",
	}

	_, err := r.Generate("prompt")
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
	if !strings.Contains(err.Error(), "no image URL") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRaphaelProvider_downloadImage_NetworkFailure(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts.Close()

	r := &RaphaelProvider{
		Client: &http.Client{},
	}

	_, err := r.downloadImage(ts.URL + "/image.png")
	if err == nil {
		t.Fatal("expected error for closed server")
	}
}

func TestRaphaelProvider_downloadImage_NotFound(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer ts.Close()

	r := &RaphaelProvider{
		Client: &http.Client{},
	}

	_, err := r.downloadImage(ts.URL + "/missing.png")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error should contain 404: %v", err)
	}
}

func TestRaphaelProvider_downloadImage_Success(t *testing.T) {
	t.Parallel()

	want := []byte("raphael-downloaded-image")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(want)
	}))
	defer ts.Close()

	r := &RaphaelProvider{
		Client: &http.Client{},
	}

	got, err := r.downloadImage(ts.URL + "/image.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("got %q, want %q", got, want)
	}
}
