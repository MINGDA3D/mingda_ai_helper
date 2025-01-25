package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"mingda_ai_helper/pkg/response"
	"mingda_ai_helper/services"
	"runtime/debug"
)

// ErrorHandler 错误处理中间件
func ErrorHandler(logService services.LogInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录错误堆栈
				stack := string(debug.Stack())
				logService.Error("Panic recovered", 
					zap.Any("error", err), 
					zap.String("stack", stack),
				)
				
				// 返回500错误
				response.ServerError(c, "Internal server error")
				c.Abort()
			}
		}()
		c.Next()
	}
}

// RequestLogger 请求日志中间件
func RequestLogger(logService services.LogInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 请求前记录
		logService.Info("Request received",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
		)

		// 处理请求
		c.Next()

		// 请求后记录
		logService.Info("Request completed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
		)
	}
} 