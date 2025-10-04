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
	"github.com/minhhoccode111/realworldgo/internal/model"
	"github.com/minhhoccode111/realworldgo/internal/utils"
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
	r.HandleFunc("/articles/{slug}", s.GetArticleHandler).Methods("GET")
	r.HandleFunc("/articles/{slug}", auth(s.PutArticleHandler)).Methods("PUT")
	r.HandleFunc("/articles/{slug}", auth(s.DeleteArticleHandler)).Methods("DELETE")

	r.HandleFunc("/articles/{slug}/favorite", auth(s.PostFavoriteHandler)).Methods("POST")
	r.HandleFunc("/articles/{slug}/favorite", auth(s.DeleteFavoriteHandler)).Methods("DELETE")

	r.HandleFunc("/articles/{slug}/comments", s.GetCommentsHandler).Methods("GET")
	r.HandleFunc("/articles/{slug}/comments", auth(s.PostCommentsHandler)).Methods("POST")
	r.HandleFunc("/articles/{slug}/comments/{id}", auth(s.DeleteCommentsHandler)).Methods("DELETE")

	r.HandleFunc("/profiles/{username}", optionalAuth(s.GetProfilehandler)).Methods("GET")
	r.HandleFunc("/profiles/{username}/follow", auth(s.PostFollowHandler)).Methods("POST")
	r.HandleFunc("/profiles/{username}/follow", auth(s.DeleteFollowHandler)).Methods("DELETE")

	r.HandleFunc("/tags", s.GetTagsHandler).Methods("GET")
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, JSON{"message": "Hello, World!"})
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, s.db.Health())
}

func (s *Server) websocketHandler(w http.ResponseWriter, r *http.Request) {
	socket, err := websocket.Accept(w, r, nil)

	if err != nil {
		log.Printf("could not open websocket: %v", err)
		_, _ = w.Write([]byte("could not open websocket"))
		w.WriteHeader(http.StatusInternalServerError)
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
	var body struct {
		User model.UserRegisterRequest `json:"user"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("Error decode request body: %v", err)
		WriteJSON(w, http.StatusBadRequest, JSON{"error": err.Error()})
		return
	}

	err := body.User.Validate()
	if err != nil {
		log.Printf("UserRegisterRequest model failed validations: %v", err)
		WriteJSON(w, http.StatusUnprocessableEntity, JSON{"error": err.Error()})
		return
	}

	// WARN: don't check for uniqueness manually, because race conditions can occur
	// we must check for uniqueness if error returned after we insert to db
	// userExisted, err := s.db.SelectUserByEmail(r.Context(), body.User.Email)

	newUser := model.User{
		Email:    body.User.Email,
		Username: body.User.Username,
		Password: body.User.Password,
	}

	err = s.db.CreateUser(r.Context(), &newUser)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				WriteJSON(w, http.StatusConflict, JSON{"error": err.Error()})
				return
			}
		}

		log.Printf("Error inserting user: %v", err)
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	}

	token, err := GenerateJWT(s.config.JWT, newUser.Id)
	if err != nil {
		log.Printf("Error generating jwt: %v", err)
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	}

	userResponse := newUser.ToUserResponse(token)
	WriteJSON(w, http.StatusCreated, JSON{"user": userResponse})
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		User model.UserLoginRequest `json:"user"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("Error decode request body: %v", err)
		WriteJSON(w, http.StatusBadRequest, JSON{"error": err.Error()})
		return
	}

	userExisted, err := s.db.SelectUser(r.Context(), "", body.User.Email, "")
	if err != nil {
		if err == sql.ErrNoRows {
			WriteJSON(w, http.StatusUnauthorized, JSON{"error": "email not found"})
			return
		}
		log.Printf("Error selecting user: %v", err)
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	}
	if !ValidatePassword(userExisted.Password, body.User.Password) {
		WriteJSON(w, http.StatusUnauthorized, JSON{"error": "password incorrect"})
		return
	}

	token, err := GenerateJWT(s.config.JWT, userExisted.Id)
	if err != nil {
		log.Printf("Error generating jwt: %v", err)
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	}

	userResponse := userExisted.ToUserResponse(token)
	WriteJSON(w, http.StatusOK, JSON{"user": userResponse})
}

func (s *Server) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(CtxUserKey).(model.User)
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, JSON{"error": "cannot authorize user in jwt"})
		return
	}
	token, err := GenerateJWT(s.config.JWT, user.Id)
	if err != nil {
		log.Printf("Error generating jwt: %v", err)
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	}

	userResponse := user.ToUserResponse(token)
	WriteJSON(w, http.StatusOK, JSON{"user": userResponse})
}

