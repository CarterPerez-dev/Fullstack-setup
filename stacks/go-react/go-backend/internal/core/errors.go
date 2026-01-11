// AngelaMos | 2026
// errors.go

package core

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrDuplicateKey = errors.New("duplicate key violation")
	ErrForeignKey   = errors.New("foreign key violation")
	ErrInvalidInput = errors.New("invalid input")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrInternal     = errors.New("internal server error")
	ErrConflict     = errors.New("resource conflict")
	ErrRateLimited  = errors.New("rate limit exceeded")
	ErrTokenExpired = errors.New("token expired")
	ErrTokenInvalid = errors.New("token invalid")
	ErrTokenRevoked = errors.New("token revoked")
)

type AppError struct {
	Err        error  `json:"-"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "unknown error"
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(
	err error,
	message string,
	statusCode int,
	code string,
) *AppError {
	return &AppError{
		Err:        err,
		Message:    message,
		StatusCode: statusCode,
		Code:       code,
	}
}

func NotFoundError(resource string) *AppError {
	return &AppError{
		Err:        ErrNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		StatusCode: http.StatusNotFound,
		Code:       "NOT_FOUND",
	}
}

func DuplicateError(field string) *AppError {
	return &AppError{
		Err:        ErrDuplicateKey,
		Message:    fmt.Sprintf("%s already exists", field),
		StatusCode: http.StatusConflict,
		Code:       "DUPLICATE",
	}
}

func ValidationError(message string) *AppError {
	return &AppError{
		Err:        ErrInvalidInput,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Code:       "VALIDATION_ERROR",
	}
}

func UnauthorizedError(message string) *AppError {
	if message == "" {
		message = "authentication required"
	}
	return &AppError{
		Err:        ErrUnauthorized,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
		Code:       "UNAUTHORIZED",
	}
}

func ForbiddenError(message string) *AppError {
	if message == "" {
		message = "access denied"
	}
	return &AppError{
		Err:        ErrForbidden,
		Message:    message,
		StatusCode: http.StatusForbidden,
		Code:       "FORBIDDEN",
	}
}

func InternalError(err error) *AppError {
	return &AppError{
		Err:        err,
		Message:    "internal server error",
		StatusCode: http.StatusInternalServerError,
		Code:       "INTERNAL_ERROR",
	}
}

func RateLimitError() *AppError {
	return &AppError{
		Err:        ErrRateLimited,
		Message:    "too many requests",
		StatusCode: http.StatusTooManyRequests,
		Code:       "RATE_LIMITED",
	}
}

func TokenExpiredError() *AppError {
	return &AppError{
		Err:        ErrTokenExpired,
		Message:    "token has expired",
		StatusCode: http.StatusUnauthorized,
		Code:       "TOKEN_EXPIRED",
	}
}

func TokenInvalidError() *AppError {
	return &AppError{
		Err:        ErrTokenInvalid,
		Message:    "invalid token",
		StatusCode: http.StatusUnauthorized,
		Code:       "TOKEN_INVALID",
	}
}

func TokenRevokedError() *AppError {
	return &AppError{
		Err:        ErrTokenRevoked,
		Message:    "token has been revoked",
		StatusCode: http.StatusUnauthorized,
		Code:       "TOKEN_REVOKED",
	}
}

func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return InternalError(err)
}
