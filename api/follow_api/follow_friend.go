package follow_api

import (
	"time"

	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/service/user_service"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

// FriendUserListRequest 好友列表请求(保护隐私,不能看别人的好友列表)
type FriendUserListRequest struct {
	common.PageInfo
}
type FriendUserListResponse struct {
	FocusUserID       uint      `json:"focusUserID"`
	FocusUserNickName string    `json:"focusUserNickname"`
	FocusUserAvatar   string    `json:"focusUserAvatar"`
	FocusUserAbstract string    `json:"focusUserAbstract"`
	CreatedAt         time.Time `json:"createdAt"`
}

// FriendUserListView 我的好友列表
func (FollowApi) FriendUserListView(c *gin.Context) {
	var req FriendUserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claim := jwts.GetClaims(c)
	if claim == nil {
		response.FailWithMsg("请登录", c)
		return
	}

	items, count, capped, err := user_service.FriendList(claim.UserID, claim.UserID, true, req.Page, req.Limit, req.Order, req.EndId)
	if err != nil {
		response.FailWithMsg("查询失败", c)
		return
	}

	list := make([]FriendUserListResponse, 0, len(items))
	for _, it := range items {
		list = append(list, FriendUserListResponse{
			FocusUserID:       it.FocusUserID,
			FocusUserNickName: it.Nickname,
			FocusUserAvatar:   it.Avatar,
			FocusUserAbstract: it.Abstract,
			CreatedAt:         it.CreatedAt,
		})
	}
	response.OkWithListCapped(list, count, capped, c)
}
