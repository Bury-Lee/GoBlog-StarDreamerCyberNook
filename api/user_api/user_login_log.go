package user_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/service/user_service"
	jwts "StarDreamerCyberNook/utils/jwts"
	"time"

	"github.com/gin-gonic/gin"
)

type UserLoginListRequest struct {
	common.PageInfo
	UserID    uint   `form:"userId"`
	Ip        string `form:"ip"`
	Addr      string `form:"addr"`
	StartTime string `form:"startTime"` // 起止时间的 年月日时分秒格式
	EndTime   string `form:"endTime"`
}
type UserLoginListResponse struct {
	models.UserLoginModel
	UserNickname string `json:"userNickname,omitempty"`
	UserAvatar   string `json:"userAvatar,omitempty"`
}

func (UserApi) UserLoginListView(c *gin.Context) {
	var req UserLoginListRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithMsg("请求参数错误", c)
		return
	}
	//不再接受前端传入的type参数,由服务端根据登录角色决定查询范围
	claims := jwts.GetClaims(c)
	if claims == nil {
		response.FailWithMsg("请登录", c)
		return
	}
	isAdmin := claims.Role == enum.AdminRole
	if !isAdmin {
		//普通用户只能查询自己的登录日志
		req.UserID = claims.UserID
	}

	if req.StartTime != "" {
		if _, err = time.Parse("2006-01-02 15:04:05", req.StartTime); err != nil {
			response.FailWithMsg("开始时间格式错误", c)
			return
		}
	}
	if req.EndTime != "" {
		if _, err = time.Parse("2006-01-02 15:04:05", req.EndTime); err != nil {
			response.FailWithMsg("结束时间格式错误", c)
			return
		}
	}

	items, count, capped, err := user_service.ListLoginLogs(
		req.UserID, req.Ip, req.Addr, req.StartTime, req.EndTime,
		req.Page, req.Limit, req.Order, req.EndId, isAdmin)
	if err != nil {
		response.FailWithMsg("查询失败", c)
		return
	}

	var list = make([]UserLoginListResponse, 0, len(items))
	for _, it := range items {
		var m models.UserLoginModel
		m.ID = it.ID
		m.CreatedAt = it.CreatedAt
		m.UserID = it.UserID
		m.IP = it.IP
		m.Addr = it.Addr
		m.UserAgent = it.UserAgent
		list = append(list, UserLoginListResponse{
			UserLoginModel: m,
			UserNickname:   it.Nickname,
			UserAvatar:     it.Avatar,
		})
	}

	response.OkWithListCapped(list, count, capped, c)
}
