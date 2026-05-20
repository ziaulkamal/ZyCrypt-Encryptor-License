package theme

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	zcrypto "github.com/ziaulkamal/zycrypt/crypto"
	"github.com/ziaulkamal/zycrypt/internal/license"
	"gorm.io/gorm"
)

type Service struct {
	db     *gorm.DB
	secret string
}

func NewService(db *gorm.DB, secret string) *Service {
	return &Service{db: db, secret: secret}
}

func (s *Service) Upload(slug, name, version string, bundle []byte, checksum string) (*license.Theme, error) {
	t := &license.Theme{
		Slug:     slug,
		Name:     name,
		Version:  version,
		Bundle:   bundle,
		Checksum: checksum,
		FileSize: int64(len(bundle)),
	}
	if err := s.db.Create(t).Error; err != nil {
		return nil, fmt.Errorf("upload theme: %w", err)
	}
	return t, nil
}

func (s *Service) List() ([]license.Theme, error) {
	var themes []license.Theme
	return themes, s.db.Order("slug ASC").Find(&themes).Error
}

func (s *Service) GetBySlug(slug string) (*license.Theme, error) {
	var t license.Theme
	if err := s.db.Where("slug = ?", slug).First(&t).Error; err != nil {
		return nil, fmt.Errorf("theme %q not found", slug)
	}
	return &t, nil
}

func (s *Service) Delete(slug string) error {
	return s.db.Where("slug = ?", slug).Delete(&license.Theme{}).Error
}

func (s *Service) AssignToPlan(themeSlug, planSlug string) error {
	var plan license.Plan
	if err := s.db.Where("slug = ?", planSlug).First(&plan).Error; err != nil {
		return fmt.Errorf("plan %q not found", planSlug)
	}
	theme, err := s.GetBySlug(themeSlug)
	if err != nil {
		return err
	}
	pt := license.PlanTheme{PlanID: plan.ID, ThemeID: theme.ID}
	return s.db.FirstOrCreate(&pt, pt).Error
}

func (s *Service) UnassignFromPlan(themeSlug, planSlug string) error {
	var plan license.Plan
	if err := s.db.Where("slug = ?", planSlug).First(&plan).Error; err != nil {
		return fmt.Errorf("plan %q not found", planSlug)
	}
	theme, err := s.GetBySlug(themeSlug)
	if err != nil {
		return err
	}
	return s.db.Where("plan_id = ? AND theme_id = ?", plan.ID, theme.ID).
		Delete(&license.PlanTheme{}).Error
}

func (s *Service) ThemesForPlan(planID uuid.UUID) ([]license.Theme, error) {
	var rows []license.PlanTheme
	if err := s.db.Preload("Theme").Where("plan_id = ?", planID).Find(&rows).Error; err != nil {
		return nil, err
	}
	themes := make([]license.Theme, 0, len(rows))
	for _, r := range rows {
		themes = append(themes, r.Theme)
	}
	return themes, nil
}

func (s *Service) IssueDeliveryToken(lic *license.License, themeSlug string) (*license.DeliveryToken, error) {
	theme, err := s.GetBySlug(themeSlug)
	if err != nil {
		return nil, err
	}

	var count int64
	s.db.Model(&license.PlanTheme{}).
		Where("plan_id = ? AND theme_id = ?", lic.PlanID, theme.ID).
		Count(&count)
	if count == 0 {
		return nil, fmt.Errorf("theme %q is not available for this license plan", themeSlug)
	}

	rawToken := make([]byte, 32)
	if _, err := rand.Read(rawToken); err != nil {
		return nil, err
	}

	dt := &license.DeliveryToken{
		Token:     hex.EncodeToString(rawToken),
		LicenseID: lic.ID,
		ThemeID:   theme.ID,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	if err := s.db.Create(dt).Error; err != nil {
		return nil, fmt.Errorf("issue delivery token: %w", err)
	}

	dt.Theme = *theme
	return dt, nil
}

// RedeemToken marks the token as used and returns the AES-256-CBC encrypted bundle.
// Encryption matches PHP BundleInstaller::decryptBundle:
//   key = hmac-sha256(shared_secret, "bundle-key")[:32]
//   wire format = iv (16 bytes) || ciphertext
func (s *Service) RedeemToken(token string) ([]byte, *license.Theme, error) {
	var dt license.DeliveryToken
	if err := s.db.Preload("Theme").Where("token = ?", token).First(&dt).Error; err != nil {
		return nil, nil, fmt.Errorf("token not found")
	}
	if dt.Used {
		return nil, nil, fmt.Errorf("token already used")
	}
	if time.Now().After(dt.ExpiresAt) {
		return nil, nil, fmt.Errorf("token expired")
	}

	s.db.Model(&dt).Update("used", true)

	encrypted, err := s.encryptBundle(dt.Theme.Bundle)
	if err != nil {
		return nil, nil, fmt.Errorf("encrypt bundle: %w", err)
	}
	return encrypted, &dt.Theme, nil
}

func (s *Service) encryptBundle(raw []byte) ([]byte, error) {
	key := s.bundleKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	padded := pkcs7Pad(raw, aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)

	out := make([]byte, 0, len(iv)+len(ciphertext))
	out = append(out, iv...)
	out = append(out, ciphertext...)
	return out, nil
}

func (s *Service) bundleKey() []byte {
	sig := zcrypto.SignHMAC(s.secret, "bundle-key")
	key := make([]byte, 32)
	copy(key, []byte(sig))
	return key
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	out := make([]byte, len(data)+padding)
	copy(out, data)
	for i := len(data); i < len(out); i++ {
		out[i] = byte(padding)
	}
	return out
}
