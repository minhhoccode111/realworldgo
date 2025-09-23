package model

import "time"

type User struct {
	Id        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Image     string    `json:"image"`
	Bio       string    `json:"bio"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Password  string
}

type UserDTO struct {
	Id        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Image     string    `json:"image"`
	Bio       string    `json:"bio"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
