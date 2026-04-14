package errors

type Reason string

const (
	REASON_NOT_FOUND Reason = "NOT_FOUND"
	REASON_INVALID_PARAMETER Reason = "INVALID_PARAMETER"
	REASON_UNAUTHORIZED Reason = "UNAUTHORIZED"
	REASON_PERMISSION_DENIED Reason = "PERMISSION_DENIED"
	REASON_INTERNAL_SERVER_ERROR Reason = "INTERNAL_SERVER_ERROR"
)

type Error struct {
	Reason  Reason
	Message string
}

func New(reason Reason, message string) *Error {
	return &Error{
		Reason:  reason,
		Message: message,
	}
}

func (e *Error) Error() string {
	return string(e.Reason) + ": " + e.Message
}