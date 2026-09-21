package follow_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

type FollowUserRequest struct {
	FocusUserID uint `json:"focusUserID" binding:"required"`
}

// FollowUserView 登录后关注用户
func (FollowApi) FollowUserView(c *gin.Context) {
	var req FollowUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	claims := jwts.GetClaims(c)
	if req.FocusUserID == claims.UserID {
		response.FailWithMsg("其实你时刻都在关注自己~", c)
		return
	}
	// 查关注的用户是否存在
	var user models.UserModel
	err := global.DB.Take(&user, req.FocusUserID).Error
	if err != nil {
		response.FailWithMsg("关注用户不存在", c)
		return
	}

	// 查之前是否已经关注过他了
	var focus models.UserFollowModel
	err = global.DB.Take(&focus, "user_id = ? and focus_user_id = ?", claims.UserID, user.ID).Error
	if err == nil {
		response.OkWithMsg("已关注", c)
		return
	}

	// 关注
	global.DB.Create(&models.UserFollowModel{
		UserID:      claims.UserID,
		FocusUserID: req.FocusUserID,
	})

	response.OkWithMsg("关注成功", c)
}

// FollowCheckView 查询当前登录用户是否已关注指定用户,用于主页按钮的初始状态
func (FollowApi) FollowCheckView(c *gin.Context) {
	var req struct {
		UserID uint `form:"userID" binding:"required"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claims := jwts.GetClaims(c)
	if claims == nil {
		response.FailWithMsg("请登录", c)
		return
	}
	var count int64
	if err := global.DB.Model(&models.UserFollowModel{}).
		Where("user_id = ? and focus_user_id = ?", claims.UserID, req.UserID).
		Count(&count).Error; err != nil {
		response.FailWithMsg("查询关注状态失败", c)
		return
	}
	response.OkWithData(gin.H{"followed": count > 0}, c)
}

// UnFollowUserView 登录人取关用户
func (FollowApi) UnFollowUserView(c *gin.Context) {
	var req FollowUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	claims := jwts.GetClaims(c)
	if req.FocusUserID == claims.UserID {
		response.FailWithMsg("你无法取关自己", c)
		return
	}
	// 查关注的用户是否存在
	var user models.UserModel
	err := global.DB.Take(&user, req.FocusUserID).Error
	if err != nil {
		response.FailWithMsg("取关用户不存在", c)
		return
	}

	// 查之前是否已经关注过他了
	var focus models.UserFollowModel
	err = global.DB.Take(&focus, "user_id = ? and focus_user_id = ?", claims.UserID, user.ID).Error
	if err != nil {
		response.FailWithMsg("未关注此用户", c)
		return
	}
	// 每天的取关也要有个限度？
	// 取关
	global.DB.Delete(&focus)
	response.OkWithMsg("取消关注成功", c)
}
