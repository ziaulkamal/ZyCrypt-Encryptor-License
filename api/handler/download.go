package handler

import (
	"encoding/base64"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ziaulkamal/zycrypt/internal/theme"
)

type DownloadHandler struct {
	themeSvc *theme.Service
}

func NewDownloadHandler(themeSvc *theme.Service) *DownloadHandler {
	return &DownloadHandler{themeSvc: themeSvc}
}

// Download handles GET /api/v1/download/{token}
// Streams the AES-256-CBC encrypted ZIP bundle. Token is single-use.
func (h *DownloadHandler) Download(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, "invalid_request", "token is required")
		return
	}

	encrypted, thm, err := h.themeSvc.RedeemToken(token)
	if err != nil {
		writeError(w, "token_invalid", err.Error())
		return
	}

	encoded := base64.StdEncoding.EncodeToString(encrypted)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="`+thm.Slug+`.bundle"`)
	w.Header().Set("X-ZyCrypt-Theme", thm.Slug)
	w.Header().Set("X-ZyCrypt-Version", thm.Version)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(encoded))
}
