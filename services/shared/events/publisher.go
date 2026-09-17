// Package events defines the event publisher interface and implementations
// for inter-service communication in the Pawradise microservice architecture.
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// Event represents a domain event published between services.
type Event struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
	Service   string      `json:"service"`
}

// EventOption configures an Event before publishing.
type EventOption func(*Event)

// WithService sets the service name on an event.
func WithService(s string) EventOption {
	return func(e *Event) { e.Service = s }
}

// NewEvent creates a new Event with the given type and payload.
func NewEvent(eventType string, payload interface{}, opts ...EventOption) Event {
	e := Event{
		ID:        generateID(),
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
		Service:   os.Getenv("SERVICE_NAME"),
	}
	for _, opt := range opts {
		opt(&e)
	}
	return e
}

// EventPublisher defines the interface for publishing events.
type EventPublisher interface {
	Publish(ctx context.Context, topic string, event Event) error
	Close() error
}

// PostgresEventPublisher publishes events to a PostgreSQL NOTIFY channel
// and an events table for persistence.
type PostgresEventPublisher struct {
	db     Execer
	logger *log.Logger
}

// Execer is the minimal interface needed for event persistence.
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (Result, error)
}

// Result is the minimal interface needed from database/sql.
type Result interface {
	RowsAffected() (int64, error)
}

// NewPostgresEventPublisher creates a Postgres-backed event publisher.
func NewPostgresEventPublisher(db Execer) *PostgresEventPublisher {
	return &PostgresEventPublisher{
		db:     db,
		logger: log.New(os.Stdout, "[events:postgres] ", log.LstdFlags),
	}
}

// Publish sends an event via PostgreSQL NOTIFY and inserts into events table.
func (p *PostgresEventPublisher) Publish(ctx context.Context, topic string, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Insert into events table
	_, err = p.db.ExecContext(ctx,
		`INSERT INTO events (id, type, payload, service, created_at) VALUES ($1, $2, $3, $4, $5)`,
		event.ID, event.Type, string(data), event.Service, event.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	// Also send NOTIFY for real-time listeners
	_, _ = p.db.ExecContext(ctx, fmt.Sprintf("NOTIFY %s, '%s'", topic, sanitizeForNotify(string(data))))

	p.logger.Printf("Published event %s to channel %s", event.Type, topic)
	return nil
}

// Close is a no-op for PostgresEventPublisher (connection managed externally).
func (p *PostgresEventPublisher) Close() error {
	return nil
}

// =========================================================================
// helpers
// =========================================================================

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return fallback
}

func generateID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
}

func sanitizeForNotify(s string) string {
	result := ""
	for _, c := range s {
		if c == '\'' {
			result += "\'\'"
		} else {
			result += string(c)
		}
	}
	return result
}
