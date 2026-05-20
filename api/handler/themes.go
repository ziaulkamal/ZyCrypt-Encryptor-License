package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	zcrypto "github.com/ziaulkamal/zycrypt/crypto"
	"github.com/ziaulkamal/zycrypt/internal/license"
	"github.com/ziaulkamal/zycrypt/internal/theme"
)

type ThemesHandler struct {
	licSvc   *license.Service
	themeSvc *theme.Service
	secret   string
	ttlMin   int
}

func NewThemesHandler(licSvc *license.Service, themeSvc *theme.Service, secret string, ttlMin int) *ThemesHandler {
	return &ThemesHandler{licSvc: licSvc, themeSvc: themeSvc, secret: secret, ttlMin: ttlMin}
}

type themeRequest struct {
	LicenseKey string `json:"license_key"`
	Domain     string `json:"domain"`
	Timestamp  int64  `json:"timestamp"`
	Signature  string `json:"signature"`
}

func (h *ThemesHandler) verifyRequest(w http.ResponseWriter, r *http.Request) (*license.License, bool) {
	var req themeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid_request", "malformed JSON body")
		return nil, false
	}

	sigData := fmt.Sprintf("%s:%s:%d", req.LicenseKey, req.Domain, req.Timestamp)
	if !zcrypto.VerifyHMAC(h.secret, sigData, req.Signature) {
		writeError(w, "invalid_signature", "HMAC signature verification failed")
		return nil, false
	}

	now := time.Now().Unix()
	ttl := int64(h.ttlMin * 60)
	if now-req.Timestamp > ttl || req.Timestamp > now+60 {
		writeError(w, "token_expired", "request timestamp is too old or in the future")
		return nil, false
	}

	lic, err := h.licSvc.GetByKey(req.LicenseKey)
	if err != nil {
		writeError(w, "license_not_found", "license key does not exist")
		return nil, false
	}

	if lic.Status == license.StatusBanned {
		writeError(w, "license_banned", "")
		return nil, false
	}
	if lic.Status == license.StatusInactive {
		writeError(w, "license_inactive", "")
		return nil, false
	}
	if !lic.IsLifetime && lic.ExpiresAt != nil && time.Now().After(*lic.ExpiresAt) {
		writeError(w, "license_expired", "")
		return nil, false
	}

	return lic, true
}

// ListThemes handles GET /api/v1/themes
// Returns all themes available to the license's plan.
func (h *ThemesHandler) ListThemes(w http.ResponseWriter, r *http.Request) {
	lic, ok := h.verifyRequest(w, r)
	if !ok {
		return
	}

	themes, err := h.themeSvc.ThemesForPlan(lic.PlanID)
	if err != nil {
		writeError(w, "server_error", err.Error())
		return
	}

	type themeItem struct {
		Slug     string `json:"slug"`
		Name     string `json:"name"`
		Version  string `json:"version"`
		FileSize int64  `json:"file_size"`
	}

	items := make([]themeItem, 0, len(themes))
	for _, t := range themes {
		items = append(items, themeItem{
			Slug:     t.Slug,
			Name:     t.Name,
			Version:  t.Version,
			FileSize: t.FileSize,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"themes": items,
		"plan":   lic.Plan.Slug,
	})
}
