package handlers

import (
	"net/http"

	"git.mi6e4ka.dev/prod-2025/internal/dto"
	"git.mi6e4ka.dev/prod-2025/internal/models"
	"git.mi6e4ka.dev/prod-2025/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdHandler struct {
	adService *services.AdService
	db        *gorm.DB
}

func NewAdHandler(db *gorm.DB) *AdHandler {
	return &AdHandler{db: db, adService: services.NewAdService(db)}
}

// GetAds godoc
// @Summary Get ads
// @Description Get ads by client ID
// @Tags ads
// @Param client_id query string true "Client ID" format(uuid)
// @Success 200 {array} dto.AdUser
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /ads [get]
func (h *AdHandler) GetAds(ctx *gin.Context) {
	var query dto.QueryClient
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"err": err.Error()})
		return
	}

	settings := ctx.MustGet("settings").(*models.Setting)
	campaignsWithScores, err := h.adService.GetAds(uuid.MustParse(query.ClientID), settings.Day, settings.Moderation)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.AbortWithStatus(404)
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch campaigns", "details": err.Error()})
		return
	}

	ctx.JSON(200, campaignsWithScores)
}

// ClickAd godoc
// @Summary Click an ad
// @Description Register a click on an ad by ID
// @Tags ads
// @Param id path string true "Ad ID" format(uuid)
// @Param client_id body dto.QueryClient true "Client ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /ads/{id}/click [post]
func (h *AdHandler) ClickAd(ctx *gin.Context) {
	var params dto.PathID
	if err := ctx.ShouldBindUri(&params); err != nil {
		ctx.AbortWithStatusJSON(400, err.Error())
		return
	}
	var body dto.QueryClient
	if err := ctx.ShouldBindBodyWithJSON(&body); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"err": err.Error()})
		return
	}
	settings := ctx.MustGet("settings").(*models.Setting)

	if err := h.adService.ClickAd(uuid.MustParse(params.ID), uuid.MustParse(body.ClientID), settings.Day); err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.AbortWithStatus(404)
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to click ad", "details": err.Error()})
		return
	}
	ctx.Status(204)
}
