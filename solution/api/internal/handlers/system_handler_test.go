package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.mi6e4ka.dev/prod-2025/internal/dto"
	"git.mi6e4ka.dev/prod-2025/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSetMLScore(t *testing.T) {
	setupTestDB()
	handler := NewSystemHandler(db)

	client := models.Client{ClientID: uuid.New(), Login: "test_login", Age: 30, Location: "Test City", Gender: "MALE"}
	db.Create(&client)
	advertiser := models.Advertiser{AdvertiserID: uuid.New(), Name: "Test Advertiser"}
	db.Create(&advertiser)

	mlScore := models.MLScore{
		ClientID:     client.ClientID,
		AdvertiserID: advertiser.AdvertiserID,
	}
	*mlScore.Score = 100
	body, _ := json.Marshal(mlScore)

	router := setupRouter()
	router.POST("/ml-scores", handler.SetMLScore)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ml-scores", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var savedMLScore models.MLScore
	db.First(&savedMLScore, "client_id = ? AND advertiser_id = ?", client.ClientID, advertiser.AdvertiserID)
	assert.Equal(t, mlScore.Score, savedMLScore.Score)
}

func TestSetMLScoreInvalidClient(t *testing.T) {
	setupTestDB()
	handler := NewSystemHandler(db)

	mlScore := models.MLScore{
		ClientID:     uuid.New(),
		AdvertiserID: uuid.New(),
	}
	*mlScore.Score = 100
	body, _ := json.Marshal(mlScore)

	router := setupRouter()
	router.POST("/ml-scores", handler.SetMLScore)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ml-scores", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAdvanceTime(t *testing.T) {
	setupTestDB()
	handler := NewSystemHandler(db)
	router := setupRouter()
	router.POST("/time/advance", handler.AdvanceTime)
	currentDate := uint(1)
	body, _ := json.Marshal(dto.TimeSetBody{CurrentDate: &currentDate})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/time/advance", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedSettings models.Setting
	db.First(&updatedSettings, 1)
	assert.Equal(t, uint(1), updatedSettings.Day)
}
