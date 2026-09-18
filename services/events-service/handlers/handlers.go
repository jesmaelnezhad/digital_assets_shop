package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/store4bots/shared/middleware"
)

type Handler struct {
	store Store
}

func New(store Store) *Handler {
	return &Handler{store: store}
}

type ingestRequest struct {
	Name        string                 `json:"name" binding:"required"`
	SessionID   string                 `json:"session_id"`
	Path        string                 `json:"path"`
	Properties  map[string]interface{} `json:"properties"`
	OccurredAt  string                 `json:"occurred_at"`
}

func (h *Handler) Ingest(c *gin.Context) {
	var req ingestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	name := strings.TrimSpace(req.Name)
	if !AllowedEvent(name) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown event"})
		return
	}
	ttl, err := h.store.GetTTL(c.Request.Context())
	if err != nil {
		ttl = DefaultTTLSeconds
	}
	now := time.Now().UTC()
	occurred := now
	if req.OccurredAt != "" {
		if t, err := time.Parse(time.RFC3339, req.OccurredAt); err == nil {
			occurred = t.UTC()
		}
	}
	userID, _ := middleware.GetUserIDFromContext(c)
	ev := Event{
		Name:       name,
		SessionID:  strings.TrimSpace(req.SessionID),
		UserID:     userID,
		Path:       strings.TrimSpace(req.Path),
		Properties: req.Properties,
		OccurredAt: occurred,
		ExpireAt:   now.Add(time.Duration(ttl) * time.Second),
		ReceivedAt: now,
	}
	if ev.Properties == nil {
		ev.Properties = map[string]interface{}{}
	}
	if err := h.store.Insert(c.Request.Context(), ev); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store event"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"ok":        true,
		"name":      ev.Name,
		"expire_at": ev.ExpireAt,
	})
}

func (h *Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	name := strings.TrimSpace(c.Query("name"))
	rows, err := h.store.List(c.Request.Context(), name, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	if rows == nil {
		rows = []Event{}
	}
	n, _ := h.store.Count(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"events": rows, "total": n})
}

func (h *Handler) GetTTL(c *gin.Context) {
	ttl, err := h.store.GetTTL(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"seconds": ttl, "hours": float64(ttl) / 3600})
}

func (h *Handler) SetTTL(c *gin.Context) {
	var req struct {
		Seconds int     `json:"seconds"`
		Hours   float64 `json:"hours"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	seconds := req.Seconds
	if seconds == 0 && req.Hours > 0 {
		seconds = int(req.Hours * 3600)
	}
	seconds = ClampTTL(seconds)
	if err := h.store.SetTTL(c.Request.Context(), seconds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save ttl"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"seconds": seconds, "hours": float64(seconds) / 3600})
}
