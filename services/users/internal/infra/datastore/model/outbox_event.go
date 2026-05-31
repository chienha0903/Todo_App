package model

import "time"

type OutboxEvent struct {
	ID            int64      `gorm:"column:id;primaryKey;autoIncrement"`
	EventID       string     `gorm:"column:event_id;not null;uniqueIndex;type:uuid"`
	AggregateType string     `gorm:"column:aggregate_type;not null"`
	AggregateID   int64      `gorm:"column:aggregate_id;not null"`
	EventType     string     `gorm:"column:event_type;not null"`
	Payload       []byte     `gorm:"column:payload;not null;type:jsonb"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime"`
	PublishedAt   *time.Time `gorm:"column:published_at"`
	RetryCount    int        `gorm:"column:retry_count;not null;default:0"`
	LastError     *string    `gorm:"column:last_error"`
}

func (OutboxEvent) TableName() string {
	return "outbox_events"
}
