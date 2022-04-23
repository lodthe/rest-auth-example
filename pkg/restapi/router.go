package restapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/lodthe/rest-auth-example/internal/auth"
	"github.com/lodthe/rest-auth-example/internal/muser"
	"github.com/lodthe/rest-auth-example/internal/statstask"
	"github.com/lodthe/rest-auth-example/internal/taskqueue"
)

func NewRouter(authService *auth.Service, userRepo muser.Repository, taskRepo statstask.Repository, producer *taskqueue.Producer, timeout time.Duration) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(timeout))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api", func(r chi.Router) {
		newAuthHandler(authService, userRepo).handle(r)
	})

	return r
}
