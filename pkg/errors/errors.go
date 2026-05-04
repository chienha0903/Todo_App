package errors

type Reason string

const (
	ReasonNotFound            Reason = "NOT_FOUND"
	ReasonInvalidParameter    Reason = "INVALID_PARAMETER"
	ReasonUnauthorized        Reason = "UNAUTHORIZED"
	ReasonPermissionDenied    Reason = "PERMISSION_DENIED"
	ReasonInternalServerError Reason = "INTERNAL_SERVER_ERROR"
)

type Error struct {
	Reason  Reason
	Message string
}

func NewAppError(reason Reason, message string) *Error {
	return &Error{
		Reason:  reason,
		Message: message,
	}
}

func (e *Error) Error() string {
	return string(e.Reason) + ": " + e.Message
}
