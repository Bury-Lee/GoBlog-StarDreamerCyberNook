package follow_api

import (
	"time"

	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/service/user_service"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
)

type FollowUserListRequest struct {
	common.PageInfo
	UserID uint `form:"userID"`
}
type FollowUserListResponse struct {
	FocusUserID       uint      `json:"focusUserID"`
	FocusUserNickName string    `json:"focusUserNickname"`
	FocusUserAvatar   string    `json:"focusUserAvatar"`
	FocusUserAbstract string    `json:"focusUserAbstract"`
	CreatedAt         time.Time `json:"createdAt"`
}

// FollowUserListView 我的关注和用户的关注
func (FollowApi) FollowUserListView(c *gin.Context) {
	var req FollowUserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claims := jwts.GetClaims(c)
	if claims == nil {
		claims, _ = jwts.ParseTokenByGin(c)
	}
	if req.UserID == 0 {
		if claims == nil {
			response.FailWithMsg("请登录", c)
			return
		}
		req.UserID = claims.UserID
	}
	var viewerID uint
	if claims != nil {
		viewerID = claims.UserID
	}

	items, count, capped, err := user_service.FollowingList(req.UserID, viewerID, claims != nil, req.Page, req.Limit, req.Order, req.EndId)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			response.FailWithMsg(st.Message(), c)
		} else {
			response.FailWithMsg("查询失败", c)
		}
		return
	}

	list := make([]FollowUserListResponse, 0, len(items))
	for _, it := range items {
		list = append(list, FollowUserListResponse{
			FocusUserID:       it.FocusUserID,
			FocusUserNickName: it.Nickname,
			FocusUserAvatar:   it.Avatar,
			FocusUserAbstract: it.Abstract,
			CreatedAt:         it.CreatedAt,
		})
	}
	response.OkWithListCapped(list, count, capped, c)
}
