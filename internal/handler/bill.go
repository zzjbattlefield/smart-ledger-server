package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"smart-ledger-server/internal/model/dto"
	"smart-ledger-server/internal/pkg/response"
	"smart-ledger-server/internal/service"
	"smart-ledger-server/pkg/errcode"
)

// BillHandler 账单处理器
type BillHandler struct {
	billService service.BillServiceInterface
}

// NewBillHandler 创建账单处理器
func NewBillHandler(billService service.BillServiceInterface) *BillHandler {
	return &BillHandler{
		billService: billService,
	}
}

// Create 创建账单
// @Summary 创建账单
// @Tags 账单
// @Accept json
// @Produce json
// @Security Bearer
// @Param body body dto.CreateBillRequest true "账单信息"
// @Success 200 {object} response.Response{data=dto.BillResponse}
// @Router /bills [post]
func (h *BillHandler) Create(c *gin.Context) {
	userID := c.GetUint64("user_id")

	var req dto.CreateBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	resp, err := h.billService.Create(c.Request.Context(), userID, &req)
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

// Get 获取账单详情
// @Summary 获取账单详情
// @Tags 账单
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "账单ID"
// @Success 200 {object} response.Response{data=dto.BillResponse}
// @Router /bills/{id} [get]
func (h *BillHandler) Get(c *gin.Context) {
	userID := c.GetUint64("user_id")

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的账单ID")
		return
	}

	resp, err := h.billService.GetByID(c.Request.Context(), userID, id)
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

// List 获取账单列表
// @Summary 获取账单列表
// @Tags 账单
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param start_date query string false "开始日期 (2006-01-02)"
// @Param end_date query string false "结束日期 (2006-01-02)"
// @Param category_id query int false "分类ID"
// @Param bill_type query int false "账单类型 (1:支出 2:收入)"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Response{data=dto.BillListResponse}
// @Router /bills [get]
func (h *BillHandler) List(c *gin.Context) {
	userID := c.GetUint64("user_id")

	var req dto.BillListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	resp, err := h.billService.List(c.Request.Context(), userID, &req)
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

// Update 更新账单
// @Summary 更新账单
// @Tags 账单
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "账单ID"
// @Param body body dto.UpdateBillRequest true "更新信息"
// @Success 200 {object} response.Response{data=dto.BillResponse}
// @Router /bills/{id} [put]
func (h *BillHandler) Update(c *gin.Context) {
	userID := c.GetUint64("user_id")

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的账单ID")
		return
	}

	var req dto.UpdateBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	resp, err := h.billService.Update(c.Request.Context(), userID, id, &req)
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

// Delete 删除账单
// @Summary 删除账单
// @Tags 账单
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "账单ID"
// @Success 200 {object} response.Response
// @Router /bills/{id} [delete]
func (h *BillHandler) Delete(c *gin.Context) {
	userID := c.GetUint64("user_id")

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的账单ID")
		return
	}

	if err := h.billService.Delete(c.Request.Context(), userID, id); err != nil {
		if e, ok := err.(*errcode.ErrCode); ok {
			response.Error(c, e)
			return
		}
		response.ServerError(c)
		return
	}

	response.Success(c, nil)
}
