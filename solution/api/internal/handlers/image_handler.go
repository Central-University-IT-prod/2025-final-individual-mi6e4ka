package handlers

import (
	"net/http"
	"strings"

	"git.mi6e4ka.dev/prod-2025/internal/dto"
	"git.mi6e4ka.dev/prod-2025/internal/models"
	"git.mi6e4ka.dev/prod-2025/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ImageHandler struct {
	s3 *services.S3Client
	db *gorm.DB
}

func NewImageHandler(db *gorm.DB, s3 *services.S3Client) *ImageHandler {
	return &ImageHandler{db: db, s3: s3}
}

type UploadedImageKey struct {
	Key string `json:"key"`
}

// @Summary Upload an image
// @Description Upload an image to a campaign
// @Tags campaigns
// @Accept image/*
// @Produce json
// @Param id path string true "Advertiser ID" format(uuid)
// @Param camp_id path string true "Campaign ID" format(uuid)
// @Param file body string true "Image file" format(binary)
// @Success 200 {object} UploadedImageKey
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /advertisers/{id}/campaigns/{camp_id}/image [put]
func (h *ImageHandler) UploadImage(ctx *gin.Context) {
	var params dto.PathID
	if err := ctx.ShouldBindUri(&params); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	var campaign models.Campaign
	if err := h.db.Where("advertiser_id = ?", params.ID).First(&campaign, "campaign_id = ?", params.CampID).Error; err != nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	// upload image to s3 storage
	file, contentLength, contentType := ctx.Request.Body, ctx.Request.ContentLength, ctx.GetHeader("Content-Type")
	defer file.Close()
	if !strings.HasPrefix(contentType, "image/") {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	key, err := h.s3.UploadImage(file, contentLength, contentType)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}
	if campaign.Image != nil {
		h.s3.DeleteImage(*campaign.Image)
	}
	campaign.Image = &key
	h.db.Save(&campaign)

	ctx.JSON(http.StatusOK, &UploadedImageKey{Key: key})
}

// @Summary Get an image
// @Description Get an image from a campaign
// @Tags campaigns
// @Produce image/*
// @Param id path string true "Advertiser ID" format(uuid)
// @Param camp_id path string true "Campaign ID" format(uuid)
// @Success 200 {file} file
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /advertisers/{id}/campaigns/{camp_id}/image [get]
func (h *ImageHandler) GetImage(ctx *gin.Context) {
	var params dto.PathID
	if err := ctx.ShouldBindUri(&params); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	var campaign models.Campaign
	if err := h.db.Where("advertiser_id = ?", params.ID).First(&campaign, "campaign_id = ?", params.CampID).Error; err != nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	if campaign.Image == nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	imageReader, imageSize, imageType, err := h.s3.GetImage(*campaign.Image)
	if err != nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	ctx.DataFromReader(http.StatusOK, imageSize, imageType, imageReader, nil)
}

// @Summary Delete an image
// @Description Delete an image from a campaign
// @Tags campaigns
// @Param id path string true "Advertiser ID" format(uuid)
// @Param camp_id path string true "Campaign ID" format(uuid)
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /advertisers/{id}/campaigns/{camp_id}/image [delete]
func (h *ImageHandler) DeleteImage(ctx *gin.Context) {
	var params dto.PathID
	if err := ctx.ShouldBindUri(&params); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	var campaign models.Campaign
	if err := h.db.Where("advertiser_id = ?", params.ID).First(&campaign, "campaign_id = ?", params.CampID).Error; err != nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	if campaign.Image == nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	err := h.s3.DeleteImage(*campaign.Image)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image"})
		return
	}
	campaign.Image = nil
	h.db.Save(&campaign)
	ctx.Status(http.StatusNoContent)
}
