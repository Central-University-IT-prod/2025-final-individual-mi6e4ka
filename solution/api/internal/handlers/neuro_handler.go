package handlers

import (
	"net/http"

	"git.mi6e4ka.dev/prod-2025/internal/dto"
	"git.mi6e4ka.dev/prod-2025/internal/models"
	"git.mi6e4ka.dev/prod-2025/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NeuroHandler struct {
	db         *gorm.DB
	llmService *services.LLMService
}

func NewNeuroHandler(db *gorm.DB, llmService *services.LLMService) *NeuroHandler {
	return &NeuroHandler{
		db:         db,
		llmService: llmService,
	}
}

type NeuroAnswer struct {
	AdText string `json:"ad_text"`
}

// GenerateDescription godoc
// @Summary Generate ad description
// @Description Generate an advertisement description using LLM service
// @Tags neuro
// @Accept json
// @Produce json
// @Param id path string true "Advertiser ID" format(uuid)
// @Param body body dto.NeuroBody true "Neuro Body"
// @Success 200 {object} NeuroAnswer
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /neuro/{id} [post]
func (h *NeuroHandler) GenerateDescription(ctx *gin.Context) {
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
	var body dto.NeuroBody
	if err := ctx.ShouldBind(&body); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	neuroDescription, err := h.llmService.GenerateDescription(body.AdTitle, advertiser.Name)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, NeuroAnswer{AdText: neuroDescription})
}
