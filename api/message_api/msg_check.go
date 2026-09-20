package message_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

func (MessageApi) SiteMessageCheckView(c *gin.Context) {
	claim := jwts.GetClaims(c)

	// 定义一个临时结构体来接收查询结果
	type MessageCount struct {
		MessageType models.MessageType `json:"message_type"` // 查询的字段名
		Count       int64              `json:"count"`        // COUNT(*) 的结果,用int64避免溢出
	}

	var counts []MessageCount
	// 使用 DB.Raw 或者更推荐的 Model 方式进行聚合查询
	// 与消息列表保持一致:过滤掉自己给自己产生的消息
	if err := global.DB.Model(&models.MessageModel{}).
		Select("type, count(*) as count"). // 选择类型和计数
		Where("rev_user_id = ? AND is_read = ? AND action_user_id <> ?", claim.UserID, false, claim.UserID).
		Group("type"). // 按类型分组
		Find(&counts).Error; err != nil {
		response.FailWithMsg("未读消息查询失败", c)
		return
	}

	// 组装结果
	IsRead := make(map[models.MessageType]int64)
	for _, item := range counts {
		IsRead[item.MessageType] = item.Count // 将每种类型的未读数量存入map
	}

	response.OkWithData(IsRead, c)
}
