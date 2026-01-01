package apperr

import (
	"fmt"
	"runtime"
	"strings"
)

type AppError struct {
	Code      string
	Message   string
	Operation string
	Err       error
}

func (e *AppError) Error() string {
	prefix := ""
	if e.Code != "" {
		prefix = fmt.Sprintf("[%s] ", e.Code)
	}

	if e.Err != nil {
		return fmt.Sprintf("%s%s: %v", prefix, e.Message, e.Err)
	}
	return fmt.Sprintf("%s%s", prefix, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func newSentinelError(code, message string) *AppError {
	return &AppError{
		Code:      code,
		Operation: sentinel_err_operation,
		Message:   message,
	}
}

func New(message string) *AppError {
	return &AppError{
		Operation: callerFuncName(2),
		Message:   message,
	}
}

func Wrap(message string, err error) *AppError {
	return &AppError{
		Operation: callerFuncName(2),
		Message:   message,
		Err:       err,
	}
}

func callerFuncName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	full := fn.Name() // 例: github.com/user/project/internal/service.UserService.FetchUser
	parts := strings.Split(full, "/")
	return parts[len(parts)-1] // 最後の部分だけ取り出す（例: UserService.FetchUser）
}
