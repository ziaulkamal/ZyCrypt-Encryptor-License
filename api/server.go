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
	"github.com/ziaulkamal/zycrypt/internal/theme"
)

func Start() error {
	r := chi.NewRouter()

	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(apimiddleware.RateLimit(config.C.RateLimit.RequestsPerMinute))

	secret  := config.C.Security.SharedSecret
	ttlMin  := config.C.Security.TokenTTLMinutes

	licSvc   := license.NewService(db.DB, secret)
	domSvc   := domain.NewService(db.DB)
	themeSvc := theme.NewService(db.DB, secret)

	ph  := handler.NewPingHandler()
	vh  := handler.NewValidateHandler(licSvc, domSvc, secret, ttlMin)
	th  := handler.NewThemesHandler(licSvc, themeSvc, secret, ttlMin)
	ah  := handler.NewActivateHandler(licSvc, themeSvc, secret, ttlMin)
	dh  := handler.NewDownloadHandler(themeSvc)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/ping",                ph.Ping)
		r.Post("/validate",           vh.Validate)
		r.Post("/themes",             th.ListThemes)
		r.Post("/activate",           ah.Activate)
		r.Get("/download/{token}",    dh.Download)
	})

	return http.ListenAndServe(config.Addr(), r)
}
