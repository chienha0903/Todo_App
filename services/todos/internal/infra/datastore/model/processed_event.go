package model

import "time"

type ProcessedEvent struct {
	EventID     string    `gorm:"column:event_id;primaryKey"`
	ProcessedAt time.Time `gorm:"column:processed_at;autoCreateTime"`
}

func (ProcessedEvent) TableName() string {
	return "processed_events"
}
