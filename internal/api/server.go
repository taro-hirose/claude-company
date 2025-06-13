package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"claude-company/internal/database"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router *gin.Engine
	server *http.Server
	config *database.Config
}

func NewServer(config *database.Config) *Server {
	return &Server{
		config: config,
		router: SetupRoutes(config),
	}
}

func (s *Server) Start(port string) error {
	if port == "" {
		port = "8080"
	}

	s.server = &http.Server{
		Addr:    ":" + port,
		Handler: s.router,
	}

	go func() {
		log.Printf("Starting Claude Company API server on port %s", port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	return s.Shutdown()
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	if err := database.CloseDB(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	log.Println("Server exited")
	return nil
}

func (s *Server) GetRouter() *gin.Engine {
	return s.router
}