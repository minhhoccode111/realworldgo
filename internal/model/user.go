package model

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type User struct {
	Id       string `json:"id"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
	Role     Role   `json:"role"`
	Password string
}

type UserDTO struct {
	Id       string `json:"id"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
	Role     Role   `json:"role"`
}
