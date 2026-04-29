package valueobject

import (
	stderrors "errors"
	"testing"

	apperrors "github.com/chienha0903/Todo_App/pkg/errors"
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

func assertInvalidParameterError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var appErr *apperrors.Error
	if !stderrors.As(err, &appErr) {
		t.Fatalf("expected *errors.Error, got %T", err)
	}
	if appErr.Reason != apperrors.REASON_INVALID_PARAMETER {
		t.Fatalf("error reason = %q, want %q", appErr.Reason, apperrors.REASON_INVALID_PARAMETER)
	}
}
