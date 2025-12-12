package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"smart-ledger-server/pkg/errcode"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errcode.Success.Code,
		Message: errcode.Success.Message,
		Data:    data,
	})
}

// SuccessWithMessage 成功响应带自定义消息
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errcode.Success.Code,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, err *errcode.ErrCode) {
	c.JSON(err.HTTPStatus, Response{
		Code:    err.Code,
		Message: err.Message,
	})
}

// ErrorWithData 错误响应带数据
func ErrorWithData(c *gin.Context, err *errcode.ErrCode, data interface{}) {
	c.JSON(err.HTTPStatus, Response{
		Code:    err.Code,
		Message: err.Message,
		Data:    data,
	})
}

// ParamError 参数错误响应
func ParamError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    errcode.ErrParams.Code,
		Message: message,
	})
}

// ServerError 服务器错误响应
func ServerError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    errcode.ErrServer.Code,
		Message: errcode.ErrServer.Message,
	})
}

// Unauthorized 未授权响应
func Unauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    errcode.ErrUnauthorized.Code,
		Message: errcode.ErrUnauthorized.Message,
	})
}

// Forbidden 禁止访问响应
func Forbidden(c *gin.Context) {
	c.JSON(http.StatusForbidden, Response{
		Code:    errcode.ErrForbidden.Code,
		Message: errcode.ErrForbidden.Message,
	})
}

// NotFound 资源不存在响应
func NotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, Response{
		Code:    errcode.ErrNotFound.Code,
		Message: errcode.ErrNotFound.Message,
	})
}

// PageResponse 分页响应结构
type PageResponse struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	List     interface{} `json:"list"`
}

// SuccessWithPage 成功响应带分页
func SuccessWithPage(c *gin.Context, total int64, page, pageSize int, list interface{}) {
	Success(c, PageResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		List:     list,
	})
}
