package event

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	redisclient "click_tracking/internal/redis"
)

const EventStream = "events_stream"

// Allowed event types
var allowedEventTypes = map[string]bool{
	"page_view":         true,
	"product_view":      true,
	"add_to_cart":       true,
	"checkout_started":  true,
	"purchase_completed": true,
}

type Handler struct {
	Redis *redis.Client
}

func NewHandler(rdb *redis.Client) *Handler {
	return &Handler{Redis: rdb}
}

func (h *Handler) IngestEvent(c *gin.Context) {
	var req struct {
		SessionID  string                 `json:"session_id" binding:"required,uuid"`
		CampaignID string                 `json:"campaign_id" binding:"required,uuid"`
		EventType  string                 `json:"event_type" binding:"required"`
		Metadata   map[string]interface{} `json:"metadata"`
	}

	// 1️⃣ Validate JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 2️⃣ Validate UUID parsing (extra safety)
	if _, err := uuid.Parse(req.SessionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
		return
	}

	if _, err := uuid.Parse(req.CampaignID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign_id"})
		return
	}

	// 3️⃣ Validate event type
	if !allowedEventTypes[req.EventType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event_type"})
		return
	}

	eventID := uuid.New().String()
	now := time.Now().UTC()

	// 4️⃣ Serialize metadata properly
	metadataBytes, err := json.Marshal(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid metadata"})
		return
	}

	// 5️⃣ Push to Redis Stream (with max length cap)
	err = h.Redis.XAdd(redisclient.Ctx, &redis.XAddArgs{
		Stream:       EventStream,
		MaxLen:       100000, // prevents infinite growth
		Approx:       true,   // faster trimming
		Values: map[string]interface{}{
			"event_id":    eventID,
			"session_id":  req.SessionID,
			"campaign_id": req.CampaignID,
			"event_type":  req.EventType,
			"event_time":  now.Format(time.RFC3339Nano),
			"metadata":    string(metadataBytes),
		},
	}).Err()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue event"})
		return
	}

	// 6️⃣ Real-time counters (fire and forget)
	counterKey := "campaign:" + req.CampaignID + ":events:" + req.EventType
	h.Redis.Incr(redisclient.Ctx, counterKey)

	if req.EventType == "purchase_completed" {
		h.Redis.Incr(redisclient.Ctx, "campaign:"+req.CampaignID+":conversions")
	}

	c.JSON(http.StatusAccepted, gin.H{
		"event_id": eventID,
	})
}
