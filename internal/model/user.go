package model

import "time"

type User struct {
	Id          string    `json:"id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	Image       string    `json:"image"`
	Bio         string    `json:"bio"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsFollowing bool      `json:"is_following"`
	Followers   uint      `json:"followers"`
	Following   uint      `json:"following"`
	Favorites   uint      `json:"favorites"`
	Password    string    `json:"password"`
}
