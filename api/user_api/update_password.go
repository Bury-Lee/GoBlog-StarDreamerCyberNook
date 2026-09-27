package user_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/service/user_service"
	Hash "StarDreamerCyberNook/utils/hash"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

type UpdatePasswordRequest struct { //TODO：要经过邮箱验证
	// OldPwd string `json:"oldPwd" binding:"required"`
	NewPwd string `json:"pwd" binding:"required"`
}

func (UserApi) UpdatePasswordView(c *gin.Context) {
	var req UpdatePasswordRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithError(err, c)
		return
	}

	claims := jwts.GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		response.FailWithMsg("用户不存在", c)
		return
	}
	detail, err := user_service.GetUserDetail(claims.UserID)
	if err != nil {
		response.FailWithMsg("用户不存在", c)
		return
	}

	// 邮箱注册的、绑了邮箱的
	if detail.Email == "" {
		response.FailWithMsg("仅支持绑定邮箱的用户修改密码", c)
		return
	}

	// // 校验之前的密码
	// if !Hash.CheckPassword(user.Password, cr.OldPwd) {
	// 	response.FailWithMsg("旧密码错误", c)
	// 	return
	// }

	hashPwd, _ := Hash.HashPassword(req.NewPwd)
	if err := user_service.UpdatePassword(claims.UserID, hashPwd); err != nil {
		response.FailWithMsg("密码重置失败", c)
		return
	}
	response.OkWithMsg("密码重置成功", c)
}
