package user_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/user_service"
	"time"

	"github.com/gin-gonic/gin"
)

// UserBaseInfoResponse 用户基本信息响应结构体
// 用于返回用户的基本信息数据
type UserBaseInfoResponse struct {
	UserID        uint      `json:"userID"`        // 用户ID，唯一标识
	Age           int       `json:"age"`           // 用户年龄
	NickName      string    `json:"nickName"`      // 用户昵称
	Avatar        string    `json:"avatar"`        // 用户头像URL
	LastLoginTime time.Time `json:"lastLoginTime"` // 最后登录时间，可能为空
	Region        string    `json:"region"`        // 用户所在地区
	ExistDay      int       `json:"existDay"`      // 存在时间

	//预备字段,这种查询对于性能要求较高,先放这四个预备字段,等以后想到解决方案了再说
	ArticleCount int64 `json:"articleCount"` // 用户发布的文章数量
	FansCount    int64 `json:"fansCount"`    // 粉丝数量
	FollowCount  int64 `json:"followCount"`  // 关注数量
}

// 查看指定用户的信息
//   - 返回用户完整的基本信息数据
func (UserApi) CheckUserBaseInfoView(c *gin.Context) {

	// 1. 解析请求参数，获取要查询的用户ID
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("参数校验失败", c)
		return
	}

	// 2. 查询用户基本信息(下沉 user 服务,含文章/粉丝/关注计数)
	info, err := user_service.GetUserBaseInfo(req.ID)
	if err != nil {
		response.FailWithMsg("用户不存在呢", c)
		return
	}

	var result = UserBaseInfoResponse{
		UserID:        info.UserID,
		Age:           info.Age,
		NickName:      info.NickName,
		Avatar:        info.Avatar,
		LastLoginTime: info.LastLoginTime,
		ExistDay:      info.ExistDay,
		Region:        info.Region,

		ArticleCount: info.ArticleCount,
		FansCount:    info.FansCount,
		FollowCount:  info.FollowCount,
	}

	// 6. 返回成功响应，包含用户完整的基本信息数据
	response.OkWithData(result, c)
}
