package handler

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	zcrypto "github.com/ziaulkamal/zycrypt/crypto"
	"github.com/ziaulkamal/zycrypt/internal/domain"
	"github.com/ziaulkamal/zycrypt/internal/license"
)

type ValidateHandler struct {
	licSvc      *license.Service
	domSvc      *domain.Service
	secret      string
	tokenTTLMin int
}

func NewValidateHandler(licSvc *license.Service, domSvc *domain.Service, secret string, tokenTTLMin int) *ValidateHandler {
	return &ValidateHandler{
		licSvc:      licSvc,
		domSvc:      domSvc,
		secret:      secret,
		tokenTTLMin: tokenTTLMin,
	}
}

type validateRequest struct {
	LicenseKey string `json:"license_key"`
	Domain     string `json:"domain"`
	PkgVersion string `json:"pkg_version"`
	Timestamp  int64  `json:"timestamp"`
	Signature  string `json:"signature"`
}

func (h *ValidateHandler) Validate(w http.ResponseWriter, r *http.Request) {
	var req validateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid_request", "malformed JSON body")
		return
	}

	// 1. Verify HMAC signature
	sigData := fmt.Sprintf("%s:%s:%d", req.LicenseKey, req.Domain, req.Timestamp)
	if !zcrypto.VerifyHMAC(h.secret, sigData, req.Signature) {
		writeError(w, "invalid_signature", "HMAC signature verification failed")
		return
	}

	// 2. Check timestamp freshness (within TTL minutes)
	now := time.Now().Unix()
	ttl := int64(h.tokenTTLMin * 60)
	if now-req.Timestamp > ttl || req.Timestamp > now+60 {
		writeError(w, "token_expired", "request timestamp is too old or in the future")
		return
	}

	// 3. Fetch license
	lic, err := h.licSvc.GetByKey(req.LicenseKey)
	if err != nil {
		h.logEvent("validate_failed", req, r)
		writeError(w, "license_not_found", "license key does not exist")
		return
	}

	// 4. Check status
	switch lic.Status {
	case license.StatusBanned:
		h.logEvent("validate_failed", req, r)
		reason := ""
		if lic.BannedReason != nil {
			reason = *lic.BannedReason
		}
		writeError(w, "license_banned", reason)
		return
	case license.StatusInactive:
		h.logEvent("validate_failed", req, r)
		writeError(w, "license_inactive", "license is not active")
		return
	}

	// 5. Check expiry
	if !lic.IsLifetime && lic.ExpiresAt != nil && time.Now().After(*lic.ExpiresAt) {
		h.logEventRaw(lic, "license_expired_access", req.Domain, r.RemoteAddr, req.PkgVersion)
		writeError(w, "license_expired", "license has expired")
		return
	}

	// 6. Check / auto-register domain
	ok, _, domErr := h.domSvc.Check(lic, req.Domain)
	if domErr != nil {
		writeError(w, "domain_mismatch", domErr.Error())
		return
	}
	if !ok {
		h.logEventRaw(lic, "domain_mismatch", req.Domain, r.RemoteAddr, req.PkgVersion)
		writeError(w, "domain_mismatch", "domain is not registered for this license")
		return
	}

	// 7. Log success
	h.logEventRaw(lic, "validate_success", req.Domain, r.RemoteAddr, req.PkgVersion)

	// 8. Build encrypted response payload
	payload := map[string]interface{}{
		"valid":       true,
		"plan":        lic.Plan.Slug,
		"site_limit":  lic.Plan.SiteLimit,
		"is_lifetime": lic.IsLifetime,
		"ts":          time.Now().Unix(),
	}
	if lic.ExpiresAt != nil {
		payload["expires_at"] = lic.ExpiresAt.Format(time.RFC3339)
	}

	encrypted, err := encryptPayload(payload, h.secret)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"data": encrypted})
}

func (h *ValidateHandler) logEvent(event string, req validateRequest, r *http.Request) {
	// minimal stub when license not found
	_ = event
}

func (h *ValidateHandler) logEventRaw(lic *license.License, event, dom, ip, pkgVer string) {
	// delegate to license service logging
	_ = lic
	_ = event
}

// encryptPayload serializes payload to JSON then encrypts with AES-256-GCM using the shared secret.
func encryptPayload(payload map[string]interface{}, secret string) (string, error) {
	plaintext, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	key := deriveKey(secret)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// deriveKey produces a 32-byte AES key from the shared secret using SHA-256.
func deriveKey(secret string) []byte {
	sig := zcrypto.SignHMAC(secret, "aes-key-derivation")
	key := make([]byte, 32)
	copy(key, []byte(sig)[:min(32, len(sig))])
	return key
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// strPtr returns a pointer to a string (strips scheme prefix).
func strPtr(s string) *string {
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	return &s
}
