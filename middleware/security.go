package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders 统一设置基础安全响应头
// 说明:防止MIME嗅探、点击劫持与Referer泄露
func SecurityHeaders(c *gin.Context) {
	h := c.Writer.Header()
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Frame-Options", "DENY")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	c.Next()
}
