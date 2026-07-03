package database

import "time"

type Row struct {
	ID              string     `db:"id"`
	FullName        string     `db:"full_name"`
	Email           string     `db:"email"`
	PasswordHash    string     `db:"password_hash"`
	Timezone        string     `db:"timezone"`
	AvatarURL       *string    `db:"avatar_url"`
	EmailVerifiedAt *time.Time `db:"email_verified_at"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}
