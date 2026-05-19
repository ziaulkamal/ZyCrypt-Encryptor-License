package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ziaulkamal/zycrypt/internal/license"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Add(licenseID uuid.UUID, domain string, siteLimit int) error {
	var count int64
	s.db.Model(&license.Domain{}).Where("license_id = ?", licenseID).Count(&count)

	if siteLimit != -1 && int(count) >= siteLimit {
		return fmt.Errorf("domain limit reached (%d/%d)", count, siteLimit)
	}

	isPrimary := count == 0
	d := &license.Domain{
		LicenseID:    licenseID,
		Domain:       domain,
		IsPrimary:    isPrimary,
		RegisteredAt: time.Now(),
	}
	if err := s.db.Create(d).Error; err != nil {
		return fmt.Errorf("add domain: %w", err)
	}
	return nil
}

func (s *Service) Remove(licenseID uuid.UUID, domain string) error {
	result := s.db.Where("license_id = ? AND domain = ?", licenseID, domain).
		Delete(&license.Domain{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("domain %q not found for this license", domain)
	}
	return nil
}

func (s *Service) List(licenseID uuid.UUID) ([]license.Domain, error) {
	var domains []license.Domain
	return domains, s.db.Where("license_id = ?", licenseID).
		Order("registered_at ASC").Find(&domains).Error
}

// Check verifies whether the given domain is registered for the license.
// Returns (registered, isNew, error). isNew=true means domain was just auto-registered.
func (s *Service) Check(lic *license.License, domain string) (ok bool, isNew bool, err error) {
	var existing license.Domain
	result := s.db.Where("license_id = ? AND domain = ?", lic.ID, domain).First(&existing)
	if result.Error == nil {
		return true, false, nil
	}

	// Domain not registered — check if we can auto-register
	var count int64
	s.db.Model(&license.Domain{}).Where("license_id = ?", lic.ID).Count(&count)

	siteLimit := lic.Plan.SiteLimit
	if siteLimit != -1 && int(count) >= siteLimit {
		return false, false, nil
	}

	// Auto-register
	if err := s.Add(lic.ID, domain, siteLimit); err != nil {
		return false, false, err
	}
	return true, true, nil
}
