package log_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/service/log_service"
	"fmt"

	"github.com/gin-gonic/gin"
)

type LogApi struct { //在这里注册路由
}

type LogListRequest struct {
	common.PageInfo
	LogType     enum.LogType  `form:"logType"`
	Level       enum.LogLevel `form:"level"`
	IP          string        `form:"ip"`
	LoginStatus *bool         `form:"loginStatus"` // 指针:false(登录失败)也要能作为筛选条件
	ServiceName string        `form:"serviceName"`
	UserID      uint          `form:"userID"`
}

type LogListResponse struct { //还需要什么就自己加
	models.LogModel
	UserNickName string `json:"userNickName"`
	UserAvatar   string `json:"userAvatar"`
}

func (LogApi) LogListView(c *gin.Context) {
	//支持分页查询和模糊匹配&精确查询
	var req LogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithError(err, c)
		return
	}

	req.PageInfo.Order = "created_at desc"

	// DB 下沉 log 服务
	list, count, capped, err := log_service.ListLogs(
		req.LogType, req.Level, req.IP, req.ServiceName, req.LoginStatus,
		req.UserID, req.Key, req.Page, req.Limit, req.Order, req.EndId, true)
	if err != nil {
		response.FailWithError(err, c)
		return
	}

	response.OkWithListCapped(list, count, capped, c)
}

func (LogApi) LogReadView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithError(err, c)
		return
	}
	if err := log_service.ReadLog(req.ID); err != nil {
		response.FailWithMsg("日志不存在", c)
		return
	}
	response.OkWithMsg("日志已读取", c) //TODO:也许要返回日志详情?
}

func (LogApi) LogRemoveView(c *gin.Context) {
	var req models.RemoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(err, c)
		return
	}
	log := log_service.GetLog(c)
	log.ShowRequest()
	log.ShowResponse()

	deleted, err := log_service.RemoveLogs(req.IDList)
	if err != nil {
		response.FailWithMsg("删除失败", c)
		return
	}

	response.OkWithMsg(fmt.Sprintf("日志删除成功,共删除%d条", deleted), c)
}
