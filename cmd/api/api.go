package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/kidusshun/planLog/service/user"
)

type APIServer struct {
	addr string
	db   *sql.DB
}

func NewAPIServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
	}
}

func (s *APIServer) Run() error {
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000"}, // Allow specific origin
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Content-Type", "Authorization"},
        ExposedHeaders:   []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           300, // Max cache duration in seconds
    }))
	router.Use(middleware.Logger)

	userStore := user.NewStore(s.db)
	userHandler := user.NewHandler(userStore)
	userHandler.RegisterRoutes(router)

	log.Println("Listening on ", s.addr)
	err := http.ListenAndServe(s.addr, router)

	return err
}
