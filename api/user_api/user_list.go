package user_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/service/user_service"

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

	list, count, capped, err := user_service.ListUsers(req.UserID, req.Page, req.Limit, req.Order, req.Key, req.EndId)
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
