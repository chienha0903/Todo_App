package gateway

import "context"

// RaceSnapshot là trạng thái todo tại thời điểm đọc trong race demo.
type RaceSnapshot struct {
	ID             int64  `json:"id"`
	Status         string `json:"status"`
	CompletedCount int    `json:"completed_count"`
}

// RaceDemoGateway là interface riêng cho debug — không dùng trong production flow.
type RaceDemoGateway interface {
	GetSnapshot(ctx context.Context, id int64) (*RaceSnapshot, error)
	GetSnapshotForUpdate(ctx context.Context, id int64) (*RaceSnapshot, error)
	SetCompletedCount(ctx context.Context, id int64, count int) error
}
