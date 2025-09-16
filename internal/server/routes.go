package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	// to access without a qualifier
	"auth/internal/model"
	"auth/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/coder/websocket"
)

type JSON map[string]any
type ctxKey string

const (
	ctxUserKey   ctxKey = "user"
	ctxUserIdKey ctxKey = "userId"
)

func WriteText(w http.ResponseWriter, status int, data string) {
	w.Header().Set("Content-Type", "application/text")
	w.WriteHeader(status)
	w.Write([]byte(data))
}

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) RegisterRoutes() http.Handler {
	r := mux.NewRouter()

	// Apply global middleware
	r.Use(s.timeoutMiddleware)
	r.Use(s.corsMiddleware)

	// Root level routes (no versioning)
	r.HandleFunc("/", s.HelloWorldHandler)
	r.HandleFunc("/healthz", s.healthHandler)

	// API v1 routes
	v1 := r.PathPrefix("/api/v1").Subrouter()
	s.registerV1Routes(v1)

	return r
}

func (s *Server) registerV1Routes(r *mux.Router) {
	// Public routes
	r.HandleFunc("/websocket", s.websocketHandler)
	r.HandleFunc("/auth/register", s.RegisterHandler).Methods("POST")
	r.HandleFunc("/auth/login", s.LoginHandler).Methods("POST")
	// WARN: must define '/all' before '/{id}'
	r.HandleFunc("/users/all", s.GetAllUsersHandler).Methods("GET")
	r.HandleFunc("/users/{id}", s.GetUserHandler).Methods("GET")

	// User-authenticated routes
	me := r.PathPrefix("/auth").Subrouter()
	me.Use(s.authMiddleware)
	me.HandleFunc("/me", s.GetMeHandler).Methods("GET")
	user := r.PathPrefix("/users").Subrouter()
	user.Use(s.authMiddleware)
	user.HandleFunc("/{id}", s.UpdateUserHandler).Methods("PATCH")
	user.HandleFunc("/{id}/password", s.PasswordUserHandler).Methods("PATCH")
	// WARN: user can deactivate their account but only admin can activate an account
	user.HandleFunc("/{id}/status", s.StatusUserHandler).Methods("PATCH")

	// Admin-authorized routes
	admin := r.PathPrefix("/users").Subrouter()
	admin.Use(s.authMiddleware)
	admin.Use(s.adminMiddleware)
	admin.HandleFunc("/{id}", s.DeleteUserHandler).Methods("DELETE")
}

// timeout middleware use context
func (s *Server) timeoutMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// WARN: only allow the request to be ran in 100 milliseconds, just playing around with context
		ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*150)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CORS middleware
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	allowedOrigins := s.config.CORS.AllowedOrigins

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if slices.Contains(allowedOrigins, "*") || slices.Contains(allowedOrigins, origin) {
			if slices.Contains(allowedOrigins, "*") {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Credentials", "false")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Authentication middleware
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get authentication header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			WriteJSON(w, http.StatusUnauthorized, JSON{"error": "no authorization header"})
			return
		}
		parts := strings.Fields(authHeader)
		if !strings.EqualFold(parts[0], "Bearer") {
			WriteJSON(w, http.StatusUnauthorized, JSON{"error": "authorization header must start with 'Bearer'"})
			return
		}
		if len(parts) != 2 {
			WriteJSON(w, http.StatusUnauthorized, JSON{"error": "authorization header must be formatted as 'Bearer <token>'"})
			return
		}
		tokenStr := parts[1]
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(s.config.JWT.Secret), nil
		})
		if err != nil {
			log.Printf("Error parsing token: %v", err)
			WriteJSON(w, http.StatusUnauthorized, JSON{"error": err.Error()})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			WriteJSON(w, http.StatusUnauthorized, JSON{"error": "invalid token"})
			return
		}
		userId, ok := claims[string(ctxUserIdKey)].(string)
		if !ok || userId == "" {
			WriteJSON(w, http.StatusUnauthorized, JSON{"error": "missing userId in token"})
			return
		}
		// pass context to database query
		user, err := s.db.SelectUserById(r.Context(), userId)
		if err != nil {
			log.Printf("Error selecting user by id: %v", err)
			if strings.Contains(err.Error(), "timeout") {
				WriteJSON(w, http.StatusUnauthorized, JSON{"error": err.Error()})
				return
			}
			WriteJSON(w, http.StatusUnauthorized, JSON{"error": "cannot authorize user in jwt"})
			return
		}
		if !user.IsActive {
			WriteJSON(w, http.StatusForbidden, JSON{"error": "user in jwt is inactive"})
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserKey, *user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Authorization middleware
func (s *Server) adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(ctxUserKey).(model.User)
		if user.Role != model.RoleAdmin {
			WriteJSON(w, http.StatusForbidden, JSON{"error": "user is not admin"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, JSON{"message": "Hello, World!"})
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, s.db.Health())
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
	email, err := utils.IsValidEmail(body.Email)
	if err != nil {
		log.Printf("Input Email Error: %v", err)
		WriteJSON(w, http.StatusBadRequest, JSON{"error": err.Error()})
		return
	}
	password, err := utils.IsValidPassword(body.Password)
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
		IsActive: true,
		Role:     model.RoleUser,
	}
	err = s.db.InsertUser(r.Context(), &user)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	}
	userDTO := utils.UserToUserDTO(&user)
	token, err := utils.GenerateJWT(s.config.JWT, &userDTO)
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
	if !utils.ValidatePassword(userExisted.Password, body.Password) {
		WriteJSON(w, http.StatusUnauthorized, JSON{"error": "password incorrect"})
		return
	}
	userDTO := utils.UserToUserDTO(userExisted)
	token, err := utils.GenerateJWT(s.config.JWT, &userDTO)
	if err != nil {
		log.Printf("Error: %v", err)
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	}
	WriteJSON(w, http.StatusOK, JSON{"user": userDTO, "token": token})
}

