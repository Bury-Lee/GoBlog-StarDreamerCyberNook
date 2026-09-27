package follow_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/user_service"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
)

// FollowerListRequest 获取粉丝列表请求参数
type FollowerListRequest struct {
	common.PageInfo
	UserID uint `form:"userID"`
}

// FollowerListView 获取指定用户或当前登录用户的粉丝列表
func (FollowApi) FollowerListView(c *gin.Context) {
	var req FollowerListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claim, _ := jwts.ParseTokenByGin(c)
	if req.UserID == 0 {
		if claim == nil {
			response.FailWithMsg("请登录", c)
			return
		}
		req.UserID = claim.UserID
	}
	var viewerID uint
	if claim != nil {
		viewerID = claim.UserID
	}

	records, count, capped, err := user_service.FollowerList(req.UserID, viewerID, claim != nil, req.Page, req.Limit, req.Order, req.EndId)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			response.FailWithMsg(st.Message(), c)
		} else {
			response.FailWithMsg("查询粉丝列表失败", c)
		}
		return
	}

	list := make([]models.UserFollowModel, 0, len(records))
	for _, r := range records {
		list = append(list, models.UserFollowModel{
			Model:       models.Model{ID: r.ID, CreatedAt: r.CreatedAt},
			UserID:      r.UserID,
			FocusUserID: r.FocusUserID,
			Friend:      r.Friend,
		})
	}
	response.OkWithListCapped(list, count, capped, c)
}
