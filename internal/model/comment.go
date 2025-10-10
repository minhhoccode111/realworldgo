package model

import (
	"fmt"
	"strings"
	"time"
)

type Comment struct {
	Id        string
	ArticleId string
	AuthorId  string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

type CommentCreate struct {
	Body string `json:"body"`
}

func (c *CommentCreate) Validate() error {
	body := strings.TrimSpace(c.Body)
	if body == "" {
		return fmt.Errorf("body cannot be empty")
	}

	c.Body = body
	return nil
}

type CommentDetail struct {
	Id        string         `json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	Body      string         `json:"body"`
	Author    ProfilePreview `json:"author"`
}
