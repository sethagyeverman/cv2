package response

import (
	"context"
	"cv2/internal/errx"
	"errors"
	"net/http"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// Success creates a successful response with code 200
func Success(ctx context.Context, data any) any {
	return &Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	}
}

// Error creates an error response, extracting the code from errx.Error if possible
func Error(ctx context.Context, err error) (int, any) {
	if err == nil {
		return http.StatusOK, Success(ctx, nil)
	}

	var e *errx.Error
	if errors.As(err, &e) {
		code := e.Code()
		if code == 0 {
			code = http.StatusInternalServerError
		}
		return code, &Response{
			Code:    code,
			Message: err.Error(),
			Data:    nil,
		}
	}

	return http.StatusInternalServerError, &Response{
		Code:    http.StatusInternalServerError,
		Message: err.Error(),
		Data:    nil,
	}
}
