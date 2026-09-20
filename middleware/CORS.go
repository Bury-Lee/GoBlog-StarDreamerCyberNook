// middleware/CORS.go
package middleware

import (
	"StarDreamerCyberNook/global"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件,仅允许配置白名单中的来源跨域访问
// 说明:不再反射任意Origin,也不再携带 Allow-Credentials(项目使用Header传递token,不需要Cookie凭据)
func CORS(c *gin.Context) {
	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		//非跨域请求,直接放行
		c.Next()
		return
	}

	allowed := false
	for _, item := range global.Config.System.CORSOrigins {
		if item == origin {
			allowed = true
			break
		}
	}

	if !allowed {
		//不在白名单中:不返回任何CORS响应头,浏览器会拦截跨域读取
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
		return
	}

	h := c.Writer.Header()
	h.Set("Access-Control-Allow-Origin", origin)
	h.Set("Vary", "Origin")
	h.Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, token, refreshToken")
	h.Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

	//专门处理 OPTIONS 预检请求
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204)
		return
	}
	c.Next()
}
