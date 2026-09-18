package handlers

import (
	"context"
	"sync"
	"time"
)

const (
	DefaultTTLSeconds = 3600
	MinTTLSeconds     = 60
	MaxTTLSeconds     = 30 * 24 * 3600
	EventProductView  = "product_view"
	EventCheckoutClick = "checkout_click"
)

var allowedEvents = map[string]bool{
	EventProductView:   true,
	EventCheckoutClick: true,
}

func AllowedEvent(name string) bool {
	return allowedEvents[name]
}

func ClampTTL(seconds int) int {
	if seconds <= 0 {
		return DefaultTTLSeconds
	}
	if seconds < MinTTLSeconds {
		return MinTTLSeconds
	}
	if seconds > MaxTTLSeconds {
		return MaxTTLSeconds
	}
	return seconds
}

type Event struct {
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name"`
	SessionID   string                 `json:"session_id,omitempty"`
	UserID      int                    `json:"user_id,omitempty"`
	Path        string                 `json:"path,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	OccurredAt  time.Time              `json:"occurred_at"`
	ExpireAt    time.Time              `json:"expire_at"`
	ReceivedAt  time.Time              `json:"received_at"`
}

type Store interface {
	Insert(ctx context.Context, ev Event) error
	List(ctx context.Context, name string, limit int) ([]Event, error)
	Count(ctx context.Context) (int64, error)
	GetTTL(ctx context.Context) (int, error)
	SetTTL(ctx context.Context, seconds int) error
}

type MemoryStore struct {
	mu     sync.Mutex
	ttl    int
	events []Event
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{ttl: DefaultTTLSeconds}
}

func (s *MemoryStore) Insert(_ context.Context, ev Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append([]Event{ev}, s.events...)
	if len(s.events) > 500 {
		s.events = s.events[:500]
	}
	return nil
}

func (s *MemoryStore) List(_ context.Context, name string, limit int) ([]Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if limit < 1 {
		limit = 20
	}
	out := []Event{}
	now := time.Now().UTC()
	for _, ev := range s.events {
		if ev.ExpireAt.Before(now) {
			continue
		}
		if name != "" && ev.Name != name {
			continue
		}
		out = append(out, ev)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (s *MemoryStore) Count(_ context.Context) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	var n int64
	for _, ev := range s.events {
		if ev.ExpireAt.After(now) {
			n++
		}
	}
	return n, nil
}

func (s *MemoryStore) GetTTL(_ context.Context) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ttl == 0 {
		return DefaultTTLSeconds, nil
	}
	return s.ttl, nil
}

func (s *MemoryStore) SetTTL(_ context.Context, seconds int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ttl = ClampTTL(seconds)
	return nil
}
