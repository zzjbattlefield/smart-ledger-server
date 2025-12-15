package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"smart-ledger-server/internal/model/dto"
	"smart-ledger-server/internal/pkg/response"
	"smart-ledger-server/internal/service"
	"smart-ledger-server/pkg/errcode"
)

// CategoryHandler 分类处理器
type CategoryHandler struct {
	categoryService service.CategoryServiceInterface
}

// NewCategoryHandler 创建分类处理器
func NewCategoryHandler(categoryService service.CategoryServiceInterface) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// Create 创建分类
// @Summary 创建分类
// @Tags 分类
// @Accept json
// @Produce json
// @Security Bearer
// @Param body body dto.CreateCategoryRequest true "分类信息"
// @Success 200 {object} response.Response{data=dto.CategoryResponse}
// @Router /categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}
	userID := c.GetUint64("user_id")
	resp, err := h.categoryService.Create(c.Request.Context(), userID, &req)
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

// Get 获取分类详情
// @Summary 获取分类详情
// @Tags 分类
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "分类ID"
// @Success 200 {object} response.Response{data=dto.CategoryResponse}
// @Router /categories/{id} [get]
func (h *CategoryHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的分类ID")
		return
	}

	userID := c.GetUint64("user_id")
	resp, err := h.categoryService.GetByID(c.Request.Context(), userID, id)
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

// List 获取分类列表
// @Summary 获取分类列表
// @Tags 分类
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.CategoryResponse}
// @Router /categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	userID := c.GetUint64("user_id")
	resp, err := h.categoryService.List(c.Request.Context(), userID)
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

// Update 更新分类
// @Summary 更新分类
// @Tags 分类
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "分类ID"
// @Param body body dto.UpdateCategoryRequest true "更新信息"
// @Success 200 {object} response.Response{data=dto.CategoryResponse}
// @Router /categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的分类ID")
		return
	}

	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	userID := c.GetUint64("user_id")
	resp, err := h.categoryService.Update(c.Request.Context(), userID, id, &req)
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

// Delete 删除分类
// @Summary 删除分类
// @Tags 分类
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "分类ID"
// @Success 200 {object} response.Response
// @Router /categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的分类ID")
		return
	}

	userID := c.GetUint64("user_id")
	if err := h.categoryService.Delete(c.Request.Context(), userID, id); err != nil {
		if e, ok := err.(*errcode.ErrCode); ok {
			response.Error(c, e)
			return
		}
		response.ServerError(c)
		return
	}

	response.Success(c, nil)
}
