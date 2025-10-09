package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/minhhoccode111/realworldgo/internal/middleware"

	// no need for prefix 'model.' and 'utils.'
	. "github.com/minhhoccode111/realworldgo/internal/model"
	. "github.com/minhhoccode111/realworldgo/internal/utils"

	"github.com/gorilla/mux"

	"github.com/coder/websocket"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := mux.NewRouter()

	corsMiddleware := middleware.CorsMiddleware(s.config.CORS.AllowedOrigins)

	// Apply global middleware
	r.Use(middleware.TimeoutMiddleware)
	r.Use(corsMiddleware)

	// Root level routes (no versioning)
	r.HandleFunc("/", s.HelloWorldHandler)
	r.HandleFunc("/healthz", s.healthHandler)

	// API v1 routes
	v1 := r.PathPrefix("/api/v1").Subrouter()
	s.registerV1Routes(v1)

	return r
}

func (s *Server) registerV1Routes(r *mux.Router) {
	// WARN: must define '/articles/feed' before '/articles/{slug}'

	auth := middleware.AuthMiddleware(false, s.config.JWT.Secret, s.db)
	optionalAuth := middleware.AuthMiddleware(true, s.config.JWT.Secret, s.db)

	r.HandleFunc("/websocket", s.websocketHandler)

	r.HandleFunc("/users", s.RegisterHandler).Methods("POST")
	r.HandleFunc("/users/login", s.LoginHandler).Methods("POST")

	r.HandleFunc("/user", auth(s.GetUserHandler)).Methods("GET")
	r.HandleFunc("/user", auth(s.PutUserHandler)).Methods("PUT")

	r.HandleFunc("/articles", auth(s.PostArticleHandler)).Methods("POST")
	r.HandleFunc("/articles", optionalAuth(s.GetAllArticlesHandler)).Methods("GET")
	r.HandleFunc("/articles/feed", auth(s.GetFeedHandler)).Methods("GET")
	r.HandleFunc("/articles/{slug}", optionalAuth(s.GetArticleHandler)).Methods("GET")
	r.HandleFunc("/articles/{slug}", auth(s.PutArticleHandler)).Methods("PUT")
	r.HandleFunc("/articles/{slug}", auth(s.DeleteArticleHandler)).Methods("DELETE")

	r.HandleFunc("/articles/{slug}/favorite", auth(s.PostFavoriteHandler)).Methods("POST")
	r.HandleFunc("/articles/{slug}/favorite", auth(s.DeleteFavoriteHandler)).Methods("DELETE")

	r.HandleFunc("/articles/{slug}/comments", optionalAuth(s.GetCommentsHandler)).Methods("GET")
	r.HandleFunc("/articles/{slug}/comments", auth(s.PostCommentsHandler)).Methods("POST")
	r.HandleFunc("/articles/{slug}/comments/{commentId}", auth(s.DeleteCommentsHandler)).
		Methods("DELETE")

	r.HandleFunc("/profiles/{username}", optionalAuth(s.GetProfilehandler)).Methods("GET")
	r.HandleFunc("/profiles/{username}/follow", auth(s.PostFollowHandler)).Methods("POST")
	r.HandleFunc("/profiles/{username}/follow", auth(s.DeleteFollowHandler)).Methods("DELETE")

	r.HandleFunc("/tags", s.GetTagsHandler).Methods("GET")
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	WriteText(w, http.StatusOK, "Hello, World!")
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, s.db.Health())
}

func (s *Server) websocketHandler(w http.ResponseWriter, r *http.Request) {
	socket, err := websocket.Accept(w, r, nil)

	if err != nil {
		log.Printf("could not open websocket: %v", err)
		WriteText(w, http.StatusInternalServerError, "could not open websocket")
		return
	}

	defer socket.Close(websocket.StatusGoingAway, "server closing websocket")

	ctx := r.Context()
	socketCtx := socket.CloseRead(ctx)

	for {
		payload := fmt.Sprintf("server timestamp: %d", time.Now().UnixNano())
		err := socket.Write(socketCtx, websocket.MessageText, []byte(payload))
		if err != nil {
			break
		}
		time.Sleep(time.Second * 2)
	}
}

