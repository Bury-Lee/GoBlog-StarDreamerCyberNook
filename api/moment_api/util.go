package moment_api

import (
	"StarDreamerCyberNook/models/enum"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

// viewerOf 读取可选的访问者身份(未登录返回 0,false)
func viewerOf(c *gin.Context) (userID uint, isAdmin bool) {
	claims, err := jwts.ParseTokenByGin(c)
	if err != nil || claims == nil {
		return 0, false
	}
	return claims.UserID, claims.Role == enum.AdminRole
}
