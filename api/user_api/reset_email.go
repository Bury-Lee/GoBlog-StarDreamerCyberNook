package user_api

//TODO:如果没记错的话邮箱验证码只能验证一次,输错一次就直接作废,到时候改一下,改为10次
import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/middleware"
	"StarDreamerCyberNook/service/user_service"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ResetEmailView 重置绑定邮箱
func (UserApi) ResetEmailView(c *gin.Context) {
	_email, _ := c.Get("email")
	email, ok := _email.(middleware.EmailVerifyInfo)
	if !ok {
		logrus.Error("邮箱验证信息类型断言失败")
		response.FailWithMsg("意外错误", c)
		return
	}
	if email.Type != "重置邮箱" {
		response.FailWithMsg("邮箱验证类型错误", c)
		return
	}
	claims := jwts.GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		response.FailWithMsg("不存在的用户", c)
		return
	}

	//更新数据(下沉 user 服务)
	if err := user_service.UpdateUserEmail(claims.UserID, email.RequstEmail); err != nil {
		response.FailWithMsg("邮箱重置失败", c)
		return
	}
	response.OkWithMsg("邮箱重置成功", c)
}
