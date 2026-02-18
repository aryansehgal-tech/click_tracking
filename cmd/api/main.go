package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"click_tracking/internal/db"
	"click_tracking/internal/models"

	redisclient "click_tracking/internal/redis"
	"click_tracking/internal/event"
)

func main() {
	// Load .env FIRST
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize Postgres
	database, err := db.NewPostgres()
	if err != nil {
		panic(err)
	}

	// Initialize Redis
	rdb, err := redisclient.NewRedis()
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
}


	// Optional: verify Redis connection
	if err := rdb.Ping(redisclient.Ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Create event handler
	eventHandler := event.NewHandler(rdb)

	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Create session
	router.POST("/session", func(c *gin.Context) {
		var req struct {
			CampaignID string `json:"campaign_id" binding:"required,uuid"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})
			return
		}

		campaignUUID, err := uuid.Parse(req.CampaignID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid campaign_id",
			})
			return
		}

		session := models.Session{
			ID:         uuid.New(),
			CampaignID: campaignUUID,
			CreatedAt:  time.Now().UTC(),
		}

		if err := database.Create(&session).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create session",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"session_id": session.ID,
		})
	})

	router.POST("/event", eventHandler.IngestEvent)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router.Run(":" + port)
}
