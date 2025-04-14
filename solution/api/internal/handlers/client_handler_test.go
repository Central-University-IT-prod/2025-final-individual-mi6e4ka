package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.mi6e4ka.dev/prod-2025/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetClient(t *testing.T) {
	setupTestDB()
	handler := NewClientHandler(db)
	client := models.Client{ClientID: uuid.New(), Login: "login without validation", Age: 52, Location: "Moscow City", Gender: "MALE"}
	db.Create(&client)

	router := gin.Default()
	router.GET("/clients/:id", handler.GetClient)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/clients/"+client.ClientID.String(), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responseClient models.Client
	err := json.Unmarshal(w.Body.Bytes(), &responseClient)
	assert.NoError(t, err)
	assert.Equal(t, client.ClientID, responseClient.ClientID)
	assert.Equal(t, client.Login, responseClient.Login)
}

func TestGetClientNotFound(t *testing.T) {
	setupTestDB()
	handler := NewClientHandler(db)

	router := gin.Default()
	router.GET("/clients/:id", handler.GetClient)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/clients/00000000-0000-0000-0000-000000000000", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestBulkCreateClients(t *testing.T) {
	setupTestDB()
	handler := NewClientHandler(db)

	router := gin.Default()
	router.POST("/clients", handler.BulkCreateClients)

	clients := []models.Client{
		{ClientID: uuid.New(), Login: "login without validation", Age: 52, Location: "Moscow City", Gender: "MALE"},
		{ClientID: uuid.New(), Login: "yet another login", Age: 42, Location: "New York", Gender: "FEMALE"},
	}
	body, _ := json.Marshal(clients)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/clients", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var responseClients []models.Client
	err := json.Unmarshal(w.Body.Bytes(), &responseClients)
	assert.NoError(t, err)
	assert.Len(t, responseClients, 2)
	assert.Equal(t, clients[0].ClientID, responseClients[0].ClientID)
	assert.Equal(t, clients[1].ClientID, responseClients[1].ClientID)
	assert.Equal(t, clients[0].Login, responseClients[0].Login)
	assert.Equal(t, clients[1].Login, responseClients[1].Login)
}

func TestBulkCreateClientsBadRequest(t *testing.T) {
	setupTestDB()
	handler := NewClientHandler(db)

	router := gin.Default()
	router.POST("/clients", handler.BulkCreateClients)

	body := []byte(`invalid json`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/clients", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
