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

	var appErr *apperrors.Error
	if !stderrors.As(err, &appErr) {
		t.Fatalf("expected *errors.Error, got %T", err)
	}
	if appErr.Reason != apperrors.ReasonInvalidParameter {
		t.Fatalf("error reason = %q, want %q", appErr.Reason, apperrors.ReasonInvalidParameter)
	}
}
