package model

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// UserAuthResponse
type UserAuthResponse struct {
	User UserAuth `json:"user"`
}

type ArticleDetailResponse struct {
	Article ArticleDetail `json:"article"`
}

// ArticlesResponse preview a list of articles
type ArticlesResponse struct {
	Articles      []ArticlePreview `json:"articles"`
	ArticlesCount int              `json:"articlesCount"`
	Limit         int              `json:"limit"`
	Offset        int              `json:"offset"`
}

type ProfilePreviewResponse struct {
	Profile ProfilePreview `json:"profile"`
}
