package response

import "github.com/gin-gonic/gin"

type SuccessResponse[T any] struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    T      `json:"data"`
	Meta    any    `json:"meta"`
}

type PaginatedResponse[T any] struct {
	Status  bool       `json:"status"`
	Message string     `json:"message"`
	Data    T          `json:"data"`
	Meta    Pagination `json:"meta"`
}

type MessageResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Meta    any    `json:"meta"`
}

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

type ErrorResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

func NewPagination(page, limit, total int) Pagination {
	totalPages := 0
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}

	return Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

func WithData[T any](ctx *gin.Context, data T, message string, code int) {
	ctx.JSON(code, SuccessResponse[T]{
		Status:  true,
		Message: message,
		Data:    data,
		Meta:    nil,
	})
}

func WithPagination[T any](ctx *gin.Context, data T, message string, code int, pagination Pagination) {
	ctx.JSON(code, PaginatedResponse[T]{
		Status:  true,
		Message: message,
		Data:    data,
		Meta:    pagination,
	})
}

func WithoutData(ctx *gin.Context, message string, code int) {
	ctx.JSON(code, MessageResponse{
		Status:  true,
		Message: message,
		Data:    nil,
		Meta:    nil,
	})
}

func Error(ctx *gin.Context, err string, message string, code int) {
	ctx.JSON(code, ErrorResponse{
		Status:  false,
		Message: message,
		Error:   err,
	})
}