func (s *Server) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	paths := strings.Split(r.URL.Path, "/")
	userId := paths[len(paths)-1] // path/user/{userId}
	existedUser, err := s.db.SelectUserById(r.Context(), userId)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteJSON(w, http.StatusNotFound, JSON{"error": "user not found"})
			return
		}
		log.Printf("Error: %v", err)
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	}
	WriteJSON(w, http.StatusOK, utils.UserToUserDTO(existedUser))
}

func (s *Server) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	perPageStr := r.URL.Query().Get("perPage")
	pageNumberStr := r.URL.Query().Get("pageNumber")
	allStr := r.URL.Query().Get("all")
	filter := r.URL.Query().Get("q")
	var err error

	limit, err := strconv.Atoi(perPageStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	pageNumber, err := strconv.Atoi(pageNumberStr)
	if err != nil || pageNumber < 1 {
		pageNumber = 1
	}
	isGetAll := allStr == "true"
	offset := (pageNumber - 1) * limit

	divideAndRoundUp := func(a, b int) int {
		return (a + b - 1) / b // e.g. 10 / 3 = (10 + 3 - 1) / 3 = 4
	}

	usersCh := make(chan []*model.UserDTO)
	countCh := make(chan int)
	errCh := make(chan error, 2)
	defer close(errCh)

	var wg sync.WaitGroup
	wg.Add(2)
	defer wg.Wait()

	go func() {
		defer wg.Done()
		defer close(usersCh)
		users, err := s.db.SelectUsers(r.Context(), limit, offset, filter, isGetAll)
		if err != nil {
			errCh <- err
			return
		}
		usersCh <- users
	}()

	go func() {
		defer wg.Done()
		defer close(countCh)
		countUsers, err := s.db.CountUsers(r.Context(), filter, isGetAll)
		if err != nil {
			errCh <- err
			return
		}
		countCh <- countUsers
	}()

	var users []*model.UserDTO
	var count int

	select {
	case err := <-errCh:
		log.Printf("Error when get all users: %v", err)
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	default:
		users = <-usersCh
		count = <-countCh
	}

	WriteJSON(w, http.StatusOK, JSON{
		"users":      users,
		"totalPage":  divideAndRoundUp(count, limit),
		"perPage":    limit,
		"pageNumber": pageNumber,
	})
}

func (s *Server) GetMeHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ctxUserKey).(model.User)
	userDTO := utils.UserToUserDTO(&user)
	token, err := utils.GenerateJWT(s.config.JWT, &userDTO)
	if err != nil {
		log.Printf("Error: %v", err)
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	}
	WriteJSON(w, http.StatusOK, JSON{"token": token, "user": userDTO})
}

