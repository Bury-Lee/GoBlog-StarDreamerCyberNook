package chat_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/chat_service"
	xss_filter "StarDreamerCyberNook/utils/XSSfilter"
	jwts "StarDreamerCyberNook/utils/jwts"
	"errors"

	"github.com/gin-gonic/gin"
)

//TODO:这里迟点再检查了,怕有bug

type ChatSendRequest struct {
	Msg       models.ChatMsg `json:"msg"`
	RevUserID uint           `json:"revUserID"` // 接收用户ID
}

func (ChatApi) ChatSendView(c *gin.Context) {
	var req ChatSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("消息发送失败", c)
		return
	}
	claim := jwts.GetClaims(c)

	var msgType models.ChatMessageType
	if req.Msg.ImageMsg != nil {
		msgType = msgType | models.ImageMsgType
	}
	if req.Msg.MarkdownMsg != nil {
		msgType = msgType | models.MarkdownMsgType
	}
	if req.Msg.TextMsg != nil {
		msgType = msgType | models.TextMsgType
	}
	if msgType == 0 {
		response.FailWithMsg("不能发送空消息", c)
		return
	}

	//消息内容防xss注入
	if req.Msg.TextMsg != nil {
		req.Msg.TextMsg.Content = xss_filter.SanitizeText(req.Msg.TextMsg.Content)
	}
	if req.Msg.MarkdownMsg != nil {
		req.Msg.MarkdownMsg.Content = xss_filter.NewXSSFilter().Sanitize(req.Msg.MarkdownMsg.Content)
	}

	// 好友校验与落库(含会话更新)下沉 chat 服务
	if err := chat_service.SendMessage(claim.UserID, req.RevUserID, req.Msg, msgType); err != nil {
		if errors.Is(err, chat_service.ErrNotFriend) {
			response.FailWithMsg("您不是好友，不能发送消息", c)
			return
		}
		response.FailWithMsg("消息发送失败", c)
		return
	}
	response.OkWithMsg("发送成功", c)
}

type ChatListRequest struct {
	common.PageInfo
	UserID uint `form:"userID" binding:"required"` // 查我和他的聊天记录
}

type ChatListResponse struct {
	models.ChatModel
	SendUserNickname string `json:"sendUserNickname"`
	SendUserAvatar   string `json:"sendUserAvatar"`
	RevUserNickname  string `json:"revUserNickname"`
	RevUserAvatar    string `json:"revUserAvatar"`
	IsMe             bool   `json:"isMe"` // 是自己发的吗
}

// TODO:撤回和删除消息功能

func (ChatApi) ChatListView(c *gin.Context) { // 查自己和指定用户的聊天记录
	var req ChatListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	claims := jwts.GetClaims(c)
	list0, count, capped, err := chat_service.ListMessages(
		claims.UserID, req.UserID, req.Page, req.Limit, "created_at desc", req.EndId)
	if err != nil {
		response.FailWithMsg("查询聊天记录失败", c)
		return
	}

	var list = make([]ChatListResponse, 0, len(list0))
	for _, model := range list0 {
		item := ChatListResponse{
			ChatModel:        model,
			SendUserNickname: model.SendUserModel.NickName,
			SendUserAvatar:   model.SendUserModel.Avatar,
			RevUserNickname:  model.RevUserModel.NickName,
			RevUserAvatar:    model.RevUserModel.Avatar,
		}
		if model.SendUserID == claims.UserID {
			item.IsMe = true
		}
		list = append(list, item)
	}

	response.OkWithListCapped(list, count, capped, c)
}

type SessionListRequest struct {
	common.PageInfo
}

func (ChatApi) SessionListView(c *gin.Context) { // 查我的会话列表
	var req SessionListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	claims := jwts.GetClaims(c)

	list, count, capped, err := chat_service.ListSessions(
		claims.UserID, req.Page, req.Limit, "last_message_time desc", req.EndId)
	if err != nil {
		response.FailWithMsg("查询会话列表失败", c)
		return
	}

	response.OkWithListCapped(list, count, capped, c)
}
