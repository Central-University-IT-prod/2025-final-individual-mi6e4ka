package handlers

import (
	"net/http"
	"strconv"

	"git.mi6e4ka.dev/prod-2025/internal/dto"
	"git.mi6e4ka.dev/prod-2025/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CampaignHandler struct {
	db *gorm.DB
}

func NewCampaignHandler(db *gorm.DB) *CampaignHandler {
	return &CampaignHandler{db: db}
}

// CreateCampaign godoc
// @Summary Create a campaign
// @Description Create a new campaign for an advertiser
// @Tags campaigns
// @Param id path string true "Advertiser ID" format(uuid)
// @Param campaign body models.Campaign true "Campaign data"
// @Success 201 {object} models.Campaign
// @Failure 400 {object} map[string]string
// @Router /advertisers/{id}/campaigns [post]
func (h *CampaignHandler) CreateCampaign(ctx *gin.Context) {
	var params dto.PathID
	if err := ctx.ShouldBindUri(&params); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	var body models.Campaign
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	body.AdvertiserID, _ = uuid.Parse(params.ID)
	body.Moderated = false
	h.db.Create(&body)
	ctx.JSON(http.StatusCreated, body)
}

// GetCampaigns godoc
// @Summary List campaigns
// @Description Get campaigns by advertiser ID with pagination
// @Tags campaigns
// @Param id path string true "Advertiser ID" format(uuid)
// @Param size query int true "Page size"
// @Param page query int true "Page number"
// @Success 200 {array} models.Campaign
// @Header 200 {integer} X-Total-Count "Total number of campaigns"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /advertisers/{id}/campaigns [get]
func (h *CampaignHandler) GetCampaigns(ctx *gin.Context) {
	var params dto.PathID
	if err := ctx.ShouldBindUri(&params); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	type PaginationParams struct {
		Size *int `form:"size" binding:"required,gte=1"`
		Page *int `form:"page" binding:"required,gte=0"`
	}
	var pagination PaginationParams
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	var advertiser models.Advertiser
	if err := h.db.First(&advertiser, "advertiser_id = ?", params.ID).Error; err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"err": "advertiser not found"})
		return
	}
	var results []models.Campaign
	size, page := *pagination.Size, *pagination.Page
	var count int64
	h.db.Model(&models.Campaign{}).Where("advertiser_id = ?", params.ID).Count(&count)
	h.db.
		Limit(size).
		Offset(page*size).
		Where("advertiser_id = ?", params.ID).
		Find(&results)
	ctx.Header("X-Total-Count", strconv.Itoa(int(count)))
	ctx.JSON(http.StatusOK, results)
}

// GetCampaign godoc
// @Summary Get a campaign
// @Description Get a campaign by advertiser ID and campaign ID
// @Tags campaigns
// @Param id path string true "Advertiser ID" format(uuid)
// @Param camp_id path string true "Campaign ID" format(uuid)
// @Success 200 {object} models.Campaign
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /advertisers/{id}/campaigns/{camp_id} [get]
func (h *CampaignHandler) GetCampaign(ctx *gin.Context) {
	var params dto.PathID
	if err := ctx.ShouldBindUri(&params); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	var campaign models.Campaign
	if err := h.db.Where("advertiser_id = ?", params.ID).First(&campaign, "campaign_id = ?", params.CampID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"err": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, campaign)
}

// UpdateCampaign godoc
// @Summary Update a campaign
// @Description Update a campaign by advertiser ID and campaign ID
// @Tags campaigns
// @Param id path string true "Advertiser ID" format(uuid)
// @Param camp_id path string true "Campaign ID" format(uuid)
// @Param campaign body models.CampaignUpdate true "Campaign update data"
// @Success 200 {object} models.Campaign
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /advertisers/{id}/campaigns/{camp_id} [put]
func (h *CampaignHandler) UpdateCampaign(ctx *gin.Context) {
	var params dto.PathID
	if err := ctx.ShouldBindUri(&params); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	var update models.CampaignUpdate
	if err := ctx.ShouldBindJSON(&update); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}

	var campaign models.Campaign
	if err := h.db.Where("advertiser_id = ?", params.ID).First(&campaign, "campaign_id = ?", params.CampID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"err": err.Error()})
		return
	}
	settings := ctx.MustGet("settings").(*models.Setting)

	if settings.Day >= *campaign.StartDate {
		if *update.ClicksLimit != *campaign.ClicksLimit || *update.ImpressionsLimit != *campaign.ImpressionsLimit {
			ctx.JSON(http.StatusConflict, gin.H{"err": "Campaign is already started"})
			return
		}
	}
	campaign.ClicksLimit = update.ClicksLimit
	campaign.ImpressionsLimit = update.ImpressionsLimit
	campaign.CostPerClick = update.CostPerClick
	campaign.CostPerImpression = update.CostPerImpression
	campaign.AdTitle = update.AdTitle
	campaign.AdText = update.AdText
	campaign.Targeting = update.Targeting

	if err := h.db.Save(&campaign).Error; err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	ctx.JSON(http.StatusOK, campaign)
}

// DeleteCampaign godoc
// @Summary Delete a campaign
// @Description Delete a campaign by advertiser ID and campaign ID
// @Tags campaigns
// @Param id path string true "Advertiser ID" format(uuid)
// @Param camp_id path string true "Campaign ID" format(uuid)
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /advertisers/{id}/campaigns/{camp_id} [delete]
func (h *CampaignHandler) DeleteCampaign(ctx *gin.Context) {
	var params dto.PathID
	if err := ctx.ShouldBindUri(&params); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	h.db.Where("campaign_id = ?", params.CampID).Delete(&models.Event{})
	if res := h.db.Where("advertiser_id = ?", params.ID).Where("campaign_id = ?", params.CampID).Delete(&models.Campaign{}); res.Error != nil || res.RowsAffected == 0 {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	ctx.Status(http.StatusNoContent)
}
