package handler

import (
	"errors"
	"net/http"
	"strings"

	authdto "timebox-backend/internal/dto/auth"
	"timebox-backend/internal/entity"
	"timebox-backend/internal/response"
	"timebox-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func newAuthHandler(services *service.Service) *AuthHandler {
	return &AuthHandler{authService: services.Auth}
}

func (h *AuthHandler) RegisterRoutes(routeGroup *gin.RouterGroup) {
	r := routeGroup.Group("/auth")

	r.POST("/register", h.Register)
	r.POST("/login", h.Login)
	r.POST("/refresh", h.Refresh)
	r.POST("/logout", h.Logout)
}

func (h *AuthHandler) Register(ctx *gin.Context) {
	var req authdto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}

	user, tokens, err := h.authService.Register(ctx, entity.User{
		FullName: req.FullName,
		Email:    req.Email,
		Timezone: req.Timezone,
	}, req.Password)
	if err != nil {
		writeAuthError(ctx, err)
		return
	}

	response.WithData(ctx, authResponse(user, tokens, true), "Account registered successfully", http.StatusCreated)
}

func (h *AuthHandler) Login(ctx *gin.Context) {
	var req authdto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}

	user, tokens, err := h.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		writeAuthError(ctx, err)
		return
	}

	response.WithData(ctx, authResponse(user, tokens, false), "Login successful", http.StatusOK)
}

func (h *AuthHandler) Refresh(ctx *gin.Context) {
	var req authdto.RefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}

	tokens, err := h.authService.Refresh(ctx, req.RefreshToken)
	if err != nil {
		writeAuthError(ctx, err)
		return
	}

	response.WithData(ctx, tokenResponse(tokens), "Token refreshed", http.StatusOK)
}

func (h *AuthHandler) Logout(ctx *gin.Context) {
	if _, err := h.authService.ValidateAccessToken(bearerToken(ctx)); err != nil {
		writeAuthError(ctx, err)
		return
	}

	var req authdto.LogoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "bad request", "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.authService.Logout(ctx, req.RefreshToken); err != nil {
		writeAuthError(ctx, err)
		return
	}

	response.WithoutData(ctx, "Logout successful", http.StatusOK)
}

func writeAuthError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrEmailAlreadyExists):
		response.Error(ctx, "conflict", "email already exists", http.StatusConflict)
	case errors.Is(err, service.ErrInvalidCredentials):
		response.Error(ctx, "unauthorized", "invalid email or password", http.StatusUnauthorized)
	case errors.Is(err, service.ErrInvalidToken):
		response.Error(ctx, "unauthorized", "invalid token", http.StatusUnauthorized)
	case errors.Is(err, service.ErrInvalidPassword), errors.Is(err, service.ErrInvalidTimezone):
		response.Error(ctx, "validation error", "invalid request", http.StatusUnprocessableEntity)
	case errors.Is(err, service.ErrRateLimited):
		response.Error(ctx, "rate limited", "too many requests", http.StatusTooManyRequests)
	default:
		response.Error(ctx, "internal server error", "auth request failed", http.StatusInternalServerError)
	}
}

func authResponse(user entity.User, tokens service.TokenSet, includeCreatedAt bool) authdto.AuthResponse {
	userResponse := toAuthUserResponse(user)
	if includeCreatedAt {
		userResponse.CreatedAt = &user.CreatedAt
	}
	return authdto.AuthResponse{
		User:   userResponse,
		Tokens: tokenResponse(tokens),
	}
}

func toAuthUserResponse(user entity.User) authdto.UserResponse {
	return authdto.UserResponse{
		ID:              user.ID,
		FullName:        user.FullName,
		Email:           user.Email,
		Timezone:        user.Timezone,
		AvatarURL:       user.AvatarURL,
		EmailVerifiedAt: user.EmailVerifiedAt,
	}
}

func tokenResponse(tokens service.TokenSet) authdto.TokenResponse {
	return authdto.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
	}
}

func bearerToken(ctx *gin.Context) string {
	authHeader := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
}
