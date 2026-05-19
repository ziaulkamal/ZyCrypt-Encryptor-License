package license

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	zcrypto "github.com/ziaulkamal/zycrypt/crypto"
	"gorm.io/gorm"
)

type Service struct {
	db     *gorm.DB
	secret string
}

func NewService(db *gorm.DB, secret string) *Service {
	return &Service{db: db, secret: secret}
}

// ---- Plan operations ----

func (s *Service) CreatePlan(name, slug string, siteLimit int, duration string, price float64) (*Plan, error) {
	plan := &Plan{
		Name:      name,
		Slug:      slug,
		SiteLimit: siteLimit,
		Duration:  duration,
		Price:     price,
		IsActive:  true,
	}
	if err := s.db.Create(plan).Error; err != nil {
		return nil, fmt.Errorf("create plan: %w", err)
	}
	return plan, nil
}

func (s *Service) ListPlans() ([]Plan, error) {
	var plans []Plan
	return plans, s.db.Order("created_at ASC").Find(&plans).Error
}

func (s *Service) GetPlanBySlug(slug string) (*Plan, error) {
	var plan Plan
	if err := s.db.Where("slug = ?", slug).First(&plan).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

func (s *Service) UpdatePlan(slug string, updates map[string]interface{}) error {
	return s.db.Model(&Plan{}).Where("slug = ?", slug).Updates(updates).Error
}

func (s *Service) DisablePlan(slug string) error {
	return s.db.Model(&Plan{}).Where("slug = ?", slug).Update("is_active", false).Error
}

// ---- License operations ----

func (s *Service) Create(customerName, email, planSlug string) (*License, error) {
	plan, err := s.GetPlanBySlug(planSlug)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	typeChar := planTypeChar(plan)
	years := planYears(plan)

	key, err := zcrypto.GenerateKey(typeChar, years, s.secret)
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}

	lic := &License{
		LicenseKey:    key,
		PlanID:        plan.ID,
		CustomerName:  customerName,
		CustomerEmail: email,
		Status:        StatusActive,
		PurchaseDate:  time.Now(),
		IsLifetime:    plan.Duration == "lifetime",
	}

	if plan.Duration != "lifetime" {
		exp := calcExpiry(time.Now(), plan.Duration)
		lic.ExpiresAt = &exp
	}

	if err := s.db.Create(lic).Error; err != nil {
		return nil, err
	}

	s.log(lic.ID, "license_created", nil, nil, nil, map[string]interface{}{
		"plan": planSlug, "email": email,
	})

	return lic, nil
}

func (s *Service) List(status string) ([]License, error) {
	var licenses []License
	q := s.db.Preload("Plan").Order("created_at DESC")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	return licenses, q.Find(&licenses).Error
}

func (s *Service) GetByKey(key string) (*License, error) {
	var lic License
	err := s.db.Preload("Plan").Preload("Domains").
		Where("license_key = ?", key).First(&lic).Error
	return &lic, err
}

func (s *Service) Ban(key, reason string) error {
	lic, err := s.GetByKey(key)
	if err != nil {
		return err
	}
	if err := s.db.Model(lic).Updates(map[string]interface{}{
		"status":        StatusBanned,
		"banned_reason": reason,
		"updated_at":    time.Now(),
	}).Error; err != nil {
		return err
	}
	s.log(lic.ID, "license_banned", nil, nil, nil, map[string]interface{}{"reason": reason})
	return nil
}

func (s *Service) Unban(key string) error {
	lic, err := s.GetByKey(key)
	if err != nil {
		return err
	}
	if err := s.db.Model(lic).Updates(map[string]interface{}{
		"status":        StatusActive,
		"banned_reason": nil,
		"updated_at":    time.Now(),
	}).Error; err != nil {
		return err
	}
	s.log(lic.ID, "license_unbanned", nil, nil, nil, nil)
	return nil
}

func (s *Service) Extend(key string, years int) error {
	lic, err := s.GetByKey(key)
	if err != nil {
		return err
	}
	if lic.IsLifetime {
		return fmt.Errorf("cannot extend a lifetime license")
	}

	base := time.Now()
	if lic.ExpiresAt != nil && lic.ExpiresAt.After(base) {
		base = *lic.ExpiresAt
	}
	newExpiry := base.AddDate(years, 0, 0)

	if err := s.db.Model(lic).Updates(map[string]interface{}{
		"expires_at": newExpiry,
		"status":     StatusActive,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return err
	}
	s.log(lic.ID, "license_extended", nil, nil, nil, map[string]interface{}{
		"years": years, "new_expiry": newExpiry,
	})
	return nil
}

func (s *Service) Revoke(key string) error {
	lic, err := s.GetByKey(key)
	if err != nil {
		return err
	}
	s.log(lic.ID, "license_revoked", nil, nil, nil, map[string]interface{}{
		"customer": lic.CustomerEmail,
	})
	return s.db.Delete(lic).Error
}

func (s *Service) RecentLogs(licenseID uuid.UUID, limit int) ([]AuditLog, error) {
	var logs []AuditLog
	return logs, s.db.Where("license_id = ?", licenseID).
		Order("created_at DESC").Limit(limit).Find(&logs).Error
}

// ---- Internal helpers ----

func (s *Service) log(licenseID uuid.UUID, event string, domain, ip, pkgVer *string, payload map[string]interface{}) {
	entry := &AuditLog{
		LicenseID:  licenseID,
		Event:      event,
		Domain:     domain,
		IPAddress:  ip,
		PkgVersion: pkgVer,
	}
	if payload != nil {
		if b, err := json.Marshal(payload); err == nil {
			entry.Payload = b
		}
	}
	s.db.Create(entry)
}

func planTypeChar(plan *Plan) string {
	switch plan.Duration {
	case "lifetime":
		return "L"
	}
	if plan.SiteLimit > 1 || plan.SiteLimit == -1 {
		return "M"
	}
	return "S"
}

func planYears(plan *Plan) int {
	switch plan.Duration {
	case "3y":
		return 3
	case "lifetime":
		return 0
	}
	return 1
}

func calcExpiry(from time.Time, duration string) time.Time {
	switch duration {
	case "3y":
		return from.AddDate(3, 0, 0)
	}
	return from.AddDate(1, 0, 0)
}
