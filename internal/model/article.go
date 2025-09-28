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
}

func (a *Article) ToArticleDetailResponse(
	author ProfilePreviewResponse,
	tagList []string,
) *ArticleDetailResponse {
	return &ArticleDetailResponse{
		Slug:           a.Slug,
		Title:          a.Title,
		Description:    a.Description,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
		Favorited:      false,
		FavoritesCount: 0,
		Author:         author,
		TagList:        tagList,
		Body:           a.Body,
	}
}
