package handlers

import (
	"net/http"

	"git.mi6e4ka.dev/prod-2025/internal/dto"
	"git.mi6e4ka.dev/prod-2025/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StatsHandler struct {
	db *gorm.DB
}

func NewStatsHandler(db *gorm.DB) *StatsHandler {
	return &StatsHandler{db: db}
}

// GetCampaignStats godoc
// @Summary Get campaign statistics
// @Description Get statistics for a specific campaign by campaign ID
// @Tags stats
// @Param id path string true "Campaign ID" format(uuid)
// @Success 200 {object} dto.Stats
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /stats/campaigns/{id} [get]
func (h *StatsHandler) GetCampaignStats(ctx *gin.Context) {
	var params dto.PathID
	if err := ctx.ShouldBindUri(&params); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	if c := h.db.Find(&models.Campaign{}, "campaign_id = ?", params.ID).RowsAffected; c == 0 {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	var stats dto.Stats
	h.db.Model(&models.Event{}).
		Select("COUNT(*) FILTER (WHERE type = 'view') AS impressions_count, "+
			"COUNT(*) FILTER (WHERE type = 'click') AS clicks_count, "+
			"SUM(cost) FILTER (WHERE type = 'click') AS spent_clicks, "+
			"SUM(cost) FILTER (WHERE type = 'view') AS spent_impressions, "+
			"SUM(cost) AS spent_total").
		Where("campaign_id = ?", params.ID).
		Find(&stats)
	if stats.ClicksCount > 0 {
		stats.Conversion = float64(stats.ImpressionsCount) / float64(stats.ClicksCount) * 100
	}
	ctx.JSON(http.StatusOK, stats)
}

// GetAdvertiserCampaignStats godoc
// @Summary Get advertiser campaign statistics
// @Description Get statistics for all campaigns of a specific advertiser by advertiser ID
// @Tags stats
// @Param id path string true "Advertiser ID" format(uuid)
// @Success 200 {object} dto.Stats
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /stats/advertisers/{id}/campaigns [get]
func (h *StatsHandler) GetAdvertiserCampaignStats(ctx *gin.Context) {
	var params dto.PathID
	if err := ctx.ShouldBindUri(&params); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	if c := h.db.Find(&models.Advertiser{}, "advertiser_id = ?", params.ID).RowsAffected; c == 0 {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	var stats dto.Stats
	h.db.Model(&models.Event{}).
		Select("COUNT(*) FILTER (WHERE type = 'view') AS impressions_count, "+
			"COUNT(*) FILTER (WHERE type = 'click') AS clicks_count, "+
			"SUM(cost) FILTER (WHERE type = 'click') AS spent_clicks, "+
			"SUM(cost) FILTER (WHERE type = 'view') AS spent_impressions, "+
			"SUM(cost) AS spent_total").
		Joins("JOIN campaigns ON campaigns.campaign_id = events.campaign_id").
		Where("campaigns.advertiser_id = ?", params.ID).
		Find(&stats)
	if stats.ClicksCount > 0 {
		stats.Conversion = float64(stats.ClicksCount) / float64(stats.ImpressionsCount) * 100
	}
	ctx.JSON(http.StatusOK, stats)
}

// GetCampaignDailyStats godoc
// @Summary Get campaign daily statistics
// @Description Get daily statistics for a specific campaign by campaign ID
// @Tags stats
// @Param id path string true "Campaign ID" format(uuid)
// @Success 200 {array} dto.DailyStats
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /stats/campaigns/{id}/daily [get]
func (h *StatsHandler) GetCampaignDailyStats(ctx *gin.Context) {
	var params dto.PathID
	if err := ctx.ShouldBindUri(&params); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	var campaign models.Campaign
	if c := h.db.Find(&campaign, "campaign_id = ?", params.ID).RowsAffected; c == 0 {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	settings := ctx.MustGet("settings").(*models.Setting)
	var dailyStats []dto.DailyStats
	maxDay := *campaign.EndDate
	if maxDay > settings.Day {
		maxDay = settings.Day
	}
	for day := *campaign.StartDate; day <= maxDay; day++ {
		var stats dto.DailyStats
		h.db.Model(&models.Event{}).
			Select("COUNT(*) FILTER (WHERE type = 'view') AS impressions_count, "+
				"COUNT(*) FILTER (WHERE type = 'click') AS clicks_count, "+
				"SUM(cost) FILTER (WHERE type = 'click') AS spent_clicks, "+
				"SUM(cost) FILTER (WHERE type = 'view') AS spent_impressions, "+
				"SUM(cost) AS spent_total").
			Where("campaign_id = ?", params.ID).
			Where("day = ?", day).
			Find(&stats)
		if stats.ClicksCount > 0 {
			stats.Conversion = float64(stats.ImpressionsCount) / float64(stats.ClicksCount) * 100
		}
		stats.Date = int64(day)
		dailyStats = append(dailyStats, stats)
	}
	ctx.JSON(http.StatusOK, dailyStats)
}

// GetAdvertiserCampaignDailyStats godoc
// @Summary Get advertiser campaign daily statistics
// @Description Get daily statistics for all campaigns of a specific advertiser by advertiser ID
// @Tags stats
// @Param id path string true "Advertiser ID" format(uuid)
// @Success 200 {array} dto.DailyStats
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /stats/advertisers/{id}/campaigns/daily [get]
func (h *StatsHandler) GetAdvertiserCampaignDailyStats(ctx *gin.Context) {
	var params dto.PathID
	if err := ctx.ShouldBindUri(&params); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	var advertiser models.Advertiser
	if c := h.db.Find(&advertiser, "advertiser_id = ?", params.ID).RowsAffected; c == 0 {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	var result struct {
		MinStartDate uint
		MaxEndDate   uint
	}

	err := h.db.Model(&models.Campaign{}).
		Select("MIN(start_date) AS min_start_date, MAX(end_date) AS max_end_date").
		Where("advertiser_id = ?", params.ID).
		Find(&result).Error
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	settings := ctx.MustGet("settings").(*models.Setting)
	var dailyStats []dto.DailyStats
	maxDay := result.MaxEndDate
	if maxDay > settings.Day {
		maxDay = settings.Day
	}
	for day := result.MinStartDate; day <= maxDay; day++ {
		var stats dto.DailyStats
		h.db.Model(&models.Event{}).
			Select("COUNT(*) FILTER (WHERE type = 'view') AS impressions_count, "+
				"COUNT(*) FILTER (WHERE type = 'click') AS clicks_count, "+
				"SUM(cost) FILTER (WHERE type = 'click') AS spent_clicks, "+
				"SUM(cost) FILTER (WHERE type = 'view') AS spent_impressions, "+
				"SUM(cost) AS spent_total").
			Joins("JOIN campaigns ON campaigns.campaign_id = events.campaign_id").
			Where("campaigns.advertiser_id = ?", params.ID).
			Where("day = ?", day).
			Find(&stats)
		if stats.ClicksCount > 0 {
			stats.Conversion = float64(stats.ImpressionsCount) / float64(stats.ClicksCount) * 100
		}
		stats.Date = int64(day)
		dailyStats = append(dailyStats, stats)
	}
	ctx.JSON(http.StatusOK, dailyStats)
}
