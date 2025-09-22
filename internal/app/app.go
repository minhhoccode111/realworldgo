package app

import (
	"auth/internal/config"
	"auth/internal/handlers"
	"auth/internal/repositories"
	"auth/internal/services"
	"database/sql"
	"net/http"
)

type App struct {
	config     *config.Config
	db         *sql.DB
	server     *http.Server
	handler    *handlers.Handlers
	service    *services.Services
	repository *repositories.Repositories
}
