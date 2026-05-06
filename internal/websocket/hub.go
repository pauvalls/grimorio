package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/game"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	// Client -> Server
	MsgPlayerAction   MessageType = "player_action"
	MsgRollDice       MessageType = "roll_dice"
	MsgMoveToken      MessageType = "move_token"
	MsgJoinSession    MessageType = "join_session"
	MsgStartCombat    MessageType = "start_combat"
	MsgEndCombat      MessageType = "end_combat"
	MsgNextTurn       MessageType = "next_turn"
	MsgAttack         MessageType = "attack"
	MsgGetState       MessageType = "get_state"
	
	// Server -> Client
	MsgNarration      MessageType = "narration"
	MsgDiceResult     MessageType = "dice_result"
	MsgStateUpdate    MessageType = "state_update"
	MsgEvent          MessageType = "event"
	MsgCombatTurn     MessageType = "combat_turn"
	MsgError          MessageType = "error"
	MsgConnected      MessageType = "connected"
	MsgDisconnected   MessageType = "disconnected"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType    `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time      `json:"timestamp"`
}

// Client represents a connected WebSocket client
type Client struct {
	hub        *Hub
	conn       *websocket.Conn
	send       chan Message
	sessionID  string
	characterID string
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	clients    map[string]*Client // sessionID -> client
	register   chan *Client
	unregister chan *Client
	broadcast  chan Message
	mu         sync.RWMutex
	engine     *game.Engine
}

// NewHub creates a new WebSocket hub
func NewHub(engine *game.Engine) *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Message, 256),
		engine:     engine,
	}
}

// Run starts the hub's event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.sessionID+":"+client.characterID] = client
			h.mu.Unlock()
			log.Printf("Client connected: %s/%s", client.sessionID, client.characterID)
			
			// Send connected confirmation
			client.send <- Message{
				Type:      MsgConnected,
				Payload:   map[string]interface{}{"session_id": client.sessionID},
				Timestamp: time.Now(),
			}
			
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.sessionID+":"+client.characterID]; ok {
				delete(h.clients, client.sessionID+":"+client.characterID)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("Client disconnected: %s/%s", client.sessionID, client.characterID)
			
		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client.sessionID+":"+client.characterID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToSession sends a message to all clients in a session
func (h *Hub) BroadcastToSession(sessionID string, message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	for key, client := range h.clients {
		if client.sessionID == sessionID {
			select {
			case client.send <- message:
			default:
				log.Printf("Failed to send to client %s", key)
			}
		}
	}
}

// HandleWebSocket handles WebSocket connections
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // Allow all origins for now
	})
	if err != nil {
		log.Printf("WebSocket accept error: %v", err)
		return
	}
	defer conn.Close(websocket.StatusInternalError, "connection closed")

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan Message, 256),
	}

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// readPump reads messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		var msg Message
		err := wsjson.Read(ctx, c.conn, &msg)
		cancel()
		
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				return
			}
			log.Printf("WebSocket read error: %v", err)
			return
		}

		msg.Timestamp = time.Now()
		c.handleMessage(msg)
	}
}

// writePump writes messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.Close(websocket.StatusInternalError, "channel closed")
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := wsjson.Write(ctx, c.conn, message)
			cancel()
			
			if err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			// Send ping to keep connection alive
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := c.conn.Ping(ctx)
			cancel()
			
			if err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages
func (c *Client) handleMessage(msg Message) {
	switch msg.Type {
	case MsgJoinSession:
		c.handleJoinSession(msg.Payload)
	case MsgPlayerAction:
		c.handlePlayerAction(msg.Payload)
	case MsgRollDice:
		c.handleRollDice(msg.Payload)
	case MsgMoveToken:
		c.handleMoveToken(msg.Payload)
	case MsgStartCombat:
		c.handleStartCombat(msg.Payload)
	case MsgEndCombat:
		c.handleEndCombat(msg.Payload)
	case MsgNextTurn:
		c.handleNextTurn(msg.Payload)
	case MsgAttack:
		c.handleAttack(msg.Payload)
	case MsgGetState:
		c.handleGetState()
	default:
		c.sendError(fmt.Sprintf("Unknown message type: %s", msg.Type))
	}
}

// Handler methods

func (c *Client) handleJoinSession(payload map[string]interface{}) {
	sessionID, _ := payload["session_id"].(string)
	characterID, _ := payload["character_id"].(string)
	
	if sessionID == "" || characterID == "" {
		c.sendError("session_id and character_id are required")
		return
	}
	
	c.sessionID = sessionID
	c.characterID = characterID
	c.hub.register <- c
	
	// Send current state
	c.handleGetState()
}

func (c *Client) handlePlayerAction(payload map[string]interface{}) {
	action, _ := payload["action"].(string)
	if action == "" {
		c.sendError("action is required")
		return
	}
	
	result, err := c.hub.engine.ProcessAction(c.sessionID, c.characterID, action)
	if err != nil {
		c.sendError(err.Error())
		return
	}
	
	// Broadcast to all clients in session
	c.hub.BroadcastToSession(c.sessionID, Message{
		Type: MsgNarration,
		Payload: map[string]interface{}{
			"text":    result.Description,
			"actor":   c.characterID,
			"success": result.Success,
		},
		Timestamp: time.Now(),
	})
}

func (c *Client) handleRollDice(payload map[string]interface{}) {
	dice, _ := payload["dice"].(string)
	if dice == "" {
		dice = "d20"
	}
	
	result, err := c.hub.engine.RollDice(c.sessionID, c.characterID, dice)
	if err != nil {
		c.sendError(err.Error())
		return
	}
	
	c.hub.BroadcastToSession(c.sessionID, Message{
		Type: MsgDiceResult,
		Payload: map[string]interface{}{
			"dice":    result.Dice,
			"results": result.Results,
			"total":   result.Total,
			"actor":   c.characterID,
		},
		Timestamp: time.Now(),
	})
}

func (c *Client) handleMoveToken(payload map[string]interface{}) {
	x, _ := payload["x"].(float64)
	y, _ := payload["y"].(float64)
	
	err := c.hub.engine.MoveToken(c.sessionID, c.characterID, int(x), int(y))
	if err != nil {
		c.sendError(err.Error())
		return
	}
	
	c.broadcastStateUpdate()
}

func (c *Client) handleStartCombat(payload map[string]interface{}) {
	// TODO: Parse enemies from payload
	enemies := []domain.PlayerState{}
	
	err := c.hub.engine.StartCombat(c.sessionID, enemies)
	if err != nil {
		c.sendError(err.Error())
		return
	}
	
	c.broadcastStateUpdate()
}

func (c *Client) handleEndCombat(payload map[string]interface{}) {
	err := c.hub.engine.EndCombat(c.sessionID)
	if err != nil {
		c.sendError(err.Error())
		return
	}
	
	c.broadcastStateUpdate()
}

func (c *Client) handleNextTurn(payload map[string]interface{}) {
	err := c.hub.engine.NextTurn(c.sessionID)
	if err != nil {
		c.sendError(err.Error())
		return
	}
	
	c.broadcastStateUpdate()
}

func (c *Client) handleAttack(payload map[string]interface{}) {
	targetID, _ := payload["target_id"].(string)
	attackRoll, _ := payload["attack_roll"].(float64)
	
	if targetID == "" {
		c.sendError("target_id is required")
		return
	}
	
	result, err := c.hub.engine.Attack(c.sessionID, c.characterID, targetID, int(attackRoll))
	if err != nil {
		c.sendError(err.Error())
		return
	}
	
	c.hub.BroadcastToSession(c.sessionID, Message{
		Type: MsgEvent,
		Payload: map[string]interface{}{
			"type":     "attack",
			"actor":    c.characterID,
			"target":   targetID,
			"hit":      result.Hit,
			"damage":   result.Damage,
			"critical": result.CriticalHit,
		},
		Timestamp: time.Now(),
	})
	
	c.broadcastStateUpdate()
}

func (c *Client) handleGetState() {
	if c.sessionID == "" {
		return
	}
	
	state, err := c.hub.engine.GetState(c.sessionID)
	if err != nil {
		c.sendError(err.Error())
		return
	}
	
	stateJSON, _ := json.Marshal(state)
	var stateMap map[string]interface{}
	json.Unmarshal(stateJSON, &stateMap)
	
	c.send <- Message{
		Type:      MsgStateUpdate,
		Payload:   stateMap,
		Timestamp: time.Now(),
	}
}

func (c *Client) broadcastStateUpdate() {
	state, err := c.hub.engine.GetState(c.sessionID)
	if err != nil {
		return
	}
	
	stateJSON, _ := json.Marshal(state)
	var stateMap map[string]interface{}
	json.Unmarshal(stateJSON, &stateMap)
	
	c.hub.BroadcastToSession(c.sessionID, Message{
		Type:      MsgStateUpdate,
		Payload:   stateMap,
		Timestamp: time.Now(),
	})
}

func (c *Client) sendError(message string) {
	c.send <- Message{
		Type: MsgError,
		Payload: map[string]interface{}{
			"message": message,
		},
		Timestamp: time.Now(),
	}
}
