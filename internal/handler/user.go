package handler

import (
	"errors"
	"net/http"

	userdto "boilerplate-golang/internal/dto/user"
	"boilerplate-golang/internal/entity"
	"boilerplate-golang/internal/response"
	"boilerplate-golang/internal/service"
	"boilerplate-golang/internal/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func newUserHandler(services *service.Service) *UserHandler {
	return &UserHandler{
		userService: services.User,
	}
}

func (h *UserHandler) RegisterRoutes(routeGroup *gin.RouterGroup) {
	r := routeGroup.Group("/users")

	r.POST("/", h.Create)
	r.GET("/", h.FindAll)
	r.GET("/:id", h.FindByID)
	r.PUT("/:id", h.Update)
	r.DELETE("/:id", h.Delete)
}

func (h *UserHandler) Create(ctx *gin.Context) {
	var req userdto.CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Create(ctx, entity.User{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		writeUserError(ctx, err, "failed to create user")
		return
	}

	response.WithData(ctx, toUserResponse(user), "user created", http.StatusCreated)
}

func (h *UserHandler) FindAll(ctx *gin.Context) {
	filter, err := utils.NewPaginationFilter(ctx.Query("page"), ctx.Query("limit"))
	if err != nil {
		response.Error(ctx, err.Error(), "invalid pagination", http.StatusBadRequest)
		return
	}

	users, total, err := h.userService.FindAll(ctx, filter.Page, filter.Limit)
	if err != nil {
		response.Error(ctx, "internal server error", "failed to get users", http.StatusInternalServerError)
		return
	}

	response.WithPagination(
		ctx,
		toUserResponses(users),
		"users fetched",
		http.StatusOK,
		response.NewPagination(filter.Page, filter.Limit, total),
	)
}

func (h *UserHandler) FindByID(ctx *gin.Context) {
	id, ok := userIDParam(ctx)
	if !ok {
		return
	}

	user, err := h.userService.FindByID(ctx, id)
	if err != nil {
		writeUserError(ctx, err, "failed to get user")
		return
	}

	response.WithData(ctx, toUserResponse(user), "user fetched", http.StatusOK)
}

func (h *UserHandler) Update(ctx *gin.Context) {
	id, ok := userIDParam(ctx)
	if !ok {
		return
	}

	var req userdto.UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Update(ctx, entity.User{
		ID:    id,
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		writeUserError(ctx, err, "failed to update user")
		return
	}

	response.WithData(ctx, toUserResponse(user), "user updated", http.StatusOK)
}

func (h *UserHandler) Delete(ctx *gin.Context) {
	id, ok := userIDParam(ctx)
	if !ok {
		return
	}

	if err := h.userService.Delete(ctx, id); err != nil {
		writeUserError(ctx, err, "failed to delete user")
		return
	}

	response.WithoutData(ctx, "user deleted", http.StatusOK)
}

func toUserResponses(users []entity.User) []userdto.Response {
	res := make([]userdto.Response, 0, len(users))
	for _, user := range users {
		res = append(res, toUserResponse(user))
	}

	return res
}

func userIDParam(ctx *gin.Context) (string, bool) {
	id := ctx.Param("id")
	if !validUUID(id) {
		response.Error(ctx, "bad request", "invalid user id", http.StatusBadRequest)
		return "", false
	}
	return id, true
}

func writeUserError(ctx *gin.Context, err error, fallbackMessage string) {
	errorText, message, code := userErrorResponse(err, fallbackMessage)
	response.Error(ctx, errorText, message, code)
}

func userErrorResponse(err error, fallbackMessage string) (string, string, int) {
	if errors.Is(err, service.ErrUserNotFound) {
		return "not found", "user not found", http.StatusNotFound
	}
	if errors.Is(err, service.ErrEmailAlreadyExists) {
		return "conflict", "email already exists", http.StatusConflict
	}
	return "internal server error", fallbackMessage, http.StatusInternalServerError
}

func validUUID(id string) bool {
	if len(id) != 36 {
		return false
	}

	for i, c := range id {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
			continue
		}
		if !isHex(c) {
			return false
		}
	}

	return true
}

func isHex(c rune) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}

func toUserResponse(user entity.User) userdto.Response {
	return userdto.Response{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
