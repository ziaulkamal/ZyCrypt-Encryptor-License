package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// SignHMAC returns HMAC-SHA256 hex of data using secret.
func SignHMAC(secret, data string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyHMAC checks whether the provided signature matches.
func VerifyHMAC(secret, data, signature string) bool {
	expected := SignHMAC(secret, data)
	return hmac.Equal([]byte(expected), []byte(signature))
}
