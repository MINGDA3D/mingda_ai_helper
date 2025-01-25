package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Response 统一的响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 返回成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Error 返回错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

// ValidationError 返回参数验证错误
func ValidationError(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// ServerError 返回服务器内部错误
func ServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

// UnauthorizedError 返回未授权错误
func UnauthorizedError(c *gin.Context) {
	Error(c, http.StatusUnauthorized, "unauthorized")
} 