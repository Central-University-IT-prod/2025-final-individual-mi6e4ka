package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.mi6e4ka.dev/prod-2025/internal/dto"
	"git.mi6e4ka.dev/prod-2025/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetAds(t *testing.T) {
	setupTestDB()
	handler := NewAdHandler(db)
	client := models.Client{ClientID: uuid.New(), Login: "test_login", Age: 30, Location: "Test City", Gender: "MALE"}
	db.Create(&client)
	advertiser := models.Advertiser{AdvertiserID: uuid.New(), Name: "Test Advertiser"}
	db.Create(&advertiser)
	adCost := 1.0
	campaign := models.Campaign{
		CampaignID:        uuid.New(),
		AdvertiserID:      advertiser.AdvertiserID,
		ImpressionsLimit:  new(uint),
		ClicksLimit:       new(uint),
		CostPerImpression: &adCost,
		CostPerClick:      &adCost,
		AdTitle:           "Test Ad",
		AdText:            "Test Ad Text",
		StartDate:         new(uint),
		EndDate:           new(uint),
	}
	db.Create(&campaign)
	settings := models.Setting{ID: 1, Day: 0}
	db.Create(&settings)

	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("settings", &settings)
		c.Next()
	})
	router.GET("/ads", handler.GetAds)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ads?client_id="+client.ClientID.String(), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var ad dto.AdUser
	err := json.Unmarshal(w.Body.Bytes(), &ad)
	assert.NoError(t, err)
	assert.Equal(t, "Test Ad", ad.AdTitle)
	assert.Equal(t, campaign.CampaignID.String(), ad.AdID)
}

func TestGetAdsClientNotFound(t *testing.T) {
	setupTestDB()
	handler := NewAdHandler(db)

	settings := models.Setting{ID: 1, Day: 0}
	db.Create(&settings)

	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("settings", &settings)
		c.Next()
	})
	router.GET("/ads", handler.GetAds)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ads?client_id=00000000-0000-0000-0000-000000000000", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestClickAd(t *testing.T) {
	setupTestDB()
	handler := NewAdHandler(db)
	client := models.Client{ClientID: uuid.New(), Login: "sigma", Age: 52, Location: "Moscow City", Gender: "MALE"}
	db.Create(&client)
	advertiser := models.Advertiser{AdvertiserID: uuid.New(), Name: "Z-Bank"}
	db.Create(&advertiser)
	adCost := 1.0
	campaign := models.Campaign{
		CampaignID:        uuid.New(),
		AdvertiserID:      advertiser.AdvertiserID,
		ImpressionsLimit:  new(uint),
		ClicksLimit:       new(uint),
		CostPerImpression: &adCost,
		CostPerClick:      &adCost,
		AdTitle:           "Test Ad",
		AdText:            "Test Ad Text",
		StartDate:         new(uint),
		EndDate:           new(uint),
	}
	db.Create(&campaign)
	settings := models.Setting{ID: 1, Day: 0}
	db.Create(&settings)
	db.Create(&models.Event{
		ClientID:   client.ClientID,
		CampaignID: campaign.CampaignID,
		Type:       "view",
		Cost:       *campaign.CostPerImpression,
		Day:        settings.Day,
	})

	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("settings", &settings)
		c.Next()
	})
	router.POST("/ads/:id/click", handler.ClickAd)

	body, _ := json.Marshal(dto.QueryClient{ClientID: client.ClientID.String()})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ads/"+campaign.CampaignID.String()+"/click", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	// Check that the event was created
	var event models.Event
	db.Where("client_id = ? AND campaign_id = ? AND type = 'click'", client.ClientID, campaign.CampaignID).First(&event)
	assert.Equal(t, "click", event.Type)
	assert.Equal(t, *campaign.CostPerClick, event.Cost)
}

func TestClickAdNotFound(t *testing.T) {
	setupTestDB()
	handler := NewAdHandler(db)

	settings := models.Setting{ID: 1, Day: 0}
	db.Create(&settings)

	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("settings", &settings)
		c.Next()
	})
	router.POST("/ads/:id/click", handler.ClickAd)

	body, _ := json.Marshal(dto.QueryClient{ClientID: "00000000-0000-0000-0000-000000000000"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ads/00000000-0000-0000-0000-000000000000/click", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
