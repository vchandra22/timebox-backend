package authdto

import "time"

type UserResponse struct {
	ID              string     `json:"id"`
	FullName        string     `json:"full_name"`
	Email           string     `json:"email"`
	Timezone        string     `json:"timezone"`
	AvatarURL       *string    `json:"avatar_url"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type AuthResponse struct {
	User   UserResponse  `json:"user"`
	Tokens TokenResponse `json:"tokens"`
}
