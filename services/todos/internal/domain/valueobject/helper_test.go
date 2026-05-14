package valueobject

import (
	stderrors "errors"
	"testing"

	apperrors "github.com/chienha0903/Todo_App/pkg/errors"
)

func assertInvalidParameterError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var appErr *apperrors.AppError
	if !stderrors.As(err, &appErr) {
		t.Fatalf("expected *errors.AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrInvalidParameter {
		t.Fatalf("error code = %q, want %q", appErr.Code, apperrors.ErrInvalidParameter)
	}
}
