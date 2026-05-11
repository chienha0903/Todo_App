package apperror

type Code int

const (
	CodeInternal Code = iota
	CodeNotFound
	CodeInvalidArgument
	CodeUnauthorized
	CodePermissionDenied
	CodeTimeout
	CodeUnavailable
)

type Error struct {
	Code    Code
	Message string
}

func (e *Error) Error() string { return e.Message }

func New(code Code, msg string) *Error  { return &Error{Code: code, Message: msg} }
func Internal() *Error                  { return New(CodeInternal, "internal server error") }
func NotFound(msg string) *Error        { return New(CodeNotFound, msg) }
func InvalidArgument(msg string) *Error { return New(CodeInvalidArgument, msg) }
func Unauthorized() *Error              { return New(CodeUnauthorized, "unauthorized") }
func PermissionDenied() *Error          { return New(CodePermissionDenied, "permission denied") }
func Timeout() *Error                   { return New(CodeTimeout, "request timed out") }
func Unavailable() *Error               { return New(CodeUnavailable, "service unavailable") }
