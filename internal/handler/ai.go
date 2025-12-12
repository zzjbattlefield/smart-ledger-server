package handler

import (
	"github.com/gin-gonic/gin"

	"smart-ledger-server/internal/pkg/response"
	"smart-ledger-server/internal/service"
	"smart-ledger-server/pkg/errcode"
)

// AIHandler AI处理器
type AIHandler struct {
	aiService service.AIServiceInterface
}

// NewAIHandler 创建AI处理器
func NewAIHandler(aiService service.AIServiceInterface) *AIHandler {
	return &AIHandler{
		aiService: aiService,
	}
}

// Recognize 识别支付截图
// @Summary 识别支付截图
// @Tags AI
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param image formData file true "支付截图"
// @Success 200 {object} response.Response{data=dto.AIRecognizeResponse}
// @Router /ai/recognize [post]
func (h *AIHandler) Recognize(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		response.ParamError(c, "请上传图片")
		return
	}

	resp, err := h.aiService.RecognizeImage(c.Request.Context(), file)
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

// RecognizeAndSave 识别支付截图并保存
// @Summary 识别支付截图并保存为账单
// @Tags AI
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param image formData file true "支付截图"
// @Success 200 {object} response.Response{data=dto.BillResponse}
// @Router /ai/recognize-and-save [post]
func (h *AIHandler) RecognizeAndSave(c *gin.Context) {
	userID := c.GetUint64("user_id")

	file, err := c.FormFile("image")
	if err != nil {
		response.ParamError(c, "请上传图片")
		return
	}

	resp, err := h.aiService.RecognizeAndCreateBill(c.Request.Context(), userID, file)
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

// BatchRecognize 批量识别支付截图
// @Summary 批量识别支付截图
// @Tags AI
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param images formData file true "支付截图（最多20张）"
// @Success 200 {object} response.Response{data=dto.BatchRecognizeResponse}
// @Router /ai/batch-recognize [post]
func (h *AIHandler) BatchRecognize(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		response.ParamError(c, "请上传图片")
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		response.ParamError(c, "请至少上传一张图片")
		return
	}

	resp, err := h.aiService.BatchRecognizeImages(c.Request.Context(), files)
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

// BatchRecognizeAndSave 批量识别并保存
// @Summary 批量识别支付截图并保存为账单
// @Tags AI
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param images formData file true "支付截图（最多20张）"
// @Success 200 {object} response.Response{data=dto.BatchRecognizeAndSaveResponse}
// @Router /ai/batch-recognize-and-save [post]
func (h *AIHandler) BatchRecognizeAndSave(c *gin.Context) {
	userID := c.GetUint64("user_id")

	form, err := c.MultipartForm()
	if err != nil {
		response.ParamError(c, "请上传图片")
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		response.ParamError(c, "请至少上传一张图片")
		return
	}

	resp, err := h.aiService.BatchRecognizeAndSave(c.Request.Context(), userID, files)
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