func (s *Server) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var body UserRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("Error decode request body: %v", err)
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err := body.User.Validate()
	if err != nil {
		log.Printf("UserRegisterRequest model failed validations: %v", err)
		WriteJSON(w, http.StatusUnprocessableEntity, ErrorResponse{Error: err.Error()})
		return
	}

	// WARN: don't check for uniqueness manually, because race conditions can occur
	// we must check for uniqueness if error returned after we insert to db
	// userExisted, err := s.db.SelectUserByEmail(r.Context(), body.User.Email)

	newUser := User{
		Email:    body.User.Email,
		Username: body.User.Username,
		Password: body.User.Password,
	}

	err = s.db.CreateUser(r.Context(), &newUser)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				WriteJSON(w, http.StatusConflict, ErrorResponse{Error: err.Error()})
				return
			}
		}

		log.Printf("Error inserting user: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	token, err := GenerateJWT(s.config.JWT, newUser.Id)
	if err != nil {
		log.Printf("Error generating jwt: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	userAuth := newUser.ToUserAuth(token)
	WriteJSON(w, http.StatusCreated, UserAuthResponse{User: *userAuth})
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var body UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("Error decode request body: %v", err)
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userExisted, err := s.db.SelectUser(r.Context(), "", body.User.Email, "")
	if err != nil {
		if err == sql.ErrNoRows {
			WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "email not found"})
			return
		}
		log.Printf("Error selecting user: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	if !ValidatePassword(userExisted.Password, body.User.Password) {
		WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "password incorrect"})
		return
	}

	token, err := GenerateJWT(s.config.JWT, userExisted.Id)
	if err != nil {
		log.Printf("Error generating jwt: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	userAuth := userExisted.ToUserAuth(token)
	WriteJSON(w, http.StatusOK, UserAuthResponse{User: *userAuth})
}

func (s *Server) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(CtxUserKey).(User)
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "cannot authorize user in jwt"})
		return
	}
	token, err := GenerateJWT(s.config.JWT, user.Id)
	if err != nil {
		log.Printf("Error generating jwt: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	userAuth := user.ToUserAuth(token)
	WriteJSON(w, http.StatusOK, UserAuthResponse{User: *userAuth})
}

func (s *Server) PutUserHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var body UserUpdateRequest
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("Error decode request body: %v", err)
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	currentUser, ok := r.Context().Value(CtxUserKey).(User)
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "cannot authorize user in jwt"})
		return
	}

	err = currentUser.ValidateUserUpdate(&body.User)
	if err != nil {
		log.Printf("Error validating user update request: %v", err)
		WriteJSON(w, http.StatusUnprocessableEntity, ErrorResponse{Error: err.Error()})
		return
	}

	err = s.db.UpdateUser(r.Context(), &currentUser)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	token, err := GenerateJWT(s.config.JWT, currentUser.Id)
	userAuth := currentUser.ToUserAuth(token)
	WriteJSON(w, http.StatusOK, UserAuthResponse{User: *userAuth})
}

