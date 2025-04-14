package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.mi6e4ka.dev/prod-2025/internal/dto"
	"git.mi6e4ka.dev/prod-2025/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var campaignID = uuid.New()
var clientID1 = uuid.New()
var clientID2 = uuid.New()
var clientID3 = uuid.New()
var clientID4 = uuid.New()
var advertiser = uuid.New()

func testDataLoad() {
	setupTestDB()
	db.Save(&models.Advertiser{AdvertiserID: advertiser, Name: "Test Advertiser"})
	campaign := &models.Campaign{CampaignID: campaignID, AdTitle: "Test Campaign", AdvertiserID: advertiser, StartDate: new(uint), EndDate: new(uint)}
	*campaign.StartDate = 0
	*campaign.EndDate = 2
	db.Save(campaign)
	db.Save([]models.Client{
		{ClientID: clientID1, Login: "test_login1", Age: 20, Location: "test-town"},
		{ClientID: clientID2, Login: "test_login2", Age: 30, Location: "test-town"},
		{ClientID: clientID3, Login: "test_login3", Age: 40, Location: "test-town"},
		{ClientID: clientID4, Login: "test_login4", Age: 50, Location: "test-town"},
	})
	stats := []models.Event{
		{CampaignID: campaignID, Type: "click", ClientID: clientID1, Cost: 1, Day: 0},
		{CampaignID: campaignID, Type: "click", ClientID: clientID2, Cost: 2, Day: 1},
		{CampaignID: campaignID, Type: "click", ClientID: clientID3, Cost: 10, Day: 1},
		{CampaignID: campaignID, Type: "click", ClientID: clientID4, Cost: 10, Day: 2},
		{CampaignID: campaignID, Type: "view", ClientID: clientID1, Cost: 14, Day: 0},
		{CampaignID: campaignID, Type: "view", ClientID: clientID2, Cost: 88, Day: 2},
	}
	db.Save(stats)
}

func TestGetCampaignStats(t *testing.T) {
	testDataLoad()
	router := setupRouter()
	handler := NewStatsHandler(db)

	router.GET("/campaigns/:id/stats", handler.GetCampaignStats)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/campaigns/"+campaignID.String()+"/stats", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responseStats dto.Stats
	json.Unmarshal(w.Body.Bytes(), &responseStats)

	assert.Equal(t, 4, responseStats.ClicksCount)
	assert.Equal(t, 2, responseStats.ImpressionsCount)
	assert.Equal(t, float64(125), responseStats.SpentTotal)
	assert.Equal(t, float64(102), responseStats.SpentImpressions)
	assert.Equal(t, float64(23), responseStats.SpentClicks)
	assert.Equal(t, float64(50), responseStats.Conversion)
}

func TestGetCampaignStatNotFount(t *testing.T) {
	testDataLoad()
	router := setupRouter()
	handler := NewStatsHandler(db)

	router.GET("/campaigns/:id/stats", handler.GetCampaignStats)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/campaigns/"+uuid.New().String()+"/stats", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetCampaignsCampaignDailyStats(t *testing.T) {
	testDataLoad()
	router := setupRouter()
	handler := NewStatsHandler(db)

	router.GET("/campaigns/:id/stats/daily", handler.GetCampaignDailyStats)
	w := httptest.NewRecorder()
	settings.Day = 1
	req, _ := http.NewRequest("GET", "/campaigns/"+campaignID.String()+"/stats/daily", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responseStats []dto.DailyStats
	json.Unmarshal(w.Body.Bytes(), &responseStats)
	assert.Len(t, responseStats, 2)

	settings.Day = 3
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/campaigns/"+campaignID.String()+"/stats/daily", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &responseStats)
	assert.Len(t, responseStats, 3)
	testData := []dto.DailyStats{
		{Stats: dto.Stats{ImpressionsCount: 1, ClicksCount: 1, Conversion: 100, SpentImpressions: 14, SpentClicks: 1, SpentTotal: 15}, Date: 0},
		{Stats: dto.Stats{ImpressionsCount: 0, ClicksCount: 2, Conversion: 0, SpentImpressions: 0, SpentClicks: 12, SpentTotal: 12}, Date: 1},
		{Stats: dto.Stats{ImpressionsCount: 1, ClicksCount: 1, Conversion: 100, SpentImpressions: 88, SpentClicks: 10, SpentTotal: 98}, Date: 2},
	}

	assert.True(t, assert.ObjectsAreEqual(testData, responseStats))
	settings.Day = 0
}
