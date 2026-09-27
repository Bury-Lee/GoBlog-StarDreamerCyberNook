package user_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/service/user_service"
	jwts "StarDreamerCyberNook/utils/jwts"
	"time"

	"github.com/gin-gonic/gin"
)

type UserDetailResponse struct { //个人主页的返回
	models.Model
	UserName    string            ` json:"username"` //用户名
	NickName    string            ` json:"nickname"` //昵称
	Avatar      string            ` json:"avatar"`   //头像
	Abstract    string            ` json:"abstract"` //简介
	Age         int               `json:"Age"`       //年龄
	LikeTags    []string          ` json:"likeTags"` //兴趣标签
	ContactInfo map[string]string ` json:"contactInfo"`
	Role        enum.RoleType     `json:"role"`

	//以下为配置表的字段
	UpdateUsernameDate *time.Time `json:"updateUsernameDate"` // 上次修改用户名的时间,因为可能没改过,避免无法区分nil,使用指针
	OpenFollow         bool       `json:"openFollow"`         // 公开我的关注
	OpenFans           bool       `json:"openFans"`           // 公开我的粉丝
	HomeStyleID        uint       `json:"homeStyleID"`        // 主页样式的id
}

func (UserApi) UserDetailView(c *gin.Context) {
	claims := jwts.GetClaims(c)
	if claims == nil {
		response.FailWithMsg("未登录", c)
		return
	}
	user, err := user_service.GetUserDetail(claims.UserID)
	if err != nil {
		response.FailWithMsg("用户不存在", c)
		return
	}

	var result = UserDetailResponse{
		Model:       models.Model{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt},
		UserName:    user.UserName,
		NickName:    user.NickName,
		Avatar:      user.Avatar,
		Abstract:    user.Abstract,
		Age:         user.Age,
		LikeTags:    user.LikeTags,
		ContactInfo: user.ContactInfo,
		Role:        user.Role,

		UpdateUsernameDate: user.UpdateUsernameDate,
		OpenFollow:         user.OpenFollow,
		OpenFans:           user.OpenFans,
		HomeStyleID:        user.HomeStyleID,
	}
	response.OkWithData(result, c)
}
