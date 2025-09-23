package model

import "time"

type Comment struct {
	Id        string    `json:"id"`
	ArticleId string    `json:"article_id"`
	AuthorId  string    `json:"author_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	DeletedAt time.Time `json:"deleted_at"`
	Article   Article   `json:"article"`
	Author    User      `json:"author"`
}
