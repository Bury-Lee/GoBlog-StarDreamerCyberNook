package moment_api

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	jwts "StarDreamerCyberNook/utils/jwts"
	"encoding/json"

	"github.com/gin-gonic/gin"
)

// canViewMoment 判断访问者是否可见该动态
func canViewMoment(moment models.MomentModel, viewerID uint, isAdmin bool) bool {
	//本人或管理员可见全部
	if isAdmin || (viewerID != 0 && viewerID == moment.UserID) {
		return true
	}
	//他人只能看到已发布内容
	if moment.Status != models.StatusPublished {
		return false
	}
	switch moment.Visibility {
	case models.MomentVisibilityPublic:
		return true
	case models.MomentVisibilityFriends:
		if viewerID == 0 {
			return false
		}
		//好友判定:访问者对作者存在 friend=true 的关注记录(双向关注)
		var rel models.UserFollowModel
		return global.DB.Take(&rel, "user_id = ? and focus_user_id = ? and friend = ?", viewerID, moment.UserID, true).Error == nil
	default: //私密
		return false
	}
}

// viewerOf 读取可选的访问者身份(未登录返回 0,false)
func viewerOf(c *gin.Context) (userID uint, isAdmin bool) {
	claims, err := jwts.ParseTokenByGin(c)
	if err != nil || claims == nil {
		return 0, false
	}
	return claims.UserID, claims.Role == enum.AdminRole
}

// imagesJSON 将图片URL列表序列化为JSON字符串(用于 map 增量更新,serializer 不会自动生效)
func imagesJSON(images []string) string {
	if len(images) == 0 {
		return "[]"
	}
	b, err := json.Marshal(images)
	if err != nil {
		return "[]"
	}
	return string(b)
}
