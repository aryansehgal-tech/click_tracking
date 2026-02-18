package main

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"click_tracking/internal/db"
	"click_tracking/internal/models"
	redisclient "click_tracking/internal/redis"
)

var ctx = context.Background()

func main() {
	// Load env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect Postgres
	database, err := db.NewPostgres()
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}

	// Connect Redis
	rdb, err := redisclient.NewRedis()
	if err != nil {
		log.Fatalf("Redis init failed: %v", err)
	}

	streamName := "events_stream"
	groupName := "event_consumers"
	consumerName := "worker-1"

	// Create consumer group (ignore if exists)
	err = rdb.XGroupCreateMkStream(ctx, streamName, groupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Fatalf("Failed to create consumer group: %v", err)
	}

	log.Println("Worker started...")

	for {
		streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    groupName,
			Consumer: consumerName,
			Streams:  []string{streamName, ">"},
			Count:    10,
			Block:    0,
		}).Result()

		if err != nil {
			log.Println("Read error:", err)
			continue
		}

		for _, stream := range streams {
			for _, message := range stream.Messages {

				// Parse values
				sessionIDStr := message.Values["session_id"].(string)
				campaignIDStr := message.Values["campaign_id"].(string)
				eventType := message.Values["event_type"].(string)

				sessionID, err := uuid.Parse(sessionIDStr)
				if err != nil {
					log.Println("Invalid session_id:", err)
					continue
				}

				campaignID, err := uuid.Parse(campaignIDStr)
				if err != nil {
					log.Println("Invalid campaign_id:", err)
					continue
				}

				event := models.Event{
					ID:         uuid.New(),
					SessionID:  sessionID,
					CampaignID: campaignID,
					EventType:  eventType,
					CreatedAt:  time.Now().UTC(),
				}

				// Save to Postgres
				if err := database.Create(&event).Error; err != nil {
					log.Println("DB insert failed:", err)
					continue
				}

				// ACK only after DB success
				if err := rdb.XAck(ctx, streamName, groupName, message.ID).Err(); err != nil {
					log.Println("ACK failed:", err)
				}

				log.Println("Processed and stored event:", message.ID)
			}
		}
	}
}
