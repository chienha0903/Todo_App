package valueobject

import "testing"

func TestNewTodoStatus(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantValue TodoStatus
		wantErr   bool
	}{
		{
			name:      "pending",
			value:     "PENDING",
			wantValue: TODO_STATUS_PENDING,
		},
		{
			name:      "in progress",
			value:     "IN_PROGRESS",
			wantValue: TODO_STATUS_IN_PROGRESS,
		},
		{
			name:      "completed",
			value:     "COMPLETED",
			wantValue: TODO_STATUS_COMPLETED,
		},
		{
			name:      "normalizes lowercase and spaces",
			value:     "  completed  ",
			wantValue: TODO_STATUS_COMPLETED,
		},
		{
			name:    "empty status",
			value:   "",
			wantErr: true,
		},
		{
			name:    "invalid status",
			value:   "DONE",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTodoStatus(tt.value)
			if tt.wantErr {
				assertInvalidParameterError(t, err)
				return
			}
			if err != nil {
				t.Fatalf("NewTodoStatus() error = %v", err)
			}
			if got != tt.wantValue {
				t.Fatalf("NewTodoStatus() = %q, want %q", got, tt.wantValue)
			}
			if got.String() != string(tt.wantValue) {
				t.Fatalf("TodoStatus.String() = %q, want %q", got.String(), tt.wantValue)
			}
		})
	}
}
