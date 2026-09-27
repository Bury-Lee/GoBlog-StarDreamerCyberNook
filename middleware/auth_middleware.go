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

// loadClaimRole 获取用户最新角色:优先读 Redis 缓存,未命中回查数据库并回填。
// 返回:false 表示用户不存在或已被封禁。
func loadClaimRole(userID uint) (enum.RoleType, bool) {
	if role, ok := redis_jwt.GetRoleCache(userID); ok {
		return role, role != enum.BlackRole
	}
	var user models.UserModel
	if err := global.DB.Take(&user, userID).Error; err != nil {
		return 0, false
	}
	redis_jwt.SetRoleCache(userID, user.Role)
	return user.Role, user.Role != enum.BlackRole
}

// refreshClaimRole 回查用户状态,保证封禁与角色变更即时生效
// 返回:false表示用户不存在或已被封禁
func refreshClaimRole(claim *jwts.MyClaims) bool {
	if claim == nil {
		return false
	}
	role, ok := loadClaimRole(claim.UserID)
	if !ok {
		return false
	}
	//以最新角色为准,覆盖token中的旧角色
	claim.Role = role
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
