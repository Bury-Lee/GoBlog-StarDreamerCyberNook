package message_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type OneKeyReadRequest struct {
	CommentMessage        bool `json:"commentMessage"`        // 评论消息
	DiggAndCollectMessage bool `json:"diggAndCollectMessage"` // 点赞收藏消息
	PrivateMessage        bool `json:"privateMessage"`        // 私信消息
	SystemMessage         bool `json:"systemMessage"`         // 系统消息
}

func (MessageApi) OneKeyClearView(c *gin.Context) {
	var req OneKeyReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	claim := jwts.GetClaims(c)

	// 1. 提前校验：如果全是 false，直接返回，无需数据库操作
	if !req.CommentMessage && !req.DiggAndCollectMessage &&
		!req.PrivateMessage && !req.SystemMessage {
		response.OkWithMsg("成功", c) // 或者返回提示"未选择任何消息类型"
		return
	}

	// 2. 按开关顺序收集需要更新的类型
	// 注意:不能用 map[bool] 作为映射,多个开关同时为 true 时 map 的重复key会互相覆盖
	var targetTypes []models.MessageType
	if req.CommentMessage {
		targetTypes = append(targetTypes, models.MessageTypeAt, models.MessageTypeComment, models.MessageTypeReply)
	}
	if req.DiggAndCollectMessage {
		targetTypes = append(targetTypes, models.MessageTypeDigg, models.MessageTypeCollect)
	}
	if req.PrivateMessage {
		targetTypes = append(targetTypes, models.MessageTypePrivate)
	}
	if req.SystemMessage {
		targetTypes = append(targetTypes, models.MessageTypeSystem)
	}

	// 4. 单次数据库更新
	// WHERE rev_user_id = ? AND type IN (收集到的所有类型)
	result := global.DB.Model(&models.MessageModel{}).
		Where("rev_user_id = ?", claim.UserID).
		Where("type IN ?", targetTypes).
		Update("is_read", true)

	// 5. 处理结果
	if result.Error != nil {
		logrus.Error("一键已读失败:", result.Error)
		response.FailWithMsg("服务器错误", c)
		return
	}

	// result.RowsAffected 可以用来获取实际更新了多少条记录
	response.OkWithMsg("成功", c)
}