func (s *Server) PostArticleHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var body ArticleCreateRequest
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("Error decode request body: %v", err)
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	currentUser, ok := r.Context().Value(CtxUserKey).(User)
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "cannot authorize user in jwt"})
		return
	}

	err = body.Article.Validate()
	if err != nil {
		log.Printf("Error validating article create request: %v", err)
		WriteJSON(w, http.StatusUnprocessableEntity, ErrorResponse{Error: err.Error()})
		return
	}

	newArticle := Article{
		AuthorId:    currentUser.Id,
		Title:       body.Article.Title,
		Body:        body.Article.Body,
		Description: body.Article.Description,
	}

	newSlug, err := s.db.CreateArticle(r.Context(), &newArticle, body.Article.TagList)
	if err != nil {
		log.Printf("Error creating article: %v", err)
		WriteJSON(w, http.StatusUnprocessableEntity, ErrorResponse{Error: err.Error()})
		return
	}

	articleDetail, err := s.db.SelectArticleDetail(r.Context(), currentUser.Id, newSlug)
	if err != nil {
		log.Printf("Error selecting new article details: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, ArticleDetailResponse{Article: *articleDetail})
}

func (s *Server) GetAllArticlesHandler(w http.ResponseWriter, r *http.Request) {
	isAuth := r.Context().Value(CtxIsAuthKey).(bool)
	currentUser, ok := r.Context().Value(CtxUserKey).(User)
	if !ok && isAuth {
		WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "cannot authorize user in jwt"})
		return
	}

	var currentUserId string
	if isAuth {
		currentUserId = currentUser.Id
	}

	tag, author, favorited, limit, offset := SearchQueries(w, r)

	articles, articlesCount, err := s.db.SelectArticles(
		r.Context(),
		currentUserId,
		tag,
		author,
		favorited,
		limit,
		offset,
	)
	if err != nil {
		log.Printf("Error selecting articles: %v", err)
		WriteJSON(w, http.StatusUnprocessableEntity, ErrorResponse{Error: err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, ArticlesResponse{
		Articles:      articles,
		ArticlesCount: articlesCount,
		Limit:         limit,
		Offset:        offset,
	})
}

func (s *Server) GetFeedHandler(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := r.Context().Value(CtxUserKey).(User)
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "cannot authorize user in jwt"})
		return
	}

	// TODO: allow filter feed?
	_, _, _, limit, offset := SearchQueries(w, r)

	articles, articlesCount, err := s.db.SelectArticlesFeed(
		r.Context(),
		currentUser.Id,
		limit,
		offset,
	)
	if err != nil {
		log.Printf("Error selecting feeed articles: %v", err)
		WriteJSON(w, http.StatusUnprocessableEntity, ErrorResponse{Error: err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, ArticlesResponse{
		Articles:      articles,
		ArticlesCount: articlesCount,
		Limit:         limit,
		Offset:        offset,
	})
}

func (s *Server) GetArticleHandler(w http.ResponseWriter, r *http.Request) {
	isAuth := r.Context().Value(CtxIsAuthKey).(bool)
	currentUser, ok := r.Context().Value(CtxUserKey).(User)
	if !ok && isAuth {
		WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "cannot authorize user in jwt"})
		return
	}

	vars := mux.Vars(r)
	slug, ok := vars["slug"]
	if !ok {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "slug is required"})
		return
	}

	var currentUserId string
	if isAuth {
		currentUserId = currentUser.Id
	}

	articleDetail, err := s.db.SelectArticleDetail(
		r.Context(),
		currentUserId,
		slug,
	)
	if errors.Is(err, sql.ErrNoRows) {
		WriteJSON(w, http.StatusNotFound, ErrorResponse{Error: "Article not found"})
		return
	}
	if err != nil {
		log.Printf("Error selecting article: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, ArticleDetailResponse{
		Article: *articleDetail,
	})
}

func (s *Server) PutArticleHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var body ArticleUpdateRequest

	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Printf("Error decode request body: %v", err)
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	currentUser, ok := r.Context().Value(CtxUserKey).(User)
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "cannot authorize user in jwt"})
		return
	}

	vars := mux.Vars(r)
	slug, ok := vars["slug"]
	if !ok {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "slug is required"})
		return
	}

	article, err := s.db.SelectArticle(
		r.Context(),
		slug,
	)
	if errors.Is(err, sql.ErrNoRows) {
		WriteJSON(w, http.StatusNotFound, ErrorResponse{Error: "Article not found"})
		return
	}
	if err != nil {
		log.Printf("Error selecting article: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// authz
	if article.AuthorId != currentUser.Id {
		WriteJSON(w, http.StatusForbidden,
			ErrorResponse{Error: "Only article author can update it"},
		)
		return
	}

	// input validation and sanitization
	err = article.ValidateArticleUpdate(&body.Article)
	if err != nil {
		WriteJSON(w, http.StatusUnprocessableEntity, ErrorResponse{Error: err.Error()})
		return
	}

	newSlug, err := s.db.UpdateArticle(r.Context(), article)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	articleDetail, err := s.db.SelectArticleDetail(r.Context(), currentUser.Id, newSlug)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, ArticleDetailResponse{Article: *articleDetail})
}

