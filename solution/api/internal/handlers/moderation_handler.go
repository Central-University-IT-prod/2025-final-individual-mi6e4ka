package handlers

import (
	"net/http"

	"git.mi6e4ka.dev/prod-2025/internal/dto"
	"git.mi6e4ka.dev/prod-2025/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ModerationHandler struct {
	db *gorm.DB
}

func NewModerationHandler(db *gorm.DB) *ModerationHandler {
	return &ModerationHandler{
		db: db,
	}
}

// @Summary Change moderation settings
// @Description Enable or disable moderation
// @Tags moderation
// @Accept json
// @Produce json
// @Param body body dto.ModerationBody true "Moderation settings"
// @Success 200 {object} nil
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /moderation [post]
func (h *ModerationHandler) ToggleModeration(ctx *gin.Context) {
	var body dto.ModerationBody
	if err := ctx.ShouldBindBodyWithJSON(&body); err != nil {
		ctx.AbortWithStatus(400)
		return
	}
	settings := ctx.MustGet("settings").(*models.Setting)
	settings.Moderation = *body.Moderation
	if err := h.db.Save(&settings).Error; err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK)
}

// @Summary Get unmoderated campaigns
// @Description Retrieve a list of unmoderated campaigns
// @Tags moderation
// @Accept json
// @Produce json
// @Param size query int true "Page size"
// @Param page query int true "Page number"
// @Success 200 {array} models.Campaign
// @Failure 400 {object} map[string]string
// @Router /moderation/campaigns [get]
func (h *ModerationHandler) GetUnmoderatedCampaigns(ctx *gin.Context) {
	type PaginationParams struct {
		Size *int `form:"size" binding:"required,gte=1"`
		Page *int `form:"page" binding:"required,gte=0"`
	}
	var pagination PaginationParams
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	var results []models.Campaign
	size, page := *pagination.Size, *pagination.Page
	h.db.
		Limit(size).
		Offset(page*size).
		Where("moderated = ? OR moderated IS NULL", true).
		Find(&results)

	ctx.JSON(http.StatusOK, results)
}

// @Summary Moderate a campaign
// @Description Approve or reject a campaign based on the verdict
// @Tags moderation
// @Accept json
// @Produce json
// @Param id path string true "Campaign ID" format(uuid)
// @Param body body dto.ModerationVerdictBody true "Moderation verdict"
// @Success 200 {object} nil
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /moderation/campaigns/{id} [post]
func (h *ModerationHandler) ModerateCampaign(ctx *gin.Context) {
	var params dto.PathID
	if err := ctx.ShouldBindUri(&params); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	var body dto.ModerationVerdictBody
	if err := ctx.ShouldBindBodyWithJSON(&body); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if *body.Verdict {
		h.db.Model(models.Campaign{}).Where("campaign_id = ?", params.ID).Update("moderated", true)
	} else {
		h.db.Delete(&models.Event{}, "campaign_id = ?", params.ID)
		h.db.Delete(&models.Campaign{}, "campaign_id = ?", params.ID)
	}
	ctx.Status(http.StatusOK)
}
