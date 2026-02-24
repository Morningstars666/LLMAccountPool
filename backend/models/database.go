package models

import (
	"llmaccountpool/config"
	"llmaccountpool/utils"
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(cfg *config.Config) {
	var err error
	DB, err = gorm.Open(sqlite.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	sqlDB.Exec("PRAGMA journal_mode=WAL")
	sqlDB.Exec("PRAGMA busy_timeout=5000")

	if err := DB.AutoMigrate(
		&User{},
		&ExternalModel{},
		&RequestSource{},
		&APIKey{},
		&UsageRecord{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	var count int64
	DB.Model(&User{}).Count(&count)
	if count == 0 {
		hash, _ := utils.HashPassword("admin123")
		defaultUser := User{
			Username: "admin",
			Password: hash,
		}
		DB.Create(&defaultUser)
		log.Println("Default admin user created: admin / admin123")
	}

	log.Println("Database initialized successfully")
}
