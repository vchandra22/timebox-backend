package entity

import "time"

type User struct {
	ID              string
	FullName        string
	Email           string
	PasswordHash    string
	Timezone        string
	AvatarURL       *string
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
