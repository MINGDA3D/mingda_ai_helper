package middleware

import (
	"github.com/gin-gonic/gin"
	"mingda_ai_helper/handlers"
	"mingda_ai_helper/services"
	"runtime/debug"
)

// ErrorHandler 错误处理中间件
func ErrorHandler(logService *services.LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录错误堆栈
				stack := string(debug.Stack())
				logService.Error("Panic recovered", "error", err, "stack", stack)
				
				// 返回500错误
				handlers.ServerError(c, "Internal server error")
				c.Abort()
			}
		}()
		c.Next()
	}
}

// RequestLogger 请求日志中间件
func RequestLogger(logService *services.LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 请求前记录
		logService.Info("Request received",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"client_ip", c.ClientIP(),
		)

		// 处理请求
		c.Next()

		// 请求后记录
		logService.Info("Request completed",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
		)
	}
} 