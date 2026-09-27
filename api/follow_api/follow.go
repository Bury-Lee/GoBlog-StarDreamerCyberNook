package follow_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/service/user_service"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
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

	already, err := user_service.Follow(claims.UserID, req.FocusUserID)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			response.FailWithMsg(st.Message(), c)
		} else {
			response.FailWithMsg("关注失败", c)
		}
		return
	}
	if already {
		response.OkWithMsg("已关注", c)
		return
	}
	response.OkWithMsg("关注成功", c)
}

// FollowCheckView 查询当前登录用户是否已关注指定用户
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
	followed, err := user_service.FollowCheck(claims.UserID, req.UserID)
	if err != nil {
		response.FailWithMsg("查询关注状态失败", c)
		return
	}
	response.OkWithData(gin.H{"followed": followed}, c)
}

// UnFollowUserView 登录人取关用户
func (FollowApi) UnFollowUserView(c *gin.Context) {
	var req FollowUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claims := jwts.GetClaims(c)

	if err := user_service.Unfollow(claims.UserID, req.FocusUserID); err != nil {
		if st, ok := status.FromError(err); ok {
			response.FailWithMsg(st.Message(), c)
		} else {
			response.FailWithMsg("取关失败", c)
		}
		return
	}
	response.OkWithMsg("取消关注成功", c)
}
