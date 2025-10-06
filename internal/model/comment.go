package model

import (
	"fmt"
	"strings"
	"time"
)

type Comment struct {
	Id        string    `json:"id"`
	ArticleId string    `json:"article_id"`
	AuthorId  string    `json:"author_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	DeletedAt time.Time `json:"deleted_at"`
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
	Id        string
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	Body      string         `json:"body"`
	Author    ProfilePreview `json:"author"`
}
