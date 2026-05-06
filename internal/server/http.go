package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/pauvalls/grimorio/internal/config"
	"github.com/pauvalls/grimorio/internal/game"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/websocket"
)

// HTTPServer handles HTTP requests for the game engine
type HTTPServer struct {
	config      *config.Config
	engine      *game.Engine
	hub         *websocket.Hub
	campaignRepo repository.CampaignRepository
}

// NewHTTPServer creates a new HTTP server
func NewHTTPServer(cfg *config.Config, engine *game.Engine, hub *websocket.Hub) *HTTPServer {
	return &HTTPServer{
		config:       cfg,
		engine:       engine,
		hub:          hub,
		campaignRepo: repository.NewFilesystemCampaignRepository(cfg.OutputDir),
	}
}

// ServeHTTP implements the http.Handler interface
func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Route handling
	path := r.URL.Path
	
	// WebSocket endpoint
	if path == "/ws" || strings.HasPrefix(path, "/ws/") {
		s.hub.HandleWebSocket(w, r)
		return
	}
	
	// API routes
	if strings.HasPrefix(path, "/api/") {
		s.handleAPI(w, r)
		return
	}
	
	// Health check
	if path == "/health" {
		s.handleHealth(w, r)
		return
	}
	
	// Static files (frontend)
	if path == "/" || path == "/index.html" {
		s.serveStatic(w, r, "index.html")
		return
	}
	
	// 404 for everything else
	http.NotFound(w, r)
}

func (s *HTTPServer) handleAPI(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/")
	parts := strings.Split(path, "/")
	
	if len(parts) == 0 {
		s.sendError(w, http.StatusNotFound, "Not found")
		return
	}
	
	switch parts[0] {
	case "campaigns":
		s.handleCampaigns(w, r, parts[1:])
	case "sessions":
		s.handleSessions(w, r, parts[1:])
	case "characters":
		s.handleCharacters(w, r, parts[1:])
	default:
		s.sendError(w, http.StatusNotFound, "Not found")
	}
}

func (s *HTTPServer) handleCampaigns(w http.ResponseWriter, r *http.Request, parts []string) {
	switch r.Method {
	case http.MethodGet:
		if len(parts) == 0 {
			// List campaigns
			campaigns, err := s.campaignRepo.List()
			if err != nil {
				s.sendError(w, http.StatusInternalServerError, err.Error())
				return
			}
			s.sendJSON(w, http.StatusOK, campaigns)
		} else {
			// Get specific campaign
			campaign, err := s.campaignRepo.Read(parts[0])
			if err != nil {
				s.sendError(w, http.StatusNotFound, "Campaign not found")
				return
			}
			s.sendJSON(w, http.StatusOK, campaign)
		}
	default:
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *HTTPServer) handleSessions(w http.ResponseWriter, r *http.Request, parts []string) {
	switch r.Method {
	case http.MethodPost:
		if len(parts) == 0 {
			// Create session
			var req struct {
				CampaignID string   `json:"campaign_id"`
				Players    []string `json:"players"`
			}
			
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				s.sendError(w, http.StatusBadRequest, "Invalid request body")
				return
			}
			
			if req.CampaignID == "" || len(req.Players) == 0 {
				s.sendError(w, http.StatusBadRequest, "campaign_id and players are required")
				return
			}
			
			session, err := s.engine.StartSession(req.CampaignID, req.Players)
			if err != nil {
				s.sendError(w, http.StatusInternalServerError, err.Error())
				return
			}
			
			s.sendJSON(w, http.StatusCreated, session)
		} else if len(parts) == 2 && parts[1] == "end" {
			// End session
			summary, err := s.engine.EndSession(parts[0])
			if err != nil {
				s.sendError(w, http.StatusInternalServerError, err.Error())
				return
			}
			s.sendJSON(w, http.StatusOK, summary)
		}
		
	case http.MethodGet:
		if len(parts) == 0 {
			// List sessions would need a query parameter for campaign_id
			s.sendError(w, http.StatusBadRequest, "campaign_id query parameter required")
			return
		}
		
		// Get session state
		if len(parts) == 2 && parts[1] == "state" {
			state, err := s.engine.GetState(parts[0])
			if err != nil {
				s.sendError(w, http.StatusNotFound, err.Error())
				return
			}
			s.sendJSON(w, http.StatusOK, state)
			return
		}
		
		// Get session events
		if len(parts) == 2 && parts[1] == "events" {
			limitStr := r.URL.Query().Get("limit")
			limit := 50
			if limitStr != "" {
				fmt.Sscanf(limitStr, "%d", &limit)
			}
			
			events, err := s.engine.GetRecentEvents(parts[0], limit)
			if err != nil {
				s.sendError(w, http.StatusInternalServerError, err.Error())
				return
			}
			s.sendJSON(w, http.StatusOK, events)
			return
		}
		
		// Get session by ID
		session, err := s.engine.GetSession(parts[0])
		if err != nil {
			s.sendError(w, http.StatusNotFound, err.Error())
			return
		}
		s.sendJSON(w, http.StatusOK, session)
		
	default:
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *HTTPServer) handleCharacters(w http.ResponseWriter, r *http.Request, parts []string) {
	switch r.Method {
	case http.MethodGet:
		if len(parts) < 2 {
			s.sendError(w, http.StatusBadRequest, "campaign_id and character_name required")
			return
		}
		
		// This would need a character repository
		s.sendError(w, http.StatusNotImplemented, "Not implemented")
		
	default:
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"mode":   "game_engine",
	})
}

func (s *HTTPServer) serveStatic(w http.ResponseWriter, r *http.Request, filename string) {
	// For now, serve a simple HTML page
	// In production, this would serve the built frontend
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<title>Grimorio Game Engine</title>
	<style>
		body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
		h1 { color: #333; }
		.status { padding: 10px; background: #e8f5e9; border-radius: 4px; }
	</style>
</head>
<body>
	<h1>Grimorio Game Engine</h1>
	<div class="status">
		<p>Server is running! Connect via WebSocket at <code>/ws</code></p>
		<p>API endpoints available at <code>/api/</code></p>
	</div>
</body>
</html>`))
}

func (s *HTTPServer) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func (s *HTTPServer) sendError(w http.ResponseWriter, status int, message string) {
	s.sendJSON(w, status, map[string]interface{}{
		"error":   message,
		"status":  status,
	})
}

// Start starts the HTTP server
func Start(cfg *config.Config, engine *game.Engine, hub *websocket.Hub) error {
	server := NewHTTPServer(cfg, engine, hub)
	
	addr := fmt.Sprintf("%s:%d", "localhost", 8080)
	if cfg.GameEngine.BaseURL != "" {
		// Extract port from base URL if provided
		// For now, just use default
	}
	
	log.Printf("Starting Grimorio Game Engine on %s", addr)
	
	return http.ListenAndServe(addr, server)
}
