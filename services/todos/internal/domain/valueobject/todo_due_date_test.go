package valueobject

import (
	"testing"
	"time"
)

func TestNewTodoDueDate(t *testing.T) {
	validDueDate := time.Date(2026, 5, 1, 9, 30, 0, 0, time.UTC)

	tests := []struct {
		name    string
		value   time.Time
		wantErr bool
	}{
		{
			name:  "valid due date",
			value: validDueDate,
		},
		{
			name:    "zero due date",
			value:   time.Time{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTodoDueDate(tt.value)
			if tt.wantErr {
				assertInvalidParameterError(t, err)
				return
			}
			if err != nil {
				t.Fatalf("NewTodoDueDate() error = %v", err)
			}
			if !got.Value().Equal(tt.value) {
				t.Fatalf("NewTodoDueDate().Value() = %v, want %v", got.Value(), tt.value)
			}
		})
	}
}
