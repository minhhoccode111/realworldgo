package model

import "strings"

type ResponseError struct {
	Errors  []string `json:"error"`
	Details string   `json:"details,omitempty"`
	Code    int      `json:"code,omitempty"`
}

func (re ResponseError) Error() string {
	switch {
	case len(re.Errors) == 0:
		return "unknown error"
	case len(re.Errors) == 1:
		return re.Errors[0]
	default:
		return strings.Join(re.Errors, "; ")
	}
}

func NewError(err ...error) ResponseError {
	errors := []string{}
	for _, v := range err {
		errors = append(errors, v.Error())
	}
	return ResponseError{Errors: errors}
}

type UserAuthResponse struct {
	User UserAuth `json:"user"`
}

type ArticleDetailResponse struct {
	Article ArticleDetail `json:"article"`
}

type ArticlesResponse struct {
	Articles      []ArticlePreview `json:"articles"`
	ArticlesCount int              `json:"articlesCount"`
	Limit         int              `json:"limit"`
	Offset        int              `json:"offset"`
}

type ProfilePreviewResponse struct {
	Profile ProfilePreview `json:"profile"`
}

type CommentDetailResponse struct {
	Comment CommentDetail `json:"comment"`
}

type CommentsResponse struct {
	Comments      []CommentDetail `json:"comments"`
	CommentsCount int             `json:"comments_count"`
	Limit         int             `json:"limit"`
	Offset        int             `json:"offset"`
}

type TagsResponse struct {
	Tags      []TagName `json:"tags"`
	TagsCount int       `json:"tags_count"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
}
