package model

import (
	"time"
)

type Article struct {
	Id          string    `json:"id"`
	AuthorId    string    `json:"author_id"`
	Slug        string    `json:"slug"`
	Body        string    `json:"body"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   time.Time `json:"deleted_at"`
	Author      User      `json:"author"`
	Tags        []Tag     `json:"tags"`
	Favorites   uint      `json:"favorites"`
	IsFavorited bool      `json:"is_favorited"`
}
