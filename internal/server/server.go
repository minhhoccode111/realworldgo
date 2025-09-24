package server

import (
	"log"
	"net/http"

	"github.com/minhhoccode111/realworldgo/internal/config"
	"github.com/minhhoccode111/realworldgo/internal/database"
)

type Server struct {
	config *config.Config
	db     database.Service
}

func NewServer() *http.Server {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	server := &Server{
		config: cfg,
		db:     database.New(cfg.Database.DatabaseURL()),
	}

	httpServer := &http.Server{
		Addr:           cfg.Server.ServerAddress(),
		Handler:        server.RegisterRoutes(),
		IdleTimeout:    cfg.Server.IdleTimeout,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		MaxHeaderBytes: 1 << 10,
	}

	return httpServer
}
