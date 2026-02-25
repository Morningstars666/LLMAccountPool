package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                  uint           `gorm:"primarykey" json:"id"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
	Username            string         `gorm:"unique;not null" json:"username"`
	Password            string         `gorm:"not null" json:"-"`
	FailedLoginAttempts int            `gorm:"default:0" json:"-"`
	LockedUntil         *time.Time     `json:"-"`
	LastLoginAt         *time.Time     `json:"-"`
}

type ExternalModel struct {
	ID        uint            `gorm:"primarykey" json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt gorm.DeletedAt  `gorm:"index" json:"-"`
	Name      string          `gorm:"not null" json:"name"`
	Model     string          `gorm:"not null" json:"model"`
	Strategy  string          `gorm:"default:'round_robin'" json:"strategy"`
	Sources   []RequestSource `gorm:"foreignKey:ExternalModelID" json:"sources,omitempty"`
}

type RequestSource struct {
	ID                 uint           `gorm:"primarykey" json:"id"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
	ExternalModelID    uint           `gorm:"not null" json:"external_model_id"`
	Name               string         `gorm:"not null" json:"name"`
	APIURL             string         `gorm:"not null" json:"api_url"`
	APIKey             string         `gorm:"not null" json:"api_key"`
	ModelName          string         `gorm:"not null" json:"model_name"`
	BillingMode        string         `gorm:"default:'count'" json:"billing_mode"`
	LimitCount         int64          `gorm:"default:0" json:"limit_count"`
	LimitTokens        int64          `gorm:"default:0" json:"limit_tokens"`
	LimitResetInterval int64          `gorm:"default:0" json:"limit_reset_interval"`
	LastResetAt        time.Time      `json:"last_reset_at"`
	UsedCount          int64          `gorm:"default:0" json:"used_count"`
	UsedTokens         int64          `gorm:"default:0" json:"used_tokens"`
	IsActive           bool           `gorm:"default:true" json:"is_active"`
}

type APIKey struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	Key             string         `gorm:"unique;not null" json:"key"`
	Note            string         `json:"note"`
	ExternalModelID uint           `gorm:"default:0" json:"external_model_id"`
	UsedCount       int64          `gorm:"default:0" json:"used_count"`
	UsedTokens      int64          `gorm:"default:0" json:"used_tokens"`
	InputTokens     int64          `gorm:"default:0" json:"input_tokens"`
	OutputTokens    int64          `gorm:"default:0" json:"output_tokens"`
}

type UsageRecord struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	APIKeyID        uint      `gorm:"not null" json:"api_key_id"`
	ExternalModelID uint      `gorm:"not null" json:"external_model_id"`
	SourceID        uint      `gorm:"not null" json:"source_id"`
	Model           string    `gorm:"not null" json:"model"`
	InputTokens     int64     `gorm:"default:0" json:"input_tokens"`
	OutputTokens    int64     `gorm:"default:0" json:"output_tokens"`
	Success         bool      `gorm:"default:true" json:"success"`
}
