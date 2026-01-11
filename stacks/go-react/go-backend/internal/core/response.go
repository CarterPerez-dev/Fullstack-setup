// AngelaMos | 2026
// response.go

package core

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Page       int `json:"page,omitempty"`
	PageSize   int `json:"page_size,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := Response{
		Success: status >= 200 && status < 300,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func JSONWithMeta(w http.ResponseWriter, status int, data any, meta *Meta) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func JSONError(w http.ResponseWriter, err error) {
	appErr := GetAppError(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.StatusCode)

	response := Response{
		Success: false,
		Error: &Error{
			Code:    appErr.Code,
			Message: appErr.Message,
		},
	}

	if encErr := json.NewEncoder(w).Encode(response); encErr != nil {
		slog.Error("failed to encode error response", "error", encErr)
	}
}

func Created(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, data)
}

func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, data)
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func BadRequest(w http.ResponseWriter, message string) {
	JSONError(w, ValidationError(message))
}

func NotFound(w http.ResponseWriter, resource string) {
	JSONError(w, NotFoundError(resource))
}

func Unauthorized(w http.ResponseWriter, message string) {
	JSONError(w, UnauthorizedError(message))
}

func Forbidden(w http.ResponseWriter, message string) {
	JSONError(w, ForbiddenError(message))
}

func InternalServerError(w http.ResponseWriter, err error) {
	slog.Error("internal server error", "error", err)
	JSONError(w, InternalError(err))
}

func Paginated(w http.ResponseWriter, data any, page, pageSize, total int) {
	totalPages := (total + pageSize - 1) / pageSize
	JSONWithMeta(w, http.StatusOK, data, &Meta{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	})
}
