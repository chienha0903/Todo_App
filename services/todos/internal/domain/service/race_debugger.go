package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/chienha0903/Todo_App/services/todos/internal/domain/gateway"
)

const raceSleep = 5 * time.Second

// RaceResult kết quả trả về từ một lần chạy race demo.
type RaceResult struct {
	Pod            string `json:"pod"`
	TodoID         int64  `json:"todo_id"`
	WithLock       bool   `json:"with_lock"`
	ReadCount      int    `json:"read_count"`      // completed_count đọc được trước sleep
	WrittenCount   int    `json:"written_count"`   // giá trị ghi vào DB
}

// RaceDebugger thực hiện pattern read-modify-write để demo lost update.
// Không dùng trong production.
type RaceDebugger struct {
	gw         gateway.RaceDemoGateway
	transactor gateway.TransactionGateway
	pod        string
}

func NewRaceDebugger(gw gateway.RaceDemoGateway, transactor gateway.TransactionGateway) *RaceDebugger {
	hostname, _ := os.Hostname()
	return &RaceDebugger{gw: gw, transactor: transactor, pod: hostname}
}

// RunNoLock: không dùng SELECT FOR UPDATE.
// Hai pod đọc cùng count=N, cả hai ghi N+1 → lost update, kết quả N+1 thay vì N+2.
func (s *RaceDebugger) RunNoLock(ctx context.Context, id int64) (*RaceResult, error) {
	return s.run(ctx, id, false)
}

// RunWithLock: dùng SELECT FOR UPDATE.
// Pod sau BLOCK đến khi pod trước COMMIT, đọc lại giá trị mới, ghi đúng.
func (s *RaceDebugger) RunWithLock(ctx context.Context, id int64) (*RaceResult, error) {
	return s.run(ctx, id, true)
}

func (s *RaceDebugger) run(ctx context.Context, id int64, withLock bool) (*RaceResult, error) {
	result := &RaceResult{Pod: s.pod, TodoID: id, WithLock: withLock}
	pod := s.pod

	err := s.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		slog.Info(fmt.Sprintf("[POD=%s] BEGIN tx", pod), "id", id, "lock", withLock)

		var snap *gateway.RaceSnapshot
		var err error
		if withLock {
			snap, err = s.gw.GetSnapshotForUpdate(ctx, id)
		} else {
			snap, err = s.gw.GetSnapshot(ctx, id)
		}
		if err != nil {
			return err
		}

		// Đọc giá trị vào bộ nhớ app — đây là bước tạo ra race condition.
		readCount := snap.CompletedCount
		newCount := readCount + 1
		result.ReadCount = readCount
		result.WrittenCount = newCount

		slog.Info(fmt.Sprintf("[POD=%s] READ", pod),
			"id", id,
			"status", snap.Status,
			"completed_count", readCount,
			"will_write", newCount,
		)

		slog.Info(fmt.Sprintf("[POD=%s] SLEEP 5s", pod), "id", id)
		time.Sleep(raceSleep)

		// Ghi absolute value — khi không có lock, 2 pod cùng ghi N+1 = lost update.
		if err := s.gw.SetCompletedCount(ctx, id, newCount); err != nil {
			return err
		}

		slog.Info(fmt.Sprintf("[POD=%s] UPDATE completed_count=%d & COMMIT", pod, newCount), "id", id)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
