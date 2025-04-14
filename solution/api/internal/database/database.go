package database

import (
	"log"

	"git.mi6e4ka.dev/prod-2025/internal/models"
	"gorm.io/gorm"
)

func Init(conn gorm.Dialector) (*gorm.DB, error) {
	db, err := gorm.Open(conn, &gorm.Config{TranslateError: true})
	if err != nil {
		log.Println("failed to connect to database")
		return nil, err
	}

	if err := db.AutoMigrate(&models.Client{}, &models.Advertiser{}, &models.Campaign{}, &models.Event{}, &models.MLScore{}, &models.Setting{}); err != nil {
		log.Println("failed to migrate")
		return nil, err
	}

	log.Println("migrated successfully")

	return db, err
}
