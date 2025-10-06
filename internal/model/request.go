package model

type UserRegisterRequest struct {
	User UserRegister `json:"user"`
}

type UserLoginRequest struct {
	User UserLogin `json:"user"`
}

type UserUpdateRequest struct {
	User UserUpdate `json:"user"`
}

type ArticleCreateRequest struct {
	Article ArticleCreate `json:"article"`
}

type ArticleUpdateRequest struct {
	Article ArticleUpdate `json:"article"`
}

type CommentCreateRequest struct {
	Comment CommentCreate `json:"comment"`
}
