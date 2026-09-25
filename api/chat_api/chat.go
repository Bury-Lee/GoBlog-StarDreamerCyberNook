package chat_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	xss_filter "StarDreamerCyberNook/utils/XSSfilter"
	jwts "StarDreamerCyberNook/utils/jwts"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

//TODO:这里迟点再检查了,怕有bug

type ChatSendRequest struct {
	Msg       models.ChatMsg `json:"msg"`
	RevUserID uint           `json:"revUserID"` // 接收用户ID
}

func (ChatApi) ChatSendView(c *gin.Context) { //查询会话是否存在,没有就创建一个,有的话就更新会话信息
	var req ChatSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("消息发送失败", c)
		return
	}
	claim := jwts.GetClaims(c)
	// 加入检查是否有好友关系
	var revUser models.UserFollowModel
	global.DB.Where("(user_id = ? AND focus_user_id = ?) or (user_id = ? AND focus_user_id = ?)", claim.UserID, req.RevUserID, req.RevUserID, claim.UserID).First(&revUser)
	if !revUser.Friend {
		response.FailWithMsg("您不是好友，不能发送消息", c)
		return
	}
	//TODO:以后加入拉黑机制
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

	var chatModel = models.ChatModel{
		SendUserID: claim.UserID,
		RevUserID:  req.RevUserID,
		Msg:        req.Msg,
		MsgType:    msgType,
	}
	if err := global.DB.Create(&chatModel).Error; err != nil {
		response.FailWithMsg("消息发送失败", c)
		return
	}

	//我给别人发了消息,那改变的就应该是别人的会话状态
	uniqueID := fmt.Sprintf("%d_%d", req.RevUserID, claim.UserID)
	var session models.SessionModel
	err := global.DB.Where("unique_id = ?", uniqueID).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		//首次对话:创建接收方的会话(原实现先判断session.ID导致永远返回失败,无法建立会话)
		session = models.SessionModel{
			UniqueID:        uniqueID,
			UserID:          req.RevUserID,
			LastMessageID:   chatModel.ID,
			LastMessageTime: time.Now(),
			IsRead:          false,
			UnreadCount:     1,
		}
		if err = global.DB.Create(&session).Error; err != nil {
			response.FailWithMsg("创建会话失败", c)
			return
		}
	} else if err != nil {
		response.FailWithMsg("查询会话失败", c)
		return
	} else { //更新会话信息
		session.IsRead = false
		session.UnreadCount += 1
		session.LastMessageID = chatModel.ID
		session.LastMessageTime = time.Now()
		if err = global.DB.Save(&session).Error; err != nil {
			response.FailWithMsg("更新会话失败", c)
			return
		}
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
	query := global.DB.Where("(send_user_id = ? and rev_user_id = ?) or(send_user_id = ? and rev_user_id = ?) ",
		req.UserID, claims.UserID, claims.UserID, req.UserID,
	)

	req.Order = "created_at desc"
	_list, count, capped, err := common.ListQuery[models.ChatModel](models.ChatModel{}, common.Options{
		PageInfo:      req.PageInfo,
		Preloads:      []string{"SendUserModel", "RevUserModel"},
		Where:         query,
		AllowedOrders: []string{"id", "created_at"},
		CountCap:      common.DefaultCountCap, //总数封顶
	})
	if err != nil {
		response.FailWithMsg("查询聊天记录失败", c)
		return
	}

	// 提取所有消息ID
	var chatIDs []uint
	for _, v := range _list {
		chatIDs = append(chatIDs, v.ID)
	}

	// 如果没有消息ID，直接返回空列表
	if len(chatIDs) == 0 {
		var list []ChatListResponse
		response.OkWithListCapped(list, count, capped, c)
		return
	}

	// // 根据消息ID查询用户操作记录，获取删除状态
	// var userActions []models.UserChatActionModel
	// if err := global.DB.Where("chat_id IN ? and user_id = ? and is_delete = ?", chatIDs, claims.UserID, true).
	// 	Find(&userActions).Error; err != nil {
	// 	response.FailWithMsg("查询用户操作记录失败", c)
	// 	return
	// } //一会检查一下,看看是这个字段吗

	// // 创建一个map来快速查找某个消息是否被删除
	// deletedMap := make(map[uint]bool)
	// for _, action := range userActions {
	// 	if action.IsDelete {
	// 		deletedMap[action.ChatID] = true
	// 	}
	// }

	var list = make([]ChatListResponse, 0)
	for _, model := range _list {
		// // 检查该消息是否被当前用户删除
		// if deletedMap[model.ID] {
		// 	continue // 跳过已删除的消息
		// }

		//屏蔽掉配置字段
		model.SendUserModel.Password = ""
		model.SendUserModel.UserName = ""
		model.SendUserModel.Email = ""
		model.SendUserModel.OpenID = ""
		model.SendUserModel.Role = enum.UserRole
		model.SendUserModel.UserConfModel = nil

		model.RevUserModel.Password = ""
		model.RevUserModel.UserName = ""
		model.RevUserModel.Email = ""
		model.RevUserModel.OpenID = ""
		model.RevUserModel.Role = enum.UserRole
		model.RevUserModel.UserConfModel = nil

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
	//先更新会话信息再返回响应,避免响应写出后又调用FailWithMsg导致双重响应
	//查询了就是把消息读取了
	uniqueID := fmt.Sprintf("%d_%d", claims.UserID, req.UserID)
	var session models.SessionModel
	err = global.DB.Where("unique_id = ?", uniqueID).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		session = models.SessionModel{
			UniqueID:        uniqueID,
			UserID:          claims.UserID,
			LastMessageTime: time.Now(),
			IsRead:          true,
			UnreadCount:     0,
		}
		if err = global.DB.Create(&session).Error; err != nil {
			logrus.Errorf("创建会话失败: %v", err)
		}
	} else if err != nil {
		logrus.Errorf("查询会话失败: %v", err)
	} else {
		session.IsRead = true
		session.UnreadCount = 0
		if err = global.DB.Save(&session).Error; err != nil {
			logrus.Errorf("更新会话失败: %v", err)
		}
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

	list, count, capped, err := common.ListQuery[models.SessionModel](
		models.SessionModel{},
		common.Options{
			PageInfo:      req.PageInfo,
			Where:         global.DB.Where("user_id = ?", claims.UserID),
			Preloads:      []string{"UserModel"},
			DefaultOrder:  "last_message_time desc",
			AllowedOrders: []string{"id", "created_at", "last_message_time"},
			CountCap:      common.DefaultCountCap, //总数封顶
		},
	)
	if err != nil {
		response.FailWithMsg("查询会话列表失败", c)
		return
	}

	response.OkWithListCapped(list, count, capped, c)
}
