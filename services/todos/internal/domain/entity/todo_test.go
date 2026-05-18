package entity

import (
	"testing"
	"time"

	vo "github.com/chienha0903/Todo_App/services/todos/internal/domain/valueobject"
)

func TestIsOverdue(t *testing.T) {
	now := time.Date(2026, 5, 7, 12, 0, 0, 0, time.UTC)

	pastDueDate, _ := vo.NewTodoDueDate(now.Add(-24 * time.Hour))
	futureDueDate, _ := vo.NewTodoDueDate(now.Add(24 * time.Hour))

	tests := []struct {
		name string
		todo *Todo
		want bool
	}{
		{
			name: "nil todo",
			todo: nil,
			want: false,
		},
		{
			name: "no due date",
			todo: &Todo{Status: vo.TodoStatusPending},
			want: false,
		},
		{
			name: "overdue - pending",
			todo: &Todo{Status: vo.TodoStatusPending, DueDate: &pastDueDate},
			want: true,
		},
		{
			name: "overdue - in progress",
			todo: &Todo{Status: vo.TodoStatusInProgress, DueDate: &pastDueDate},
			want: true,
		},
		{
			name: "overdue - but completed",
			todo: &Todo{Status: vo.TodoStatusCompleted, DueDate: &pastDueDate},
			want: false,
		},
		{
			name: "due date in future",
			todo: &Todo{Status: vo.TodoStatusPending, DueDate: &futureDueDate},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsOverdue(tt.todo, now)
			if got != tt.want {
				t.Fatalf("IsOverdue() = %v, want %v", got, tt.want)
			}
		})
	}
}
