package database

import (
	"fmt"

	"github.com/Aks98068/forensics/internal/models"
	"gorm.io/gorm"
)

func MigrateDB(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.PasswordResetToken{},
		&models.EmailVerificationToken{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate the database schema : %w", err)
	}
	return nil
}
