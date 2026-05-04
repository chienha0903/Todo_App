package valueobject

import (
	"testing"
)

func TestNewTodoTitle(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "valid title",
			value:     "Buy milk",
			wantValue: "Buy milk",
		},
		{
			name:      "trims spaces",
			value:     "  Buy milk  ",
			wantValue: "Buy milk",
		},
		{
			name:    "empty title",
			value:   "",
			wantErr: true,
		},
		{
			name:    "blank title",
			value:   "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTodoTitle(tt.value)
			if tt.wantErr {
				assertInvalidParameterError(t, err)
				return
			}
			if err != nil {
				t.Fatalf("NewTodoTitle() error = %v", err)
			}
			if got.Value() != tt.wantValue {
				t.Fatalf("NewTodoTitle().Value() = %q, want %q", got.Value(), tt.wantValue)
			}
		})
	}
}
