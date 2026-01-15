package streaming

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development
		// TODO: Restrict in production
		return true
	},
}

// Client represents a connected WebSocket client
type Client struct {
	hub        *Hub
	conn       *websocket.Conn
	send       chan *StreamEvent
	workflowID string
	sequence   int64
}

// Hub manages all WebSocket connections and broadcasts
type Hub struct {
	// Registered clients by workflow ID
	clients map[string]map[*Client]bool

	// Inbound events to broadcast
	broadcast chan *BroadcastMessage

	// Register/unregister channels
	register   chan *Client
	unregister chan *Client

	mu sync.RWMutex
}

// BroadcastMessage is a message to broadcast to clients of a specific workflow
type BroadcastMessage struct {
	WorkflowID string
	Event      *StreamEvent
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		broadcast:  make(chan *BroadcastMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[client.workflowID]; !ok {
				h.clients[client.workflowID] = make(map[*Client]bool)
			}
			h.clients[client.workflowID][client] = true
			h.mu.Unlock()

			log.Printf("Client connected to workflow %s", client.workflowID)

			// Send connected event
			client.send <- NewEvent(EventConnected, map[string]interface{}{
				"workflow_id": client.workflowID,
			})

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.workflowID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.clients, client.workflowID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("Client disconnected from workflow %s", client.workflowID)

		case message := <-h.broadcast:
			h.mu.RLock()
			clients := h.clients[message.WorkflowID]
			h.mu.RUnlock()

			for client := range clients {
				// Assign sequence number
				message.Event.Sequence = atomic.AddInt64(&client.sequence, 1)

				select {
				case client.send <- message.Event:
				default:
					// Client too slow, disconnect
					h.mu.Lock()
					delete(h.clients[message.WorkflowID], client)
					close(client.send)
					h.mu.Unlock()
				}
			}
		}
	}
}

// Broadcast sends an event to all clients of a workflow
func (h *Hub) Broadcast(workflowID string, event *StreamEvent) {
	h.broadcast <- &BroadcastMessage{
		WorkflowID: workflowID,
		Event:      event,
	}
}

// BroadcastPhaseChange broadcasts a phase change event
func (h *Hub) BroadcastPhaseChange(workflowID, previousPhase, currentPhase, message string) {
	h.Broadcast(workflowID, NewEvent(EventPhaseChange, map[string]interface{}{
		"previous_phase": previousPhase,
		"current_phase":  currentPhase,
		"message":        message,
	}))
}

// BroadcastCodeChunk broadcasts a code chunk event
func (h *Hub) BroadcastCodeChunk(workflowID, filePath string, chunkIndex int, content string, isComplete bool) {
	h.Broadcast(workflowID, NewEvent(EventCodeChunk, map[string]interface{}{
		"file_path":   filePath,
		"chunk_index": chunkIndex,
		"content":     content,
		"is_complete": isComplete,
	}))
}

// BroadcastChatMessage broadcasts a chat message event
func (h *Hub) BroadcastChatMessage(workflowID, role, agent, content string) {
	h.Broadcast(workflowID, NewEvent(EventChatMessage, map[string]interface{}{
		"role":    role,
		"agent":   agent,
		"content": content,
	}))
}

// BroadcastFileStart broadcasts a file start event
func (h *Hub) BroadcastFileStart(workflowID, filePath, language string) {
	h.Broadcast(workflowID, NewEvent(EventFileStart, map[string]interface{}{
		"file_path": filePath,
		"language":  language,
	}))
}

// BroadcastFileEnd broadcasts a file end event
func (h *Hub) BroadcastFileEnd(workflowID, filePath string) {
	h.Broadcast(workflowID, NewEvent(EventFileEnd, map[string]interface{}{
		"file_path": filePath,
	}))
}

// BroadcastError broadcasts an error event
func (h *Hub) BroadcastError(workflowID, code, message, details string) {
	h.Broadcast(workflowID, NewEvent(EventError, map[string]interface{}{
		"code":    code,
		"message": message,
		"details": details,
	}))
}

// ServeWs handles WebSocket upgrade and connection
func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request, workflowID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		hub:        h,
		conn:       conn,
		send:       make(chan *StreamEvent, 256),
		workflowID: workflowID,
		sequence:   0,
	}

	h.register <- client

	// Start read/write pumps
	go client.writePump()
	go client.readPump()
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second) // Ping interval
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case event, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			data, err := json.Marshal(event)
			if err != nil {
				log.Printf("JSON marshal error: %v", err)
				continue
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512 * 1024) // 512KB max message size
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming messages (e.g., user commands)
		log.Printf("Received message from client: %s", message)
		// TODO: Process client messages if needed
	}
}

// GetClientCount returns the number of connected clients for a workflow
func (h *Hub) GetClientCount(workflowID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[workflowID])
}