func (s *Server) DeleteArticleHandler(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := r.Context().Value(CtxUserKey).(User)
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "cannot authorize user in jwt"})
		return
	}

	vars := mux.Vars(r)
	slug, ok := vars["slug"]
	if !ok {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "slug is required"})
		return
	}

	err := s.db.DeleteArticle(r.Context(), currentUser.Id, slug)
	if err != nil {
		log.Printf("Error deleting article: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	WriteJSON(w, http.StatusNoContent, nil)
}

func (s *Server) PostFavoriteHandler(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := r.Context().Value(CtxUserKey).(User)
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "cannot authorize user in jwt"})
		return
	}

	vars := mux.Vars(r)
	slug, ok := vars["slug"]
	if !ok {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "slug is required"})
		return
	}

	err := s.db.CreateFavorite(r.Context(), currentUser.Id, slug)
	if err != nil {
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			log.Printf("Error creating favorite: %v", err)
			WriteJSON(w, http.StatusInternalServerError,
				ErrorResponse{Error: err.Error()},
			)
			return
		}

		if pgErr.Code != "23505" {
			log.Printf("Error creating favorite: %v", err)
			WriteJSON(w, http.StatusInternalServerError,
				ErrorResponse{Error: err.Error()},
			)
			return
		}

		// skip error if it's unique constraint violation, to behave like unfavorite
	}

	articleDetail, err := s.db.SelectArticleDetail(r.Context(), currentUser.Id, slug)
	if err != nil {
		log.Printf("Error selecting article detail: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, ArticleDetailResponse{Article: *articleDetail})
}

func (s *Server) DeleteFavoriteHandler(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := r.Context().Value(CtxUserKey).(User)
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "cannot authorize user in jwt"})
		return
	}

	vars := mux.Vars(r)
	slug, ok := vars["slug"]
	if !ok {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "slug is required"})
		return
	}

	err := s.db.DeleteFavorite(r.Context(), currentUser.Id, slug)
	if err != nil {
		log.Printf("Error creating favorite: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	articleDetail, err := s.db.SelectArticleDetail(r.Context(), currentUser.Id, slug)
	if err != nil {
		log.Printf("Error selecting article detail: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, ArticleDetailResponse{Article: *articleDetail})
}

func (s *Server) GetCommentsHandler(w http.ResponseWriter, r *http.Request) {
	isAuth := r.Context().Value(CtxIsAuthKey).(bool)
	currentUser, ok := r.Context().Value(CtxUserKey).(User)
	if !ok && isAuth {
		WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "cannot authorize user in jwt"})
		return
	}

	var currentUserId string
	if isAuth {
		currentUserId = currentUser.Id
	}

	_, _, _, limit, offset := SearchQueries(w, r)

	vars := mux.Vars(r)
	slug, ok := vars["slug"]
	if !ok {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "slug is required"})
		return
	}

	comments, commentsCount, err := s.db.SelectComments(
		r.Context(),
		currentUserId,
		slug,
		limit,
		offset,
	)
	if err != nil {
		log.Printf("Error selecting comments: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, CommentsResponse{
		Comments:      comments,
		CommentsCount: commentsCount,
		Limit:         limit,
		Offset:        offset,
	})

}

