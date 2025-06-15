package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"orders-service/internal/domain/dto"
	"orders-service/internal/infrastructure/pubsub/redis"
	"sync"
	"time"
)

// Client represents a single SSE client connection.
type Client struct {
	ID     string
	UserID string
	Events chan *dto.SSEMessage
	Done   chan bool
}

// Manager handles all SSE client connections.
type Manager struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	subscriber *redis.Subscriber
	mutex      sync.RWMutex
}

// NewManager creates a new SSE Manager.
func NewManager(subscriber *redis.Subscriber) *Manager {
	return &Manager{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		subscriber: subscriber,
	}
}

// handleRedisMessage is the handler for messages received from the Redis subscriber.
func (m *Manager) handleRedisMessage(message *dto.SSEMessage) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, client := range m.clients {
		if client.UserID == message.UserID {
			select {
			case client.Events <- message:
			case <-time.After(time.Second):
				log.Printf("SSE client %s event channel is full. Skipping message.", client.ID)
			}
		}
	}
}

// Start begins the SSE manager's event loop.
func (m *Manager) Start(ctx context.Context) {
	go m.subscriber.Subscribe(ctx, m.handleRedisMessage)

	go func() {
		for {
			select {
			case client := <-m.register:
				m.mutex.Lock()
				m.clients[client.ID] = client
				m.mutex.Unlock()
				log.Printf("SSE client registered: %s (UserID: %s)", client.ID, client.UserID)

			case client := <-m.unregister:
				m.mutex.Lock()
				if _, ok := m.clients[client.ID]; ok {
					delete(m.clients, client.ID)
					close(client.Events)
					close(client.Done)
				}
				m.mutex.Unlock()
				log.Printf("SSE client unregistered: %s", client.ID)

			case <-ctx.Done():
				m.mutex.Lock()
				for _, client := range m.clients {
					close(client.Events)
					close(client.Done)
				}
				m.clients = make(map[string]*Client)
				m.mutex.Unlock()
				log.Println("SSE Manager shutting down.")
				return
			}
		}
	}()
}

// RegisterClient creates and registers a new SSE client.
func (m *Manager) RegisterClient(clientID, userID string) *Client {
	client := &Client{
		ID:     clientID,
		UserID: userID,
		Events: make(chan *dto.SSEMessage, 10),
		Done:   make(chan bool),
	}
	m.register <- client
	return client
}

// UnregisterClient unregisters an SSE client.
func (m *Manager) UnregisterClient(client *Client) {
	m.unregister <- client
}

// HandleSSE is the HTTP handler for new SSE connections.
func (m *Manager) HandleSSE(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id parameter is required", http.StatusBadRequest)
		return
	}

	log.Printf("New SSE connection for user: %s", userID)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	clientID := fmt.Sprintf("%s-%d", userID, time.Now().UnixNano())
	client := m.RegisterClient(clientID, userID)
	defer m.UnregisterClient(client)

	// Send a connected message
	connectedMsg := map[string]string{"message": "Connected to order status updates", "user_id": userID}
	connectedEvent := &dto.SSEMessage{UserID: userID, Event: "connected", Payload: connectedMsg}

	eventData, _ := json.Marshal(connectedEvent.Payload)
	fmt.Fprintf(w, "event: %s\n", connectedEvent.Event)
	fmt.Fprintf(w, "data: %s\n\n", eventData)
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	for {
		select {
		case msg := <-client.Events:
			log.Printf("Sending SSE event '%s' to client %s", msg.Event, client.ID)

			payloadData, err := json.Marshal(msg.Payload)
			if err != nil {
				log.Printf("Failed to marshal SSE payload: %v", err)
				continue
			}

			fmt.Fprintf(w, "event: %s\n", msg.Event)
			fmt.Fprintf(w, "data: %s\n\n", payloadData)

			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

		case <-client.Done:
			log.Printf("Client %s done.", client.ID)
			return

		case <-r.Context().Done():
			log.Printf("Client %s connection closed by remote.", client.ID)
			return
		}
	}
}
