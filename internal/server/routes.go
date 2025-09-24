package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/minhhoccode111/realworldgo/internal/middleware"
	"github.com/minhhoccode111/realworldgo/internal/model"
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

	auth := middleware.AuthMiddleware(s.config.JWT.Secret, s.db)

	r.HandleFunc("/websocket", s.websocketHandler)

	r.HandleFunc("/users", s.RegisterHandler).Methods("POST")
	r.HandleFunc("/users/login", s.LoginHandler).Methods("POST")

	r.HandleFunc("/user", auth(s.GetUserHandler)).Methods("GET")
	r.HandleFunc("/user", auth(s.PutUserHandler)).Methods("PUT")

	r.HandleFunc("/articles", auth(s.PostArticleHandler)).Methods("POST")
	r.HandleFunc("/articles", s.GetAllArticlesHandler).Methods("GET")
	r.HandleFunc("/articles/feed", auth(s.GetFeedHandler)).Methods("GET")
	r.HandleFunc("/articles/{slug}", s.GetArticleHandler).Methods("GET")
	r.HandleFunc("/articles/{slug}", auth(s.PutArticleHandler)).Methods("PUT")
	r.HandleFunc("/articles/{slug}", auth(s.DeleteArticleHandler)).Methods("DELETE")

	r.HandleFunc("/articles/{slug}/favorite", auth(s.PostFavoriteHandler)).Methods("POST")
	r.HandleFunc("/articles/{slug}/favorite", auth(s.DeleteFavoriteHandler)).Methods("DELETE")

	r.HandleFunc("/articles/{slug}/comments", s.GetCommentsHandler).Methods("GET")
	r.HandleFunc("/articles/{slug}/comments", auth(s.PostCommentsHandler)).Methods("POST")
	r.HandleFunc("/articles/{slug}/comments/{id}", auth(s.DeleteCommentsHandler)).Methods("DELETE")

	r.HandleFunc("/profiles/{username}", s.GetProfilehandler).Methods("GET")
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
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("Error decode request body: %v", err)
		WriteJSON(w, http.StatusBadRequest, JSON{"error": err.Error()})
		return
	}
	email, err := IsValidEmail(body.Email)
	if err != nil {
		log.Printf("Input Email Error: %v", err)
		WriteJSON(w, http.StatusBadRequest, JSON{"error": err.Error()})
		return
	}
	password, err := IsValidPassword(body.Password)
	if err != nil {
		log.Printf("Error: %v", err)
		WriteJSON(w, http.StatusBadRequest, JSON{"error": err.Error()})
		return
	}
	userExisted, err := s.db.SelectUserByEmail(r.Context(), email)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error: %v", err)
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	}
	if userExisted != nil {
		WriteJSON(w, http.StatusConflict, JSON{"error": "email already existed"})
		return
	}
	user := model.User{
		Email:    email,
		Password: password,
		// IsActive: true,
		// Role:     model.RoleUser,
	}
	err = s.db.InsertUser(r.Context(), &user)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	}
	userDTO := model.UserToUserDTO(&user)
	token, err := GenerateJWT(s.config.JWT, &userDTO)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	}
	WriteJSON(w, http.StatusCreated, JSON{"user": userDTO, "token": token})
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("Error decode request body: %v", err)
		WriteJSON(w, http.StatusBadRequest, JSON{"error": err.Error()})
		return
	}

	userExisted, err := s.db.SelectUserByEmail(r.Context(), body.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteJSON(w, http.StatusUnauthorized, JSON{"error": "email not found"})
			return
		}
		log.Printf("Error: %v", err)
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	}
	if !ValidatePassword(userExisted.Password, body.Password) {
		WriteJSON(w, http.StatusUnauthorized, JSON{"error": "password incorrect"})
		return
	}
	userDTO := model.UserToUserDTO(userExisted)
	token, err := GenerateJWT(s.config.JWT, &userDTO)
	if err != nil {
		log.Printf("Error: %v", err)
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	}
	WriteJSON(w, http.StatusOK, JSON{"user": userDTO, "token": token})
}

func (s *Server) GetUserHandler(w http.ResponseWriter, r *http.Request)        {}
func (s *Server) PutUserHandler(w http.ResponseWriter, r *http.Request)        {}
func (s *Server) PostArticleHandler(w http.ResponseWriter, r *http.Request)    {}
func (s *Server) GetAllArticlesHandler(w http.ResponseWriter, r *http.Request) {}
func (s *Server) GetFeedHandler(w http.ResponseWriter, r *http.Request)        {}
func (s *Server) GetArticleHandler(w http.ResponseWriter, r *http.Request)     {}
func (s *Server) PutArticleHandler(w http.ResponseWriter, r *http.Request)     {}
func (s *Server) DeleteArticleHandler(w http.ResponseWriter, r *http.Request)  {}
func (s *Server) PostFavoriteHandler(w http.ResponseWriter, r *http.Request)   {}
func (s *Server) DeleteFavoriteHandler(w http.ResponseWriter, r *http.Request) {}
func (s *Server) GetCommentsHandler(w http.ResponseWriter, r *http.Request)    {}
func (s *Server) PostCommentsHandler(w http.ResponseWriter, r *http.Request)   {}
func (s *Server) DeleteCommentsHandler(w http.ResponseWriter, r *http.Request) {}
func (s *Server) GetProfilehandler(w http.ResponseWriter, r *http.Request)     {}
func (s *Server) PostFollowHandler(w http.ResponseWriter, r *http.Request)     {}
func (s *Server) DeleteFollowHandler(w http.ResponseWriter, r *http.Request)   {}
func (s *Server) GetTagsHandler(w http.ResponseWriter, r *http.Request)        {}
