package sse

import (
	"sync"
	"time"

	"github.com/hycat-team/littlecafe-backend/shared/config"
)

// ISSEManager defines the interface for the Server-Sent Events manager.
type ISSEManager interface {
	Subscribe(streamId string) chan []byte
	Unsubscribe(streamId string, ch chan []byte)
	Publish(streamId string, data []byte)
}

// subscriberBufferSize is the per-subscriber channel capacity for live events.
const subscriberBufferSize = 10

// streamBuffer keeps the most recent events of one stream so a client that
// reconnects can replay them.
type streamBuffer struct {
	events  [][]byte
	expires time.Time
}

// SSEManager implements ISSEManager to manage thread-safe event subscriptions
// plus a per-stream replay buffer.
type SSEManager struct {
	mu       sync.RWMutex
	channels map[string][]chan []byte // streamId -> subscriber channels
	buffers  map[string]*streamBuffer // streamId -> replay buffer
	maxSize  int
	ttl      time.Duration
	now      func() time.Time
}

// NewSSEManager creates a manager with the configured replay buffer size and
// TTL (defaults 100 events / 300s when config is missing or non-positive).
func NewSSEManager(cfg *config.Config) *SSEManager {
	maxSize := 100
	ttlSeconds := 300
	if cfg != nil {
		if cfg.Server.SSE.ReplayBufferSize > 0 {
			maxSize = cfg.Server.SSE.ReplayBufferSize
		}
		if cfg.Server.SSE.ReplayBufferTTLSeconds > 0 {
			ttlSeconds = cfg.Server.SSE.ReplayBufferTTLSeconds
		}
	}
	return &SSEManager{
		channels: make(map[string][]chan []byte),
		buffers:  make(map[string]*streamBuffer),
		maxSize:  maxSize,
		ttl:      time.Duration(ttlSeconds) * time.Second,
		now:      time.Now,
	}
}

// Subscribe registers a new subscriber channel for a streamId and immediately
// replays the buffered events (FIFO) so late reconnects do not miss updates.
func (m *SSEManager) Subscribe(streamId string) chan []byte {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan []byte, m.subscriberCap())
	if buf := m.activeBufferLocked(streamId); buf != nil {
		for _, ev := range buf.events {
			select {
			case ch <- ev:
			default:
			}
		}
	}
	m.channels[streamId] = append(m.channels[streamId], ch)
	IncRealtimeSubscribers()
	return ch
}

// Unsubscribe closes and removes a subscriber channel for a streamId. The replay
// buffer is intentionally kept so a later reconnect can still replay.
func (m *SSEManager) Unsubscribe(streamId string, ch chan []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	channels, exists := m.channels[streamId]
	if !exists {
		return
	}

	for i, c := range channels {
		if c == ch {
			close(c)
			m.channels[streamId] = append(channels[:i], channels[i+1:]...)
			DecRealtimeSubscribers()
			break
		}
	}

	if len(m.channels[streamId]) == 0 {
		delete(m.channels, streamId)
	}
}

// Publish buffers data for replay and broadcasts it to all active subscriber
// channels for a streamId.
func (m *SSEManager) Publish(streamId string, data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	buf := m.activeBufferLocked(streamId)
	if buf == nil {
		buf = &streamBuffer{}
		m.buffers[streamId] = buf
	}
	buf.events = append(buf.events, data)
	if m.maxSize > 0 && len(buf.events) > m.maxSize {
		buf.events = buf.events[len(buf.events)-m.maxSize:]
	}
	buf.expires = m.now().Add(m.ttl)

	channels, exists := m.channels[streamId]
	if !exists {
		return
	}
	for _, ch := range channels {
		select {
		case ch <- data:
		default:
			// Avoid blocking if the receiver is slow or disconnected.
		}
	}
}

// activeBufferLocked returns the live buffer for streamId, lazily evicting it when
// its TTL has elapsed. Callers MUST hold m.mu.
func (m *SSEManager) activeBufferLocked(streamId string) *streamBuffer {
	buf, ok := m.buffers[streamId]
	if !ok {
		return nil
	}
	if m.now().After(buf.expires) {
		delete(m.buffers, streamId)
		return nil
	}
	return buf
}

// subscriberCap sizes a subscriber channel so a full replay fits without drops.
func (m *SSEManager) subscriberCap() int {
	if m.maxSize > subscriberBufferSize {
		return m.maxSize
	}
	return subscriberBufferSize
}