func (s *Server) PutUserHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var body struct {
		User model.UserUpdateRequest `json:"user"`
	}
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("Error decode request body: %v", err)
		WriteJSON(w, http.StatusBadRequest, JSON{"error": err.Error()})
		return
	}

	currentUser, ok := r.Context().Value(CtxUserKey).(model.User)
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, JSON{"error": "cannot authorize user in jwt"})
		return
	}

	err = currentUser.ValidateUserUpdateRequest(&body.User)
	if err != nil {
		log.Printf("Error validating user update request: %v", err)
		WriteJSON(w, http.StatusUnprocessableEntity, JSON{"error": err.Error()})
		return
	}

	err = s.db.UpdateUser(r.Context(), &currentUser)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	}

	token, err := GenerateJWT(s.config.JWT, currentUser.Id)
	userResponse := currentUser.ToUserResponse(token)
	WriteJSON(w, http.StatusOK, JSON{"user": userResponse})
}

func (s *Server) PostArticleHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var body struct {
		Article model.ArticleCreateRequest `json:"article"`
	}

	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("Error decode request body: %v", err)
		WriteJSON(w, http.StatusBadRequest, JSON{"error": err.Error()})
		return
	}

	currentUser, ok := r.Context().Value(CtxUserKey).(model.User)
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, JSON{"error": "cannot authorize user in jwt"})
		return
	}

	err = body.Article.Validate()
	if err != nil {
		log.Printf("Error validating article create request: %v", err)
		WriteJSON(w, http.StatusUnprocessableEntity, JSON{"error": err.Error()})
		return
	}

	newArticle := model.Article{
		AuthorId:    currentUser.Id,
		Title:       body.Article.Title,
		Body:        body.Article.Body,
		Description: body.Article.Description,
	}

	err = s.db.CreateArticle(r.Context(), &newArticle, body.Article.TagList)
	if err != nil {
		log.Printf("Error creating article: %v", err)
		WriteJSON(w, http.StatusUnprocessableEntity, JSON{"error": err.Error()})
		return
	}

	var following bool
	following, err = s.db.IsFollowing(r.Context(), currentUser.Id, currentUser.Username)

	articleResponse := newArticle.ToArticleDetailResponse(
		*currentUser.ToProfilePreviewResponse(following),
		body.Article.TagList,
	)
	WriteJSON(w, http.StatusOK, JSON{"article": articleResponse})
}

