package valueobject

import "testing"

func TestNewTodoPriority(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantValue TodoPriority
		wantErr   bool
	}{
		{
			name:      "low",
			value:     "LOW",
			wantValue: TodoPriorityLow,
		},
		{
			name:      "medium",
			value:     "MEDIUM",
			wantValue: TodoPriorityMedium,
		},
		{
			name:      "high",
			value:     "HIGH",
			wantValue: TodoPriorityHigh,
		},
		{
			name:      "normalizes lowercase and spaces",
			value:     "  high  ",
			wantValue: TodoPriorityHigh,
		},
		{
			name:    "empty priority",
			value:   "",
			wantErr: true,
		},
		{
			name:    "invalid priority",
			value:   "URGENT",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTodoPriority(tt.value)
			if tt.wantErr {
				assertInvalidParameterError(t, err)
				return
			}
			if err != nil {
				t.Fatalf("NewTodoPriority() error = %v", err)
			}
			if got != tt.wantValue {
				t.Fatalf("NewTodoPriority() = %q, want %q", got, tt.wantValue)
			}
			if got.String() != string(tt.wantValue) {
				t.Fatalf("TodoPriority.String() = %q, want %q", got.String(), tt.wantValue)
			}
		})
	}
}
