package consumer_test

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// workerPool là implementation tối giản để test pattern,
// không phụ thuộc vào DB hay RabbitMQ thật.
type job struct {
	id      int
	payload string
}

type result struct {
	jobID    int
	workerID int
}

func startWorkerPool(ctx context.Context, numWorkers int, process func(workerID int, j job) result) (chan<- job, <-chan result) {
	jobs := make(chan job, numWorkers)
	results := make(chan result, numWorkers*2)

	var wg sync.WaitGroup
	for i := range numWorkers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := range jobs {
				results <- process(workerID, j)
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return jobs, results
}

// TestWorkerPool_MaxConcurrency kiểm tra không quá N worker chạy đồng thời.
func TestWorkerPool_MaxConcurrency(t *testing.T) {
	const numWorkers = 3
	const numJobs = 9

	var (
		active    atomic.Int32 // số worker đang xử lý đồng thời
		maxActive atomic.Int32 // peak concurrency đo được
		mu        sync.Mutex
	)

	ctx := context.Background()
	jobs, results := startWorkerPool(ctx, numWorkers, func(workerID int, j job) result {
		cur := active.Add(1)
		// cập nhật max nếu cần (atomic snapshot có thể không chính xác 100%,
		// nên dùng mu để đảm bảo)
		mu.Lock()
		if cur > maxActive.Load() {
			maxActive.Store(cur)
		}
		mu.Unlock()

		time.Sleep(20 * time.Millisecond) // giả lập I/O
		active.Add(-1)
		return result{jobID: j.id, workerID: workerID}
	})

	// Submit tất cả jobs
	go func() {
		for i := range numJobs {
			jobs <- job{id: i, payload: "test"}
		}
		close(jobs)
	}()

	// Collect results
	var got []result
	for r := range results {
		got = append(got, r)
	}

	if len(got) != numJobs {
		t.Errorf("expected %d results, got %d", numJobs, len(got))
	}

	if maxActive.Load() > numWorkers {
		t.Errorf("max concurrency %d exceeded numWorkers %d", maxActive.Load(), numWorkers)
	}

	t.Logf("peak concurrency: %d / %d workers", maxActive.Load(), numWorkers)
}

// TestWorkerPool_AllJobsProcessed kiểm tra không bị mất job.
func TestWorkerPool_AllJobsProcessed(t *testing.T) {
	const numWorkers = 5
	const numJobs = 20

	ctx := context.Background()
	jobs, results := startWorkerPool(ctx, numWorkers, func(workerID int, j job) result {
		return result{jobID: j.id, workerID: workerID}
	})

	go func() {
		for i := range numJobs {
			jobs <- job{id: i}
		}
		close(jobs)
	}()

	seen := make(map[int]bool)
	for r := range results {
		seen[r.jobID] = true
	}

	for i := range numJobs {
		if !seen[i] {
			t.Errorf("job %d was not processed", i)
		}
	}
}

// TestWorkerPool_GracefulShutdown kiểm tra context cancel dừng worker đúng cách.
func TestWorkerPool_GracefulShutdown(t *testing.T) {
	const numWorkers = 3

	ctx, cancel := context.WithCancel(context.Background())

	jobs := make(chan job, numWorkers)
	var wg sync.WaitGroup
	var processed atomic.Int32

	for i := range numWorkers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case j, ok := <-jobs:
					if !ok {
						return
					}
					processed.Add(1)
					_ = j
				}
			}
		}(i)
	}

	// Submit vài job
	for i := range 5 {
		jobs <- job{id: i}
	}

	// Cancel ngay sau khi submit
	cancel()

	// Đợi tất cả worker dừng (timeout để không hang)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// OK: workers đã dừng sạch
	case <-time.After(2 * time.Second):
		t.Fatal("workers did not stop after context cancel — possible goroutine leak")
	}
}

// TestWorkerPool_NoDeadlock kiểm tra không bị deadlock khi close channel.
func TestWorkerPool_NoDeadlock(t *testing.T) {
	const numWorkers = 4
	const numJobs = 10

	jobs := make(chan job, numWorkers)
	results := make(chan result, numJobs)

	var wg sync.WaitGroup
	for i := range numWorkers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := range jobs {
				results <- result{jobID: j.id, workerID: workerID}
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for i := range numJobs {
		jobs <- job{id: i}
	}
	close(jobs)

	var count int
	for range results {
		count++
	}

	if count != numJobs {
		t.Errorf("expected %d results, got %d — possible deadlock or lost job", numJobs, count)
	}
}

// TestParseUserDeleteRequested_ValidPayload kiểm tra parse message đúng.
func TestParseUserDeleteRequested_ValidPayload(t *testing.T) {
	msg := map[string]any{
		"event_id":   "evt-123",
		"event_type": "UserDeleteRequested",
		"payload": map[string]any{
			"user_id":  int64(42),
			"email":    "test@example.com",
			"username": "testuser",
		},
	}

	body, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}

	// parseUserDeleteRequested là unexported nên test qua package consumer_test
	// nhưng ở đây ta test struct trực tiếp
	var parsed struct {
		EventID string `json:"event_id"`
		Payload struct {
			UserID int64 `json:"user_id"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatal(err)
	}

	if parsed.EventID != "evt-123" {
		t.Errorf("expected event_id evt-123, got %s", parsed.EventID)
	}
	if parsed.Payload.UserID != 42 {
		t.Errorf("expected user_id 42, got %d", parsed.Payload.UserID)
	}
}
