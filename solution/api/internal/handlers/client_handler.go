package handlers

import (
	"log"
	"net/http"

	"git.mi6e4ka.dev/prod-2025/internal/dto"
	"git.mi6e4ka.dev/prod-2025/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ClientHandler struct {
	db *gorm.DB
}

func NewClientHandler(db *gorm.DB) *ClientHandler {
	return &ClientHandler{db: db}
}

// GetClient godoc
// @Summary Get a client
// @Description Get a client by client ID
// @Tags clients
// @Param id path string true "Client ID" format(uuid)
// @Success 200 {object} models.Client
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /clients/{id} [get]
func (h *ClientHandler) GetClient(ctx *gin.Context) {
	var params dto.PathID
	if err := ctx.ShouldBindUri(&params); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var client models.Client
	if err := h.db.First(&client, "client_id = ?", params.ID).Error; err != nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	ctx.JSON(http.StatusOK, client)
}

// BulkCreateClients godoc
// @Summary Bulk create clients
// @Description Create multiple clients in bulk
// @Tags clients
// @Param clients body []models.Client true "List of clients"
// @Success 201 {array} models.Client
// @Failure 400 {object} map[string]string
// @Router /clients/bulk [post]
func (h *ClientHandler) BulkCreateClients(ctx *gin.Context) {
	var body []models.Client
	if err := ctx.ShouldBindJSON(&body); err != nil {
		log.Println(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if tx := h.db.Save(&body); tx.Error != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": tx.Error.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, body)
}
