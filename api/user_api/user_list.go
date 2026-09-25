package user_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type UserListRequest struct {
	common.PageInfo
	UserID uint `form:"userID"` //按用户名精确搜索
}

// UserListResponse 用户列表响应,只暴露公开字段,避免泄露邮箱/OpenID/联系方式等隐私信息
type UserListResponse struct {
	ID       uint   `json:"userID"`
	NickName string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Abstract string `json:"abstract"`
}

func (UserApi) UserListView(c *gin.Context) {
	var req UserListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	var options common.Options
	options.PageInfo = req.PageInfo
	options.Likes = []string{"nick_name"} //仅按昵称模糊匹配,不搜索简介
	options.DefaultOrder = "id desc"
	options.AllowedOrders = []string{"id", "created_at", "last_login_time", "age"}
	options.CountCap = common.DefaultCountCap //总数封顶
	if req.UserID != 0 {
		//用户名精确匹配
		options.Where = global.DB.Where("id = ?", req.UserID)
	}
	list, count, capped, err := common.ListQuery[models.UserModel](models.UserModel{}, options)
	if err != nil {
		logrus.Errorf("查询用户列表失败 %s", err)
		response.FailWithMsg("查询失败", c)
		return
	}
	result := make([]UserListResponse, 0, len(list))
	for _, item := range list {
		result = append(result, UserListResponse{
			ID:       item.ID,
			NickName: item.NickName,
			Avatar:   item.Avatar,
			Abstract: item.Abstract,
		})
	}
	response.OkWithListCapped(result, count, capped, c)
}
