package license

import (
	"time"

	"github.com/google/uuid"
)

type Plan struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Slug      string    `gorm:"size:50;uniqueIndex;not null" json:"slug"`
	SiteLimit int       `gorm:"not null;default:1" json:"site_limit"`
	Duration  string    `gorm:"size:20;not null;default:'1y'" json:"duration"`
	Price     float64   `gorm:"type:decimal(12,2);not null;default:0" json:"price"`
	IsActive  bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

func (Plan) TableName() string { return "license_plans" }

type License struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	LicenseKey    string     `gorm:"size:40;uniqueIndex;not null" json:"license_key"`
	PlanID        uuid.UUID  `gorm:"type:uuid;not null" json:"plan_id"`
	Plan          Plan       `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
	CustomerName  string     `gorm:"size:150;not null" json:"customer_name"`
	CustomerEmail string     `gorm:"size:150;not null" json:"customer_email"`
	Status        string     `gorm:"size:20;not null;default:'active'" json:"status"`
	PurchaseDate  time.Time  `gorm:"type:date;not null" json:"purchase_date"`
	ExpiresAt     *time.Time `json:"expires_at"`
	IsLifetime    bool       `gorm:"not null;default:false" json:"is_lifetime"`
	BannedReason  *string    `json:"banned_reason,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Domains       []Domain   `gorm:"foreignKey:LicenseID" json:"domains,omitempty"`
}

func (License) TableName() string { return "licenses" }

type Domain struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	LicenseID    uuid.UUID `gorm:"type:uuid;not null" json:"license_id"`
	Domain       string    `gorm:"size:255;not null" json:"domain"`
	IsPrimary    bool      `gorm:"not null;default:false" json:"is_primary"`
	RegisteredAt time.Time `gorm:"not null" json:"registered_at"`
}

func (Domain) TableName() string { return "license_domains" }

type AuditLog struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	LicenseID  uuid.UUID  `gorm:"type:uuid;not null" json:"license_id"`
	Event      string     `gorm:"size:50;not null" json:"event"`
	Domain     *string    `gorm:"size:255" json:"domain,omitempty"`
	IPAddress  *string    `gorm:"size:45" json:"ip_address,omitempty"`
	PkgVersion *string    `gorm:"size:20" json:"pkg_version,omitempty"`
	Payload    []byte     `gorm:"type:jsonb" json:"payload,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (AuditLog) TableName() string { return "license_audit_logs" }

const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusBanned   = "banned"
	StatusExpired  = "expired"
)

type Theme struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Slug      string    `gorm:"size:100;uniqueIndex;not null"                  json:"slug"`
	Name      string    `gorm:"size:255;not null"                              json:"name"`
	Version   string    `gorm:"size:20;not null;default:'1.0.0'"               json:"version"`
	Bundle    []byte    `gorm:"type:bytea;not null"                            json:"-"`
	Checksum  string    `gorm:"size:64;not null"                               json:"checksum"`
	FileSize  int64     `gorm:"not null;default:0"                             json:"file_size"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Theme) TableName() string { return "themes" }

type PlanTheme struct {
	PlanID  uuid.UUID `gorm:"type:uuid;primaryKey" json:"plan_id"`
	ThemeID uuid.UUID `gorm:"type:uuid;primaryKey" json:"theme_id"`
	Theme   Theme     `gorm:"foreignKey:ThemeID"  json:"theme,omitempty"`
}

func (PlanTheme) TableName() string { return "plan_themes" }

type DeliveryToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Token     string    `gorm:"size:64;uniqueIndex;not null"                   json:"token"`
	LicenseID uuid.UUID `gorm:"type:uuid;not null"                             json:"license_id"`
	ThemeID   uuid.UUID `gorm:"type:uuid;not null"                             json:"theme_id"`
	Theme     Theme     `gorm:"foreignKey:ThemeID"                             json:"theme,omitempty"`
	ExpiresAt time.Time `gorm:"not null"                                       json:"expires_at"`
	Used      bool      `gorm:"not null;default:false"                         json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

func (DeliveryToken) TableName() string { return "delivery_tokens" }
