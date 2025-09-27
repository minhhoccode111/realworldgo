package model

import "time"

type UserResponse struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Token    string `json:"token"`
	Bio      string `json:"bio"`
	Image    string `json:"image"`
}

type ProfilePreviewResponse struct {
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Following bool   `json:"following"`
}

// ArticlePreviewResponse view an article without its 'Body'
type ArticlePreviewResponse struct {
	Slug           string                 `json:"slug"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	TagList        []string               `json:"tagList"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
	Favorited      bool                   `json:"favorited"`
	FavoritesCount int                    `json:"favoritesCount"`
	Author         ProfilePreviewResponse `json:"author"`
}

// ArticleDetailResponse view an article
type ArticleDetailResponse struct {
	ArticlePreviewResponse
	Body string `json:"body"`
}

// ArticlesResponse preview a list of articles
type ArticlesResponse struct {
	Articles      []ArticlePreviewResponse `json:"articles"`
	ArticlesCount int                      `json:"articlesCount"`
}
