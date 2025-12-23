package handler

import (
	"github.com/gin-gonic/gin"

	"smart-ledger-server/internal/model/dto"
	"smart-ledger-server/internal/pkg/response"
	"smart-ledger-server/internal/service"
	"smart-ledger-server/pkg/errcode"
)

// StatsHandler 统计处理器
type StatsHandler struct {
	statsService service.StatsServiceInterface
}

// NewStatsHandler 创建统计处理器
func NewStatsHandler(statsService service.StatsServiceInterface) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
	}
}

// GetSummary 获取统计摘要
// @Summary 获取统计摘要
// @Tags 统计
// @Accept json
// @Produce json
// @Security Bearer
// @Param period query string true "统计周期 (day/week/month/year)"
// @Param date query string true "日期 (day:2006-01-02, week:2006-01-02, month:2006-01, year:2006)"
// @Success 200 {object} response.Response{data=dto.StatsSummaryResponse}
// @Router /stats/summary [get]
func (h *StatsHandler) GetSummary(c *gin.Context) {
	userID := c.GetUint64("user_id")

	var req dto.StatsSummaryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	resp, err := h.statsService.GetSummary(c.Request.Context(), userID, &req)
	if err != nil {
		if e, ok := err.(*errcode.ErrCode); ok {
			response.Error(c, e)
			return
		}
		response.ServerError(c)
		return
	}

	response.Success(c, resp)
}

func (h *StatsHandler) GetSecondaryCategoryStats(c *gin.Context) {
	req := &dto.StatsSecondaryCategoryRequest{}
	userID := c.GetUint64("user_id")
	if err := c.ShouldBindQuery(req); err != nil {
		response.ParamError(c, err.Error())
		return
	}
	resp, err := h.statsService.GetSecondaryCategoryStats(c.Request.Context(), userID, req)
	if err != nil {
		if e, ok := err.(*errcode.ErrCode); ok {
			response.Error(c, e)
			return
		} else {
			response.ServerError(c)
			return
		}
	}
	response.Success(c, resp)
}

// GetCategoryStats 获取分类统计
// @Summary 获取分类统计
// @Tags 统计
// @Accept json
// @Produce json
// @Security Bearer
// @Param period query string true "统计周期 (day/week/month/year)"
// @Param date query string true "日期"
// @Success 200 {object} response.Response{data=dto.CategoryStatsResponse}
// @Router /stats/category [get]
func (h *StatsHandler) GetCategoryStats(c *gin.Context) {
	userID := c.GetUint64("user_id")

	var req dto.StatsCategoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	resp, err := h.statsService.GetCategoryStats(c.Request.Context(), userID, &req)
	if err != nil {
		if e, ok := err.(*errcode.ErrCode); ok {
			response.Error(c, e)
			return
		}
		response.ServerError(c)
		return
	}

	response.Success(c, resp)
}