func (s *Server) GetAllArticlesHandler(w http.ResponseWriter, r *http.Request) {
	isAuth := r.Context().Value(CtxIsAuthKey).(bool)
	currentUser, ok := r.Context().Value(CtxUserKey).(model.User)
	if !ok && isAuth {
		WriteJSON(w, http.StatusUnauthorized, JSON{"error": "cannot authorize user in jwt"})
		return
	}

	var currentUserId string
	if isAuth {
		currentUserId = currentUser.Id
	}

	tag, author, favorited, limit, offset := utils.SearchQueries(w, r)

	ar, err := s.db.SelectArticles(
		r.Context(),
		currentUserId,
		tag,
		author,
		favorited,
		limit,
		offset,
	)
	if err != nil {
		log.Printf("Error selecting article: %v", err)
		WriteJSON(w, http.StatusUnprocessableEntity, JSON{"error": err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, *ar)
}

func (s *Server) GetFeedHandler(w http.ResponseWriter, r *http.Request)        {}
func (s *Server) GetArticleHandler(w http.ResponseWriter, r *http.Request)     {}
func (s *Server) PutArticleHandler(w http.ResponseWriter, r *http.Request)     {}
func (s *Server) DeleteArticleHandler(w http.ResponseWriter, r *http.Request)  {}
func (s *Server) PostFavoriteHandler(w http.ResponseWriter, r *http.Request)   {}
func (s *Server) DeleteFavoriteHandler(w http.ResponseWriter, r *http.Request) {}
func (s *Server) GetCommentsHandler(w http.ResponseWriter, r *http.Request)    {}
func (s *Server) PostCommentsHandler(w http.ResponseWriter, r *http.Request)   {}
func (s *Server) DeleteCommentsHandler(w http.ResponseWriter, r *http.Request) {}
func (s *Server) GetProfilehandler(w http.ResponseWriter, r *http.Request) {
	isAuth := r.Context().Value(CtxIsAuthKey).(bool)
	follower, ok := r.Context().Value(CtxUserKey).(model.User)
	if !ok && isAuth {
		WriteJSON(w, http.StatusUnauthorized, JSON{"error": "cannot authorize user in jwt"})
		return
	}

	vars := mux.Vars(r)
	followingUsername, ok := vars["username"]
	if !ok {
		WriteJSON(w, http.StatusBadRequest, JSON{"error": "username is required"})
		return
	}

	followingUser, err := s.db.SelectUser(r.Context(), "", "", followingUsername)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteJSON(w, http.StatusNoContent, JSON{"error": "username not found"})
			return
		}

		log.Printf("Error selecting following user: %v", err)
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	}

	if isAuth {
		following, err := s.db.IsFollowing(r.Context(), follower.Id, followingUsername)
		if err != nil {
			log.Printf("Error checking if following: %v", err)
			WriteJSON(w, http.StatusInternalServerError, err.Error())
			return
		}

		profileResponse := followingUser.ToProfilePreviewResponse(following)
		WriteJSON(w, http.StatusOK, JSON{"profile": profileResponse})
		return
	}

	profileResponse := followingUser.ToProfilePreviewResponse(false)
	WriteJSON(w, http.StatusOK, JSON{"profile": profileResponse})
}

func (s *Server) PostFollowHandler(w http.ResponseWriter, r *http.Request) {
	follower, ok := r.Context().Value(CtxUserKey).(model.User)
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, JSON{"error": "cannot authorize user in jwt"})
		return
	}

	vars := mux.Vars(r)
	followingUsername, ok := vars["username"]
	if !ok {
		WriteJSON(w, http.StatusBadRequest, JSON{"error": "username not found"})
		return
	}

	err := s.db.CreateFollow(r.Context(), follower.Id, followingUsername)
	if err != nil {
		log.Printf("Error creating follow: %v", err)
		WriteJSON(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	followingUser, err := s.db.SelectUser(r.Context(), "", "", followingUsername)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteJSON(w, http.StatusNoContent, JSON{"error": "username not found"})
			return
		}

		log.Printf("Error selecting following user: %v", err)
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	}

	following, err := s.db.IsFollowing(r.Context(), follower.Id, followingUsername)
	if err != nil {
		log.Printf("Error checking if following: %v", err)
		WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	profileResponse := followingUser.ToProfilePreviewResponse(following)
	WriteJSON(w, http.StatusOK, JSON{"profile": profileResponse})
}

func (s *Server) DeleteFollowHandler(w http.ResponseWriter, r *http.Request) {
	follower, ok := r.Context().Value(CtxUserKey).(model.User)
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, JSON{"error": "cannot authorize user in jwt"})
		return
	}

	vars := mux.Vars(r)
	followingUsername, ok := vars["username"]
	if !ok {
		WriteJSON(w, http.StatusBadRequest, JSON{"error": "username not found"})
		return
	}

	err := s.db.DeleteFollow(r.Context(), follower.Id, followingUsername)
	if err != nil {
		log.Printf("Error deleting follow: %v", err)
		WriteJSON(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	followingUser, err := s.db.SelectUser(r.Context(), "", "", followingUsername)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteJSON(w, http.StatusNoContent, JSON{"error": "username not found"})
			return
		}

		log.Printf("Error selecting following user: %v", err)
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	}

	following, err := s.db.IsFollowing(r.Context(), follower.Id, followingUsername)
	if err != nil {
		log.Printf("Error checking if following: %v", err)
		WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	profileResponse := followingUser.ToProfilePreviewResponse(following)
	WriteJSON(w, http.StatusOK, JSON{"profile": profileResponse})
}

func (s *Server) GetTagsHandler(w http.ResponseWriter, r *http.Request) {}
