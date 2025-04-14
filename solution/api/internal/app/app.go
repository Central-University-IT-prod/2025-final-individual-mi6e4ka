package app

import (
	"fmt"
	"log"

	"git.mi6e4ka.dev/prod-2025/internal/config"
	"git.mi6e4ka.dev/prod-2025/internal/database"
	"git.mi6e4ka.dev/prod-2025/internal/models"
	"git.mi6e4ka.dev/prod-2025/internal/router"
	"git.mi6e4ka.dev/prod-2025/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	router   *gin.Engine
	db       *gorm.DB
	settings *models.Setting
	config   *config.Config
}

func New() *App {
	app := &App{}
	app.config = config.LoadConfig()

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		app.config.DB.Host, app.config.DB.User, app.config.DB.Password, app.config.DB.DBName, app.config.DB.Port)
	db, err := database.Init(postgres.Open(dsn))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	s3, err := services.NewS3Client(app.config)
	if err != nil {
		log.Fatalf("failed to create s3 client: %v", err)
	}
	db.Where(models.Setting{ID: 1}).Attrs(models.Setting{Day: 0, Moderation: false}).FirstOrCreate(&app.settings)
	app.db = db

	router := router.InitRoutes(db, s3, app.settings, app.config)
	app.router = router

	return app
}

func (a *App) Run() {
	a.router.Run(fmt.Sprintf(":%d", a.config.HTTP.Port))
	log.Printf("running on port %d with current day %d\n", a.config.HTTP.Port, a.settings.Day)
}
