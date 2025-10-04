package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/minhhoccode111/realworldgo/internal/utils"
)

// Article is the database model
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

type ArticleCreate struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Body        string   `json:"body"`
	TagList     []string `json:"tagList"`
}

func (ac *ArticleCreate) Validate() (err error) {
	ac.Title, err = utils.IsValidTitle(ac.Title)
	if err != nil {
		return err
	}

	ac.Body = strings.TrimSpace(ac.Body)
	if ac.Body == "" {
		return fmt.Errorf("body cannot be empty")
	}

	ac.Description, err = utils.IsValidDescription(ac.Description)
	if err != nil {
		return err
	}

	ac.TagList, err = utils.IsValidTagList(ac.TagList)
	if err != nil {
		return err
	}
	return nil
}

// ArticleDetail is article with body
type ArticleDetail struct {
	Slug           string         `json:"slug"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	TagList        []string       `json:"tagList"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	Favorited      bool           `json:"favorited"`
	FavoritesCount int            `json:"favoritesCount"`
	Author         ProfilePreview `json:"author"`
	Body           string         `json:"body"`
}

// ToArticleDetail attachs 'author' and 'tagList' to Article model to make it an ArticleDetail
func (a *Article) ToArticleDetail(
	author ProfilePreview,
	tagList []string,
) *ArticleDetail {
	return &ArticleDetail{
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

// ArticlePreview is article without body
type ArticlePreview struct {
	Slug           string         `json:"slug"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	TagList        []string       `json:"tagList"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	Favorited      bool           `json:"favorited"`
	FavoritesCount int            `json:"favoritesCount"`
	Author         ProfilePreview `json:"author"`
}
