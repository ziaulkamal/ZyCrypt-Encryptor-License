package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/ziaulkamal/zycrypt/api/handler"
	apimiddleware "github.com/ziaulkamal/zycrypt/api/middleware"
	"github.com/ziaulkamal/zycrypt/config"
	"github.com/ziaulkamal/zycrypt/db"
	"github.com/ziaulkamal/zycrypt/internal/domain"
	"github.com/ziaulkamal/zycrypt/internal/license"
)

func Start() error {
	r := chi.NewRouter()

	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(apimiddleware.RateLimit(config.C.RateLimit.RequestsPerMinute))

	licSvc := license.NewService(db.DB, config.C.Security.SharedSecret)
	domSvc := domain.NewService(db.DB)

	vh := handler.NewValidateHandler(licSvc, domSvc, config.C.Security.SharedSecret, config.C.Security.TokenTTLMinutes)
	ph := handler.NewPingHandler()

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/ping", ph.Ping)
		r.Post("/validate", vh.Validate)
	})

	return http.ListenAndServe(config.Addr(), r)
}
