package notification

import (
	"encoding/json"
	"log/slog"
	"sync"
)

// SSEEvent is published through the broker to SSE subscribers.
type SSEEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// MarshalJSON serializes the event for SSE transmission.
func (e SSEEvent) MarshalJSON() ([]byte, error) {
	// Use a map so we can serialize the Data field dynamically
	m := map[string]interface{}{
		"type": e.Type,
		"data": e.Data,
	}
	return json.Marshal(m)
}

// Broker is an in-memory pub/sub hub for per-user notification events.
// Each user has a set of subscriber channels (one per connected client tab).
type Broker struct {
	mu          sync.RWMutex
	subscribers map[string][]chan SSEEvent
}

// NewBroker creates a new Broker.
func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[string][]chan SSEEvent),
	}
}

// Subscribe registers a new channel for userID. The caller must drain or
// unbuffer — using a buffered channel (buffer=64) to avoid blocking
// publishers when a client is slow.
func (b *Broker) Subscribe(userID string) chan SSEEvent {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan SSEEvent, 64)
	b.subscribers[userID] = append(b.subscribers[userID], ch)
	slog.Debug("sse subscriber joined", "userID", userID)
	return ch
}

// Unsubscribe removes the given channel from userID's subscriber list.
func (b *Broker) Unsubscribe(userID string, ch chan SSEEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subs := b.subscribers[userID]
	for i, sub := range subs {
		if sub == ch {
			b.subscribers[userID] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
	if len(b.subscribers[userID]) == 0 {
		delete(b.subscribers, userID)
	}
	slog.Debug("sse subscriber left", "userID", userID)
}

// Publish sends an event to all channels subscribed to userID.
// Drops the event if a channel's buffer is full (non-blocking).
func (b *Broker) Publish(userID string, event SSEEvent) {
	b.mu.RLock()
	subs := b.subscribers[userID]
	b.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- event:
		default:
			slog.Warn("sse subscriber channel full, dropping event", "userID", userID)
		}
	}
}