func (s *Server) PostCommentsHandler(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := r.Context().Value(CtxUserKey).(User)
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "cannot authorize user in jwt"})
		return
	}

	vars := mux.Vars(r)
	slug, ok := vars["slug"]
	if !ok {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "slug is required"})
		return
	}

	var body CommentCreateRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Printf("Error decode request body: %v", err)
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err = body.Comment.Validate()
	if err != nil {
		WriteJSON(w, http.StatusUnprocessableEntity, ErrorResponse{Error: err.Error()})
		return
	}

	commentId, err := s.db.CreateComment(r.Context(), currentUser.Id, slug, body.Comment.Body)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23502" { // not null constraint violation
				WriteJSON(w, http.StatusUnprocessableEntity,
					ErrorResponse{Error: "Article not found or invalid slug"},
				)
				return
			}
		}
		log.Printf("Error inserting comment: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	commentDetail, err := s.db.SelectCommentDetail(r.Context(), currentUser.Id, commentId)
	if err != nil {
		log.Printf("Error selecting comment detail: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, CommentDetailResponse{Comment: *commentDetail})
}

func (s *Server) DeleteCommentsHandler(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := r.Context().Value(CtxUserKey).(User)
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "cannot authorize user in jwt"})
		return
	}

	vars := mux.Vars(r)
	slug, ok := vars["slug"]
	if !ok {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "slug is required"})
		return
	}
	commentId, ok := vars["commentId"]
	if !ok {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "commentId is required"})
		return
	}

	err := s.db.DeleteComment(r.Context(), currentUser.Id, slug, commentId)
	if err != nil {
		if err.Error() == "zero rows affected" {
			WriteJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		log.Printf("Error deleting comment: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	WriteJSON(w, http.StatusNoContent, nil)
}

func (s *Server) GetProfilehandler(w http.ResponseWriter, r *http.Request) {
	isAuth := r.Context().Value(CtxIsAuthKey).(bool)
	currentUser, ok := r.Context().Value(CtxUserKey).(User)
	if !ok && isAuth {
		WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "cannot authorize user in jwt"})
		return
	}

	vars := mux.Vars(r)
	followingUsername, ok := vars["username"]
	if !ok {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "username is required"})
		return
	}

	// TODO: add concurrency or use one single query instead of two
	followingUser, err := s.db.SelectUser(r.Context(), "", "", followingUsername)
	if errors.Is(err, sql.ErrNoRows) {
		WriteJSON(w, http.StatusNotFound, ErrorResponse{Error: "username not found"})
		return
	}
	if err != nil {
		log.Printf("Error selecting following user: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	if isAuth {
		following, err := s.db.IsFollowing(r.Context(), currentUser.Id, followingUsername)
		if err != nil {
			log.Printf("Error checking if following: %v", err)
			WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		profilePreview := followingUser.ToProfilePreview(following)
		WriteJSON(w, http.StatusOK, ProfilePreviewResponse{Profile: *profilePreview})
		return
	}

	profilePreview := followingUser.ToProfilePreview(false)
	WriteJSON(w, http.StatusOK, ProfilePreviewResponse{Profile: *profilePreview})
}

func (s *Server) PostFollowHandler(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := r.Context().Value(CtxUserKey).(User)
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "cannot authorize user in jwt"})
		return
	}

	vars := mux.Vars(r)
	followingUsername, ok := vars["username"]
	if !ok {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "username not found"})
		return
	}

	// TODO: add concurrency or use one single query instead of three
	err := s.db.CreateFollow(r.Context(), currentUser.Id, followingUsername)
	if err != nil {
		log.Printf("Error creating follow: %v", err)
		WriteJSON(w, http.StatusUnprocessableEntity, ErrorResponse{Error: err.Error()})
		return
	}

	followingUser, err := s.db.SelectUser(r.Context(), "", "", followingUsername)
	if errors.Is(err, sql.ErrNoRows) {
		WriteJSON(w, http.StatusNotFound, ErrorResponse{Error: "username not found"})
		return
	}
	if err != nil {
		log.Printf("Error selecting following user: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	following, err := s.db.IsFollowing(r.Context(), currentUser.Id, followingUsername)
	if err != nil {
		log.Printf("Error checking if following: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	profilePreview := followingUser.ToProfilePreview(following)
	WriteJSON(w, http.StatusOK, ProfilePreviewResponse{Profile: *profilePreview})
}

func (s *Server) DeleteFollowHandler(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := r.Context().Value(CtxUserKey).(User)
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "cannot authorize user in jwt"})
		return
	}

	vars := mux.Vars(r)
	followingUsername, ok := vars["username"]
	if !ok {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "username not found"})
		return
	}

	// TODO: add concurrency or use one single query instead of three
	err := s.db.DeleteFollow(r.Context(), currentUser.Id, followingUsername)
	if err != nil {
		log.Printf("Error deleting follow: %v", err)
		WriteJSON(w, http.StatusUnprocessableEntity, ErrorResponse{Error: err.Error()})
		return
	}

	followingUser, err := s.db.SelectUser(r.Context(), "", "", followingUsername)
	if errors.Is(err, sql.ErrNoRows) {
		WriteJSON(w, http.StatusNotFound, ErrorResponse{Error: "username not found"})
		return
	}
	if err != nil {
		log.Printf("Error selecting following user: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	following, err := s.db.IsFollowing(r.Context(), currentUser.Id, followingUsername)
	if err != nil {
		log.Printf("Error checking if following: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	profilePreview := followingUser.ToProfilePreview(following)
	WriteJSON(w, http.StatusOK, ProfilePreviewResponse{Profile: *profilePreview})
}

func (s *Server) GetTagsHandler(w http.ResponseWriter, r *http.Request) {
	_, _, _, limit, offset := SearchQueries(w, r)

	tags, tagsCount, err := s.db.SelectTags(r.Context(), limit, offset)
	if err != nil {
		log.Printf("Error selecting tags: %v", err)
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, TagsResponse{
		Tags:      tags,
		TagsCount: tagsCount,
		Limit:     limit,
		Offset:    offset,
	})
}
