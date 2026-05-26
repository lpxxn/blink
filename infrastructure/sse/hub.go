package sse

import "sync"

// Event is the payload pushed to SSE clients.
type Event struct {
	Name string // SSE event name (e.g. "notification")
	Data string // JSON payload
}

// Hub manages per-user SSE subscriber channels.
type Hub struct {
	mu   sync.RWMutex
	subs map[int64]map[chan Event]struct{}
}

func NewHub() *Hub {
	return &Hub{subs: make(map[int64]map[chan Event]struct{})}
}

// Subscribe returns a channel the caller should read from. Call Unsubscribe when done.
func (h *Hub) Subscribe(userID int64) chan Event {
	ch := make(chan Event, 8)
	h.mu.Lock()
	if h.subs[userID] == nil {
		h.subs[userID] = make(map[chan Event]struct{})
	}
	h.subs[userID][ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

// Unsubscribe removes a channel and closes it.
func (h *Hub) Unsubscribe(userID int64, ch chan Event) {
	h.mu.Lock()
	if m, ok := h.subs[userID]; ok {
		delete(m, ch)
		if len(m) == 0 {
			delete(h.subs, userID)
		}
	}
	h.mu.Unlock()
	// Drain to prevent goroutine leak if sender wrote concurrently.
	for range ch {
	}
}

// Publish sends an event to all subscribers for the given user. Non-blocking per channel.
func (h *Hub) Publish(userID int64, evt Event) {
	h.mu.RLock()
	chans := h.subs[userID]
	h.mu.RUnlock()

	for ch := range chans {
		select {
		case ch <- evt:
		default:
			// Slow consumer — drop rather than block the writer.
		}
	}
}
