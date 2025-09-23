package entity

import "time"

type User struct {
	ID        string    `json:"id"         db:"id"`
	Email     string    `json:"email"      db:"email"`
	Username  string    `json:"username"   db:"username"`
	Bio       *string   `json:"bio"        db:"bio"`
	Image     *string   `json:"image"      db:"image"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
