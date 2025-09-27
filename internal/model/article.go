package model

import (
	"time"
)

type Article struct {
	Id          string
	AuthorId    string
	Slug        string
	Title       string
	Body        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   time.Time
	TagList     []string
}
