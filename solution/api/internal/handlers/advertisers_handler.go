package handlers

import (
	"net/http"

	"git.mi6e4ka.dev/prod-2025/internal/dto"
	"git.mi6e4ka.dev/prod-2025/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdvertiserHandler struct {
	db *gorm.DB
}

func NewAdvertiserHandler(db *gorm.DB) *AdvertiserHandler {
	return &AdvertiserHandler{db: db}
}

// GetAdvertiser godoc
// @Summary Get an advertiser
// @Description Get an advertiser by ID
// @Tags advertisers
// @Param id path string true "Advertiser ID" format(uuid)
// @Success 200 {object} models.Advertiser
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /advertisers/{id} [get]
func (h *AdvertiserHandler) GetAdvertiser(ctx *gin.Context) {
	var params dto.PathID
	if err := ctx.ShouldBindUri(&params); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var advertiser models.Advertiser
	if err := h.db.First(&advertiser, "advertiser_id = ?", params.ID).Error; err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "advertiser not found"})
		return
	}
	ctx.JSON(http.StatusOK, advertiser)
}

// GetAdvertiser godoc
// @Summary Bulk create advertisers
// @Description Bulk create advertisers
// @Param advertisers body []models.Advertiser true "List of advertisers"
// @Tags advertisers
// @Success 201 {object} []models.Advertiser
// @Failure 400 {object} map[string]string
// @Router /advertisers/bulk [post]
func (h *AdvertiserHandler) BulkCreateAdvertisers(ctx *gin.Context) {
	var body []models.Advertiser
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if tx := h.db.Save(&body); tx.Error != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": tx.Error.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, body)
}
