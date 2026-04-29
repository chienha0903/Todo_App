package valueobject

import "testing"

func TestNewTodoDescription(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "valid description",
			value:     "Go to the supermarket",
			wantValue: "Go to the supermarket",
		},
		{
			name:      "trims spaces",
			value:     "  Go to the supermarket  ",
			wantValue: "Go to the supermarket",
		},
		{
			name:    "empty description",
			value:   "",
			wantErr: true,
		},
		{
			name:    "blank description",
			value:   "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTodoDescription(tt.value)
			if tt.wantErr {
				assertInvalidParameterError(t, err)
				return
			}
			if err != nil {
				t.Fatalf("NewTodoDescription() error = %v", err)
			}
			if got.Value() != tt.wantValue {
				t.Fatalf("NewTodoDescription().Value() = %q, want %q", got.Value(), tt.wantValue)
			}
		})
	}
}
