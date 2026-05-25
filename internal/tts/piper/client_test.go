package piper

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientSynthesize(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			body, _ := io.ReadAll(r.Body)
			if string(body) != "hola mundo" {
				t.Errorf("expected body 'hola mundo', got %s", string(body))
			}
			w.Header().Set("Content-Type", "audio/wav")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("RIFF....WAV"))
		}))
		defer server.Close()

		client := NewClientWithHTTP(server.URL, server.Client())
		ctx := context.Background()

		reader, err := client.Synthesize(ctx, "hola mundo")
		if err != nil {
			t.Fatalf("Synthesize() error = %v", err)
		}
		defer func() { _ = reader.Close() }()

		data, _ := io.ReadAll(reader)
		if string(data) != "RIFF....WAV" {
			t.Errorf("expected WAV data, got %s", string(data))
		}
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("internal error"))
		}))
		defer server.Close()

		client := NewClientWithHTTP(server.URL, server.Client())
		ctx := context.Background()

		_, err := client.Synthesize(ctx, "test")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "500") {
			t.Errorf("expected error to contain 500, got %v", err)
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClientWithHTTP(server.URL, server.Client())
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, err := client.Synthesize(ctx, "test")
		if err == nil {
			t.Fatal("expected error due to timeout, got nil")
		}
	})
}

func TestClientHealthCheck(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClientWithHTTP(server.URL, server.Client())
		ctx := context.Background()

		if err := client.HealthCheck(ctx); err != nil {
			t.Fatalf("HealthCheck() error = %v", err)
		}
	})

	t.Run("unhealthy", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		client := NewClientWithHTTP(server.URL, server.Client())
		ctx := context.Background()

		err := client.HealthCheck(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "503") {
			t.Errorf("expected error to contain 503, got %v", err)
		}
	})

	t.Run("server unreachable", func(t *testing.T) {
		// No server — URL that refuses connections
		client := NewClient("127.0.0.1", 59999)
		ctx := context.Background()

		err := client.HealthCheck(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestClientClose(t *testing.T) {
	client := NewClient("127.0.0.1", 5000)
	if err := client.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
