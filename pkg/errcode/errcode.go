package errcode

import (
	"fmt"
	"net/http"
)

// ErrCode 错误码类型
type ErrCode struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

// Error 实现 error 接口
func (e *ErrCode) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

// WithMessage 返回带自定义消息的错误
func (e *ErrCode) WithMessage(msg string) *ErrCode {
	return &ErrCode{
		Code:       e.Code,
		Message:    msg,
		HTTPStatus: e.HTTPStatus,
	}
}

// New 创建新错误码
func New(code int, msg string, httpStatus int) *ErrCode {
	return &ErrCode{
		Code:       code,
		Message:    msg,
		HTTPStatus: httpStatus,
	}
}

// =============== 通用错误码 (10000-19999) ===============

var (
	// Success 成功
	Success = New(0, "success", http.StatusOK)

	// ErrServer 服务器内部错误
	ErrServer = New(10001, "服务器内部错误", http.StatusInternalServerError)

	// ErrParams 参数错误
	ErrParams = New(10002, "参数错误", http.StatusBadRequest)

	// ErrNotFound 资源不存在
	ErrNotFound = New(10003, "资源不存在", http.StatusNotFound)

	// ErrTooManyRequests 请求过于频繁
	ErrTooManyRequests = New(10004, "请求过于频繁", http.StatusTooManyRequests)

	// ErrMethodNotAllowed 方法不允许
	ErrMethodNotAllowed = New(10005, "方法不允许", http.StatusMethodNotAllowed)
)

// =============== 认证错误码 (20000-29999) ===============

var (
	// ErrUnauthorized 未授权
	ErrUnauthorized = New(20001, "未授权，请先登录", http.StatusUnauthorized)

	// ErrTokenExpired Token过期
	ErrTokenExpired = New(20002, "Token已过期，请重新登录", http.StatusUnauthorized)

	// ErrTokenInvalid Token无效
	ErrTokenInvalid = New(20003, "Token无效", http.StatusUnauthorized)

	// ErrForbidden 禁止访问
	ErrForbidden = New(20004, "禁止访问", http.StatusForbidden)
)

// =============== 用户错误码 (30000-39999) ===============

var (
	// ErrUserNotFound 用户不存在
	ErrUserNotFound = New(30001, "用户不存在", http.StatusNotFound)

	// ErrUserExists 用户已存在
	ErrUserExists = New(30002, "用户已存在", http.StatusBadRequest)

	// ErrPasswordWrong 密码错误
	ErrPasswordWrong = New(30003, "密码错误", http.StatusBadRequest)

	// ErrPhoneInvalid 手机号格式错误
	ErrPhoneInvalid = New(30004, "手机号格式错误", http.StatusBadRequest)
)

// =============== 账单错误码 (40000-49999) ===============

var (
	// ErrBillNotFound 账单不存在
	ErrBillNotFound = New(40001, "账单不存在", http.StatusNotFound)

	// ErrBillCreateFailed 创建账单失败
	ErrBillCreateFailed = New(40002, "创建账单失败", http.StatusInternalServerError)

	// ErrBillUpdateFailed 更新账单失败
	ErrBillUpdateFailed = New(40003, "更新账单失败", http.StatusInternalServerError)

	// ErrBillDeleteFailed 删除账单失败
	ErrBillDeleteFailed = New(40004, "删除账单失败", http.StatusInternalServerError)

	// 导入相关错误 (45000-45999)
	ErrImportFileParse = New(45001, "文件解析失败", http.StatusInternalServerError)

	ErrImportUnsupported = New(45002, "不支持的导入格式", http.StatusBadRequest)
)

// =============== AI 错误码 (50000-59999) ===============

var (
	// ErrAIRecognizeFailed AI识别失败
	ErrAIRecognizeFailed = New(50001, "AI识别失败", http.StatusInternalServerError)

	// ErrImageTooLarge 图片过大
	ErrImageTooLarge = New(50002, "图片过大", http.StatusBadRequest)

	// ErrImageFormatInvalid 图片格式无效
	ErrImageFormatInvalid = New(50003, "图片格式无效", http.StatusBadRequest)

	// ErrAIServiceUnavailable AI服务不可用
	ErrAIServiceUnavailable = New(50004, "AI服务暂时不可用", http.StatusServiceUnavailable)
)

// =============== 分类错误码 (60000-69999) ===============

var (
	// ErrCategoryNotFound 分类不存在
	ErrCategoryNotFound = New(60001, "分类不存在", http.StatusNotFound)

	// ErrCategoryExists 分类已存在
	ErrCategoryExists = New(60002, "分类已存在", http.StatusBadRequest)

	// ErrCategoryHasChildren 分类下有子分类
	ErrCategoryHasChildren = New(60003, "分类下有子分类，无法删除", http.StatusBadRequest)

	// ErrCategoryIsSystem 系统分类不可修改
	ErrCategoryIsSystem = New(60004, "系统预设分类不可修改", http.StatusBadRequest)
)
