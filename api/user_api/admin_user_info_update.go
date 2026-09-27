package user_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/service/redis_service/redis_jwt"
	"StarDreamerCyberNook/service/user_service"
	utils_other "StarDreamerCyberNook/utils/other"
	"encoding/json"

	"github.com/gin-gonic/gin"
)

type AdminUserInfoUpdateRequest struct { //这些是最容易出现违规的地方
	UserID   uint           `json:"userID" binding:"required"`
	Username *string        `json:"username" s-u:"user_name"`
	Nickname *string        `json:"nickname" s-u:"nick_name"`
	Avatar   *string        `json:"avatar" s-u:"avatar"`
	Abstract *string        `json:"abstract" s-u:"abstract"`
	Role     *enum.RoleType `json:"role" s-u:"role"`
}

func (UserApi) AdminUserInfoUpdateView(c *gin.Context) {
	var req AdminUserInfoUpdateRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	userMap := utils_other.StructToMap(req, "s-u")
	b, err := json.Marshal(userMap)
	if err != nil {
		response.FailWithMsg("用户信息修改失败", c)
		return
	}

	if err := user_service.UpdateUser(req.UserID, string(b), ""); err != nil {
		response.FailWithMsg("用户信息修改失败", c)
		return
	}
	//角色变更(含封禁)时失效缓存,保证鉴权即时生效
	if req.Role != nil {
		redis_jwt.InvalidateRoleCache(req.UserID)
	}

	response.OkWithMsg("用户信息修改成功", c)
}
