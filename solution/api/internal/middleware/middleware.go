package middleware

import (
	"git.mi6e4ka.dev/prod-2025/internal/models"
	"github.com/gin-gonic/gin"
)


func SettingsMiddleware(settings *models.Setting) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("settings", settings)
		c.Next()
	}
}