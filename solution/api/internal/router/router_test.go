package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"git.mi6e4ka.dev/prod-2025/internal/config"
	"git.mi6e4ka.dev/prod-2025/internal/models"
	"git.mi6e4ka.dev/prod-2025/internal/router"
	"git.mi6e4ka.dev/prod-2025/internal/services"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestRoutes(t *testing.T) {
	mockDB := &gorm.DB{} // Mock DB instance
	settings := &models.Setting{}
	s3 := &services.S3Client{}
	r := router.InitRoutes(mockDB, s3, settings, &config.Config{})

	tests := []struct {
		method   string
		endpoint string
	}{
		{"POST", "/ml-scores"},
		{"POST", "/time/advance"},
		{"GET", "/stats/campaigns/:id"},
		{"GET", "/stats/advertisers/:id/campaigns"},
		{"GET", "/stats/campaigns/:id/daily"},
		{"GET", "/stats/advertisers/:id/campaigns/daily"},
		{"GET", "/ads"},
		{"POST", "/ads/:id/click"},
		{"POST", "/advertisers/:id/campaigns"},
		{"GET", "/advertisers/:id/campaigns"},
		{"GET", "/advertisers/:id/campaigns/:campId"},
		{"PUT", "/advertisers/:id/campaigns/:campId"},
		{"DELETE", "/advertisers/:id/campaigns/:campId"},
		{"GET", "/advertisers/:id"},
		{"POST", "/advertisers/bulk"},
		{"GET", "/clients/:id"},
		{"POST", "/clients/bulk"},
		{"GET", "/"},
	}

	for _, test := range tests {
		req, _ := http.NewRequest(test.method, test.endpoint, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.NotEqual(t, http.StatusNotFound, w.Code, "Endpoint %s %s should exist", test.method, test.endpoint)
	}
}
