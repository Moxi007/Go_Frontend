package middleware

import (
	"PiliPili_Frontend/config"
	"PiliPili_Frontend/logger"
	"bytes"
	"github.com/gin-gonic/gin"
	"io"
)

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优化：非 Debug 模式下不要读取 Body
		if config.GetConfig().LogLevel == "DEBUG" && (c.Request.Method == "POST" || c.Request.Method == "PUT") {
			body, err := io.ReadAll(c.Request.Body)
			if err == nil {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
				logger.Debug("Request Body: %s", string(body))
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
