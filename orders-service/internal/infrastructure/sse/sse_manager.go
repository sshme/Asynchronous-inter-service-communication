package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"orders-service/internal/domain/orders"
	"sync"
	"time"
)

type OrderStatusEvent struct {
	ID          string             `json:"id"`
	UserID      string             `json:"userID"`
	Amount      float64            `json:"amount"`
	Currency    string             `json:"currency"`
	Status      orders.OrderStatus `json:"status"`
	PaymentID   string             `json:"paymentID"`
	ErrorReason string             `json:"errorReason"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
}

type Client struct {
	ID     string
	UserID string
	Events chan OrderStatusEvent
	Done   chan bool
}

type Manager struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan OrderStatusEvent
	mutex      sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan OrderStatusEvent, 100),
	}
}

func (m *Manager) Start(ctx context.Context) {
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

			case event := <-m.broadcast:
				m.mutex.RLock()
				for _, client := range m.clients {
					if client.UserID == event.UserID {
						select {
						case client.Events <- event:
						case <-time.After(time.Second * 5):
							log.Printf("Removing unresponsive SSE client: %s", client.ID)
							go func(c *Client) {
								m.unregister <- c
							}(client)
						}
					}
				}
				m.mutex.RUnlock()

			case <-ctx.Done():
				m.mutex.Lock()
				for _, client := range m.clients {
					close(client.Events)
					close(client.Done)
				}
				m.clients = make(map[string]*Client)
				m.mutex.Unlock()
				return
			}
		}
	}()
}

func (m *Manager) RegisterClient(clientID, userID string) *Client {
	client := &Client{
		ID:     clientID,
		UserID: userID,
		Events: make(chan OrderStatusEvent, 10),
		Done:   make(chan bool),
	}

	m.register <- client
	return client
}

func (m *Manager) UnregisterClient(client *Client) {
	m.unregister <- client
}

func (m *Manager) BroadcastOrderUpdate(order *orders.Order) {
	event := OrderStatusEvent{
		ID:          order.ID,
		UserID:      order.UserID,
		Amount:      order.Amount,
		Currency:    order.Currency,
		Status:      order.Status,
		PaymentID:   order.PaymentID,
		ErrorReason: order.ErrorReason,
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
	}

	select {
	case m.broadcast <- event:
	case <-time.After(time.Second):
		log.Printf("Failed to broadcast order update due to timeout: OrderID=%s", order.ID)
	}
}

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
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	clientID := fmt.Sprintf("%s-%d", userID, time.Now().UnixNano())

	client := m.RegisterClient(clientID, userID)
	defer m.UnregisterClient(client)

	fmt.Fprintf(w, "event: connected\n")
	fmt.Fprintf(w, "data: {\"message\": \"Connected to order status updates\", \"user_id\": \"%s\"}\n\n", userID)

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	for {
		select {
		case event := <-client.Events:
			log.Printf("Sending SSE event to client %s: %+v", client.ID, event)

			eventData, err := json.Marshal(event)
			if err != nil {
				log.Printf("Failed to marshal SSE event: %v", err)
				continue
			}

			fmt.Fprintf(w, "event: order-update\n")
			fmt.Fprintf(w, "data: %s\n\n", eventData)

			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

		case <-client.Done:
			return

		case <-r.Context().Done():
			return
		}
	}
}
