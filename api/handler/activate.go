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

type ActivateHandler struct {
	licSvc   *license.Service
	themeSvc *theme.Service
	secret   string
	ttlMin   int
}

func NewActivateHandler(licSvc *license.Service, themeSvc *theme.Service, secret string, ttlMin int) *ActivateHandler {
	return &ActivateHandler{licSvc: licSvc, themeSvc: themeSvc, secret: secret, ttlMin: ttlMin}
}

type activateRequest struct {
	LicenseKey string `json:"license_key"`
	Domain     string `json:"domain"`
	ThemeSlug  string `json:"theme_slug"`
	Timestamp  int64  `json:"timestamp"`
	Signature  string `json:"signature"`
}

// Activate handles POST /api/v1/activate
// Validates license + theme access, then issues a single-use delivery token.
func (h *ActivateHandler) Activate(w http.ResponseWriter, r *http.Request) {
	var req activateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid_request", "malformed JSON body")
		return
	}

	sigData := fmt.Sprintf("%s:%s:%s:%d", req.LicenseKey, req.Domain, req.ThemeSlug, req.Timestamp)
	if !zcrypto.VerifyHMAC(h.secret, sigData, req.Signature) {
		writeError(w, "invalid_signature", "HMAC signature verification failed")
		return
	}

	now := time.Now().Unix()
	ttl := int64(h.ttlMin * 60)
	if now-req.Timestamp > ttl || req.Timestamp > now+60 {
		writeError(w, "token_expired", "request timestamp is too old or in the future")
		return
	}

	if req.ThemeSlug == "" {
		writeError(w, "invalid_request", "theme_slug is required")
		return
	}

	lic, err := h.licSvc.GetByKey(req.LicenseKey)
	if err != nil {
		writeError(w, "license_not_found", "license key does not exist")
		return
	}

	switch lic.Status {
	case license.StatusBanned:
		writeError(w, "license_banned", "")
		return
	case license.StatusInactive:
		writeError(w, "license_inactive", "")
		return
	}
	if !lic.IsLifetime && lic.ExpiresAt != nil && time.Now().After(*lic.ExpiresAt) {
		writeError(w, "license_expired", "")
		return
	}

	dt, err := h.themeSvc.IssueDeliveryToken(lic, req.ThemeSlug)
	if err != nil {
		writeError(w, "theme_not_available", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"delivery_token": dt.Token,
		"checksum":       dt.Theme.Checksum,
		"theme_slug":     dt.Theme.Slug,
		"theme_name":     dt.Theme.Name,
		"theme_version":  dt.Theme.Version,
		"expires_at":     dt.ExpiresAt.Format(time.RFC3339),
	})
}
