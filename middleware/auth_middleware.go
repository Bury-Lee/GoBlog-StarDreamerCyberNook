package middleware

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/service/redis_service/redis_jwt"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

// refreshClaimRole 回查数据库中的用户状态,保证封禁与角色变更即时生效
// 返回:false表示用户不存在或已被封禁
func refreshClaimRole(claim *jwts.MyClaims) bool {
	if claim == nil {
		return false
	}
	var user models.UserModel
	if err := global.DB.Take(&user, claim.UserID).Error; err != nil {
		return false
	}
	if user.Role == enum.BlackRole {
		return false
	}
	//以数据库中的角色为准,覆盖token中的旧角色
	claim.Role = user.Role
	return true
}

func AuthMiddleware(c *gin.Context) { //鉴权中间件
	claim, err := jwts.ParseTokenByGin(c)
	if err != nil {
		response.FailWithError(err, c)
		c.Abort()
		return
	}
	blackType, ok := redis_jwt.HasTokenByGin(c)
	if ok {
		response.FailWithMsg("由于"+blackType.String()+",服务已不可用", c)
		c.Abort()
		return
	}
	if !refreshClaimRole(claim) {
		response.FailWithMsg("账号不存在或已被封禁,服务已不可用", c)
		c.Abort()
		return
	}
	c.Set("claims", claim)
}

func AdminMiddleware(c *gin.Context) { //管理员中间件
	claim, err := jwts.ParseTokenByGin(c)
	if err != nil {
		response.FailWithMsg("无记录", c)
		c.Abort()
		return
	}
	blackType, ok := redis_jwt.HasTokenByGin(c)
	if ok {
		response.FailWithMsg("由于"+blackType.String()+",服务已不可用", c)
		c.Abort()
		return
	}
	if !refreshClaimRole(claim) {
		response.FailWithMsg("账号不存在或已被封禁,服务已不可用", c)
		c.Abort()
		return
	}
	if claim.Role != enum.AdminRole {
		response.FailWithMsg("权限不足", c)
		c.Abort()
		return
	}
	c.Set("claims", claim)
}

func VipMiddleware(c *gin.Context) { //会员中间件
	claim, err := jwts.ParseTokenByGin(c)
	if err != nil {
		response.FailWithError(err, c)
		c.Abort()
		return
	}
	blackType, ok := redis_jwt.HasTokenByGin(c)
	if ok {
		response.FailWithMsg("由于"+blackType.String()+",服务已不可用", c)
		c.Abort()
		return
	}
	if claim.Role != enum.VipRole && claim.Role != enum.AdminRole {
		response.FailWithMsg("权限不足", c)
		c.Abort()
		return
	}
}