func (s *Server) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"` // NOTE: explicitly state what we will update
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("Error decode request body: %v", err)
		WriteJSON(w, http.StatusBadRequest, JSON{"error": err.Error()})
		return
	}
	email, err := utils.IsValidEmail(body.Email)
	if err != nil {
		log.Printf("Error: %v", err)
		WriteJSON(w, http.StatusBadRequest, JSON{"error": err.Error()})
		return
	}
	paths := strings.Split(r.URL.Path, "/")
	userIdPath := paths[len(paths)-1] // path/user/{userId}
	userIdToken := r.Context().Value(ctxUserKey).(model.User).Id
	if userIdPath != userIdToken {
		WriteJSON(w, http.StatusUnauthorized, JSON{"error": "userIdToken and userIdPath mismatch"})
		return
	}
	updatedUserDTO, err := s.db.UpdateUser(r.Context(), userIdPath, email)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				WriteJSON(w, http.StatusConflict, JSON{"error": "email already existed"})
				return
			}
		}
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		log.Printf("Error: %v", err)
		return
	}
	WriteJSON(w, http.StatusOK, updatedUserDTO)
}

func (s *Server) StatusUserHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		// NOTE: pointer to differentiate between explicit-false and not-provided
		IsActive *bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("Error decode request body: %v", err)
		WriteJSON(w, http.StatusBadRequest, JSON{"error": err.Error()})
		return
	}
	if body.IsActive == nil {
		WriteJSON(w, http.StatusBadRequest, JSON{"error": "is_active is required in request body"})
		return
	}
	paths := strings.Split(r.URL.Path, "/")
	userIdPath := paths[len(paths)-2] // path/users/{userId}/status
	userInToken := r.Context().Value(ctxUserKey).(model.User)
	userIdToken := userInToken.Id
	// admin can activate or deactivate any user, user can only deactivate itself
	if userInToken.Role != model.RoleAdmin {
		// activate
		if *body.IsActive {
			WriteJSON(w, http.StatusForbidden, JSON{"error": "only admin can activate a user"})
			return
		}
		// deactivate
		if userIdToken != userIdPath {
			WriteJSON(w, http.StatusForbidden, JSON{"error": "you must be admin to deactivate other users than yourself"})
			return
		}
		// fine to continue
	}
	err := s.db.UpdateUserStatus(r.Context(), userIdPath, *body.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteJSON(w, http.StatusUnauthorized, JSON{"error": "user to be set status not found"})
			return
		}
		log.Printf("Error: %v", err)
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) PasswordUserHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("Error decode request body: %v", err)
		WriteJSON(w, http.StatusBadRequest, JSON{"error": err.Error()})
		return
	}
	newPassword, err := utils.IsValidPassword(body.NewPassword)
	if err != nil {
		log.Printf("Error: %v", err)
		WriteJSON(w, http.StatusBadRequest, JSON{"error": err.Error()})
		return
	}
	paths := strings.Split(r.URL.Path, "/")
	userIdPath := paths[len(paths)-2] // path/users/{userId}/status
	userInToken := r.Context().Value(ctxUserKey).(model.User)
	userIdToken := userInToken.Id
	if userIdPath != userIdToken {
		WriteJSON(w, http.StatusForbidden, JSON{"error": "cannot change another user's password"})
		return
	}
	if !utils.ValidatePassword(userInToken.Password, body.OldPassword) {
		WriteJSON(w, http.StatusUnauthorized, JSON{"error": "old password is not correct"})
		return
	}
	err = s.db.UpdateUserPassword(r.Context(), userIdPath, newPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteJSON(w, http.StatusUnauthorized, JSON{"error": "user to be updated password not found"})
			return
		}
		log.Printf("Error: %v", err)
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	paths := strings.Split(r.URL.Path, "/")
	userIdPath := paths[len(paths)-1] // path/users/{userId}
	userIdToken := r.Context().Value(ctxUserKey).(model.User).Id
	if userIdPath == userIdToken {
		WriteJSON(w, http.StatusForbidden, JSON{"error": "admin cannot self-delete"})
		return
	}
	err := s.db.DeleteUserById(r.Context(), userIdPath)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteJSON(w, http.StatusNotFound, JSON{"error": "user to be deleted not found"})
			return
		}
		log.Printf("Error: %v", err)
		WriteJSON(w, http.StatusInternalServerError, JSON{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
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
