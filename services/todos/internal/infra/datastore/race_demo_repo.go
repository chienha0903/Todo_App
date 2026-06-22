package datastore

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway"
)

type raceDemoRepo struct {
	db *gorm.DB
}

func NewRaceDemoRepo(db *gorm.DB) *raceDemoRepo {
	return &raceDemoRepo{db: db}
}

func NewRaceDemoGateway(r *raceDemoRepo) gateway.RaceDemoGateway {
	return r
}

// GetSnapshot đọc không dùng lock — hai pod có thể đọc cùng giá trị đồng thời.
func (r *raceDemoRepo) GetSnapshot(ctx context.Context, id int64) (*gateway.RaceSnapshot, error) {
	var snap gateway.RaceSnapshot
	err := extractDB(ctx, r.db).WithContext(ctx).
		Raw("SELECT id, status, completed_count FROM todos WHERE id = ? AND deleted_at IS NULL", id).
		Scan(&snap).Error
	if err != nil {
		return nil, fmt.Errorf("race: get snapshot: %w", err)
	}
	if snap.ID == 0 {
		return nil, fmt.Errorf("race: todo %d not found", id)
	}
	return &snap, nil
}

// GetSnapshotForUpdate dùng SELECT ... FOR UPDATE — pod sau sẽ block cho đến khi pod trước COMMIT.
func (r *raceDemoRepo) GetSnapshotForUpdate(ctx context.Context, id int64) (*gateway.RaceSnapshot, error) {
	var snap gateway.RaceSnapshot
	err := extractDB(ctx, r.db).WithContext(ctx).
		Raw("SELECT id, status, completed_count FROM todos WHERE id = ? AND deleted_at IS NULL FOR UPDATE", id).
		Scan(&snap).Error
	if err != nil {
		return nil, fmt.Errorf("race: get snapshot for update: %w", err)
	}
	if snap.ID == 0 {
		return nil, fmt.Errorf("race: todo %d not found", id)
	}
	return &snap, nil
}

// SetCompletedCount ghi absolute value — đây là pattern tạo ra lost update:
// cả 2 pod đọc cùng count=N, tính N+1 trong app, cả 2 ghi N+1 → chỉ còn N+1 chứ không phải N+2.
func (r *raceDemoRepo) SetCompletedCount(ctx context.Context, id int64, count int) error {
	err := extractDB(ctx, r.db).WithContext(ctx).
		Exec("UPDATE todos SET completed_count = ?, updated_at = NOW() WHERE id = ?", count, id).Error
	if err != nil {
		return fmt.Errorf("race: set completed_count: %w", err)
	}
	return nil
}
