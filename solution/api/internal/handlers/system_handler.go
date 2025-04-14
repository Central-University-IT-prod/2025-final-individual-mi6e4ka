package handlers

import (
	"net/http"

	"git.mi6e4ka.dev/prod-2025/internal/dto"
	"git.mi6e4ka.dev/prod-2025/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SystemHandler struct {
	db *gorm.DB
}

func NewSystemHandler(db *gorm.DB) *SystemHandler {
	return &SystemHandler{db: db}
}

// SetMLScore godoc
// @Summary Set ML score
// @Description Set the ML score for a specific entities
// @Tags system
// @Accept json
// @Produce json
// @Param body body models.MLScore true "ML Score"
// @Success 200
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /ml-scores [post]
func (h *SystemHandler) SetMLScore(ctx *gin.Context) {
	var body models.MLScore
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	if err := h.db.Save(&body).Error; err != nil {
		if err == gorm.ErrForeignKeyViolated {
			// попытка связать несуществующих клиента и/или рекламодателя
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	ctx.Status(http.StatusOK)
}

// AdvanceTime godoc
// @Summary Advance time
// @Description Advance the current date in the system
// @Tags system
// @Accept json
// @Produce json
// @Param body body dto.TimeSetBody true "Time Set Body"
// @Success 200 {object} dto.TimeSetBody
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /time/advance [post]
func (h *SystemHandler) AdvanceTime(ctx *gin.Context) {
	var body dto.TimeSetBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	ctx.MustGet("settings").(*models.Setting).Day = *body.CurrentDate
	if err := h.db.Model(&models.Setting{ID: 1}).Update("day", body.CurrentDate).Error; err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, body)
}
