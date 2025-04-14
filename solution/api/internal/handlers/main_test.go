package handlers

import (
	"log"
	"os"
	"testing"

	"git.mi6e4ka.dev/prod-2025/internal/models"
	"git.mi6e4ka.dev/prod-2025/tests"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var settings = models.Setting{ID: 1, Day: 0}

func TestMain(m *testing.M) {
	pg := tests.NewPostgres()
	defer pg.Cleanup()
	var err error
	db, err = gorm.Open(postgres.Open(pg.MustConnStr()), &gorm.Config{TranslateError: true})
	db.AutoMigrate(&models.Client{}, &models.Advertiser{}, &models.Campaign{}, &models.Event{}, &models.MLScore{}, &models.Setting{})
	if err != nil {
		panic("failed to connect to database")
	}
	log.Println("connected to test database")
	code := m.Run()
	os.Exit(code)
}

func setupTestDB() {
	db.Exec("TRUNCATE TABLE clients, advertisers, campaigns, events, ml_scores, settings RESTART IDENTITY CASCADE")
}

func setupRouter() *gin.Engine {
	db.Save(&settings)

	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("settings", &settings)
		c.Next()
	})
	return router
}
