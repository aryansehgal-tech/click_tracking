package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Event struct {
	ID 	  uuid.UUID           `gorm:"type:uuid;primaryKey"`
	SessionID  uuid.UUID	  `gorm:"type:uuid;index;not null"`
	CampaignID uuid.UUID	  `gorm:"type:uuid;index;not null"`
	EventType  string         `gorm:"type:varchar(50);index;not null"`
	EventTime time.Time	  	  `gorm:"index;not null"`
	MetaData datatypes.JSON   `gorm:"type:jsonb"`
	CreatedAt time.Time 	  `gorm:"not null"`
}