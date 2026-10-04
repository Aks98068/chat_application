package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleUser    UserRole = "USER"
	RoleAnalyst UserRole = "ANALYST"
	RoleAdmin   UserRole = "ADMIN"
)

type UserStatus string

const (
	StatusActive  UserStatus = "ACTIVE"
	StatusSuspend UserStatus = "SUSPENDED"
	StatusDisable UserStatus = "DISABLED"
)

type User struct {
	ID uuid.UUID `gorm:"type:char(36);primaryKey"`

	Email string `gorm:"type:varchar(255);uniqueIndex;not null"`

	Username string `gorm:"type:varchar(50);uniqueIndex;not null"`

	PasswordHash string `gorm:"type:varchar(255);not null"`

	FirstName string `gorm:"type:varchar(100);not null"`

	LastName string `gorm:"type:varchar(100);not null"`

	Role UserRole `gorm:"type:varchar(20);not null;default:'USER'"`

	Status UserStatus `gorm:"type:varchar(20);not null;default:'ACTIVE'"`

	EmailVerified bool `gorm:"not null;default:false"`

	FailedLoginAttempts int `gorm:"not null;default:0"`

	LockedUntil *time.Time `gorm:"null"`

	LastLoginAt *time.Time `gorm:"null"`

	CreatedAt time.Time

	UpdatedAt time.Time

	RefreshTokens []RefreshToken `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	return nil
}

type RefreshToken struct {
	ID uuid.UUID `gorm:"type:char(36);primaryKey"`

	UserID uuid.UUID `gorm:"type:char(36);not null;index"`

	TokenHash string `gorm:"type:char(64);uniqueIndex;not null"`

	TokenFamily uuid.UUID `gorm:"type:char(36);not null;index"`

	ExpiresAt time.Time `gorm:"not null;index"`

	RevokedAt *time.Time `gorm:"null"`

	ReplacedByTokenID *uuid.UUID `gorm:"type:char(36);null;index"`

	CreatedAt time.Time

	UserAgent string `gorm:"type:text"`

	IPAddress string `gorm:"type:varchar(45)"`

	User User `gorm:"foreignKey:UserID;references:ID"`
}

func (r *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}

	if r.TokenFamily == uuid.Nil {
		r.TokenFamily = uuid.New()
	}

	return nil
}

type PasswordResetToken struct {
	ID        uuid.UUID  `gorm:"type:char(36);primaryKey"`
	UserID    uuid.UUID  `gorm:"type:char(36);not null;index"`
	TokenHash string     `gorm:"type:char(64);uniqueIndex;not null"`
	ExpiresAt time.Time  `gorm:"not null;index"`
	UsedAt    *time.Time `gorm:"null"`
	RevokedAt *time.Time `gorm:"null"`
	CreatedAt time.Time

	User User `gorm:"foreignKey:UserID;references:ID"`
}

func (p *PasswordResetToken) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	return nil
}

type EmailVerificationToken struct {
	ID        uuid.UUID  `gorm:"type:char(36);primaryKey"`
	UserID    uuid.UUID  `gorm:"type:char(36);not null;index"`
	TokenHash string     `gorm:"type:char(64);uniqueIndex;not null"`
	ExpiresAt time.Time  `gorm:"not null;index"`
	UsedAt    *time.Time `gorm:"null"`
	RevokedAt *time.Time `gorm:"null"`
	CreatedAt time.Time

	User User `gorm:"foreignKey:UserID;references:ID"`
}

func (e *EmailVerificationToken) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}

	return nil
}
