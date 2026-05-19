package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

const keyChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// GenerateKey produces a license key in format ZYC-{TYPE}{DUR}{CHK2}-XXXX-XXXX-XXXX
// planType: S=Single, M=Multi, L=Lifetime
// years: 1=1yr, 3=3yr, 0=lifetime
func GenerateKey(planType string, years int, secret string) (string, error) {
	seg1, err := randomSegment(4)
	if err != nil {
		return "", err
	}
	seg2, err := randomSegment(4)
	if err != nil {
		return "", err
	}
	seg3, err := randomSegment(4)
	if err != nil {
		return "", err
	}

	typeChar := strings.ToUpper(planType)
	if typeChar != "S" && typeChar != "M" && typeChar != "L" {
		typeChar = "S"
	}

	durChar := "1"
	switch years {
	case 3:
		durChar = "3"
	case 0:
		durChar = "0"
	}

	body := fmt.Sprintf("%s-%s-%s", seg1, seg2, seg3)
	chk := keyChecksum(typeChar+durChar+body, secret)

	key := fmt.Sprintf("ZYC-%s%s%s-%s-%s-%s", typeChar, durChar, chk, seg1, seg2, seg3)
	return key, nil
}

// VerifyKey checks format, characters, and HMAC checksum of a key.
func VerifyKey(key, secret string) error {
	parts := strings.Split(key, "-")
	if len(parts) != 5 {
		return fmt.Errorf("invalid key format: expected 5 segments separated by '-'")
	}
	if parts[0] != "ZYC" {
		return fmt.Errorf("invalid key prefix: expected ZYC")
	}

	prefix := parts[1] // e.g. S1AB
	if len(prefix) != 4 {
		return fmt.Errorf("invalid key prefix segment: expected 4 chars")
	}

	typeChar := string(prefix[0])
	durChar := string(prefix[1])
	checksum := prefix[2:]

	if typeChar != "S" && typeChar != "M" && typeChar != "L" {
		return fmt.Errorf("invalid plan type char: %s", typeChar)
	}
	if durChar != "0" && durChar != "1" && durChar != "3" {
		return fmt.Errorf("invalid duration char: %s", durChar)
	}

	body := fmt.Sprintf("%s-%s-%s", parts[2], parts[3], parts[4])
	expected := keyChecksum(typeChar+durChar+body, secret)
	if checksum != expected {
		return fmt.Errorf("checksum mismatch")
	}

	return nil
}

// PlanTypeFromKey returns the plan type char from a key.
func PlanTypeFromKey(key string) string {
	parts := strings.Split(key, "-")
	if len(parts) < 2 || len(parts[1]) < 1 {
		return ""
	}
	return string(parts[1][0])
}

// DurationFromKey returns the duration years from a key (0=lifetime, 1=1yr, 3=3yr).
func DurationFromKey(key string) int {
	parts := strings.Split(key, "-")
	if len(parts) < 2 || len(parts[1]) < 2 {
		return 1
	}
	switch string(parts[1][1]) {
	case "3":
		return 3
	case "0":
		return 0
	}
	return 1
}

// keyChecksum returns a 2-char HMAC-derived checksum.
func keyChecksum(data, secret string) string {
	sig := SignHMAC(secret, data)
	raw, _ := hex.DecodeString(sig)
	if len(raw) < 2 {
		return "AA"
	}
	a := keyChars[raw[0]%byte(len(keyChars))]
	b := keyChars[raw[1]%byte(len(keyChars))]
	return string([]byte{a, b})
}

func randomSegment(length int) (string, error) {
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(keyChars))))
		if err != nil {
			return "", err
		}
		result[i] = keyChars[n.Int64()]
	}
	return string(result), nil
}
