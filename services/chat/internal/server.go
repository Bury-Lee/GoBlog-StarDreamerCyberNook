package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"StarDreamerCyberNook/common"
	chatv1 "StarDreamerCyberNook/gen/chat/v1"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/pkg/dbx"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// Config 是 chat 服务的数据库配置。
type Config struct {
	Driver     string
	DSN        string
	Standalone bool // true: 自建独立库连接池;false: 复用宿主 global.DB
}

// Server 实现 ChatService。
type Server struct {
	chatv1.UnimplementedChatServiceServer
	db *gorm.DB
}

// NewServer 获取数据库句柄:未启用独立库时复用宿主 global.DB,启用时自建连接池。
func NewServer(cfg Config) (*Server, error) {
	db := global.DB
	if cfg.Standalone {
		d, err := dbx.Open(cfg.Driver, cfg.DSN)
		if err != nil {
			return nil, fmt.Errorf("chat: open db: %w", err)
		}
		db = d
	}
	if db == nil {
		return nil, fmt.Errorf("chat: 无可用数据库(共享模式需宿主先初始化 global.DB;独立模式请开启 dbStandalone)")
	}
	return &Server{db: db}, nil
}

// SendMessage 发送私信(校验好友关系,并更新接收方会话)。
func (s *Server) SendMessage(_ context.Context, req *chatv1.SendMessageRequest) (*chatv1.Empty, error) {
	sendID, revID := uint(req.GetSendUserId()), uint(req.GetRevUserId())

	var rel models.UserFollowModel
	if err := s.db.Where(
		"((user_id = ? AND focus_user_id = ?) or (user_id = ? AND focus_user_id = ?)) AND friend = ?",
		sendID, revID, revID, sendID, true).First(&rel).Error; err != nil {
		return nil, status.Error(codes.PermissionDenied, "not friend")
	}

	var msg models.ChatMsg
	if req.GetMsgJson() != "" {
		if err := json.Unmarshal([]byte(req.GetMsgJson()), &msg); err != nil {
			return nil, status.Error(codes.InvalidArgument, "bad msg json")
		}
	}
	chatModel := models.ChatModel{
		SendUserID: sendID, RevUserID: revID, Msg: msg, MsgType: models.ChatMessageType(req.GetMsgType()),
	}
	if err := s.db.Create(&chatModel).Error; err != nil {
		return nil, err
	}

	// 我给别人发消息 → 改变对方的会话状态
	uniqueID := fmt.Sprintf("%d_%d", revID, sendID)
	var session models.SessionModel
	err := s.db.Where("unique_id = ?", uniqueID).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		session = models.SessionModel{
			UniqueID: uniqueID, UserID: revID, LastMessageID: chatModel.ID,
			LastMessageTime: time.Now(), IsRead: false, UnreadCount: 1,
		}
		if err = s.db.Create(&session).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		session.IsRead = false
		session.UnreadCount += 1
		session.LastMessageID = chatModel.ID
		session.LastMessageTime = time.Now()
		if err = s.db.Save(&session).Error; err != nil {
			return nil, err
		}
	}
	return &chatv1.Empty{}, nil
}

// ListMessages 聊天记录(并将会话标记为已读)。
func (s *Server) ListMessages(_ context.Context, req *chatv1.ListMessagesRequest) (*chatv1.MessageListReply, error) {
	me, peer := uint(req.GetUserId()), uint(req.GetPeerId())
	query := s.db.Where("(send_user_id = ? and rev_user_id = ?) or (send_user_id = ? and rev_user_id = ?)",
		peer, me, me, peer)

	list, count, capped, err := common.ListQuery[models.ChatModel](s.db, models.ChatModel{}, common.Options{
		PageInfo:      common.PageInfo{Page: int(req.GetPage()), Limit: int(req.GetLimit()), Order: req.GetOrder(), EndId: uint(req.GetEndId())},
		Preloads:      []string{"SendUserModel", "RevUserModel"},
		Where:         query,
		AllowedOrders: []string{"id", "created_at"},
		CountCap:      common.DefaultCountCap,
	})
	if err != nil {
		return nil, err
	}

	rep := &chatv1.MessageListReply{Count: int64(count), Capped: capped}
	for _, m := range list {
		mj, _ := json.Marshal(m.Msg)
		rep.List = append(rep.List, &chatv1.ChatMessage{
			Id: uint64(m.ID), SendUserId: uint64(m.SendUserID), RevUserId: uint64(m.RevUserID),
			MsgType: int32(m.MsgType), MsgJson: string(mj),
			CreatedAt: m.CreatedAt.UnixMilli(), UpdatedAt: m.UpdatedAt.UnixMilli(),
			SendNickname: m.SendUserModel.NickName, SendAvatar: m.SendUserModel.Avatar,
			RevNickname: m.RevUserModel.NickName, RevAvatar: m.RevUserModel.Avatar,
		})
	}

	// 查询了即视为已读
	uniqueID := fmt.Sprintf("%d_%d", me, peer)
	var session models.SessionModel
	err = s.db.Where("unique_id = ?", uniqueID).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		_ = s.db.Create(&models.SessionModel{
			UniqueID: uniqueID, UserID: me, LastMessageTime: time.Now(), IsRead: true, UnreadCount: 0,
		}).Error
	} else if err == nil {
		session.IsRead = true
		session.UnreadCount = 0
		_ = s.db.Save(&session).Error
	}
	return rep, nil
}

// ListSessions 我的会话列表。
func (s *Server) ListSessions(_ context.Context, req *chatv1.ListSessionsRequest) (*chatv1.SessionListReply, error) {
	list, count, capped, err := common.ListQuery[models.SessionModel](s.db, models.SessionModel{}, common.Options{
		PageInfo:      common.PageInfo{Page: int(req.GetPage()), Limit: int(req.GetLimit()), Order: req.GetOrder(), EndId: uint(req.GetEndId())},
		Where:         s.db.Where("user_id = ?", uint(req.GetUserId())),
		Preloads:      []string{"UserModel", "LastMessage"},
		DefaultOrder:  "last_message_time desc",
		AllowedOrders: []string{"id", "created_at", "last_message_time"},
		CountCap:      common.DefaultCountCap,
	})
	if err != nil {
		return nil, err
	}
	rep := &chatv1.SessionListReply{Count: int64(count), Capped: capped}
	for _, m := range list {
		lmj, _ := json.Marshal(m.LastMessage.Msg)
		rep.List = append(rep.List, &chatv1.SessionItem{
			Id: uint64(m.ID), UniqueId: m.UniqueID, UserId: uint64(m.UserID), LastMessageId: uint64(m.LastMessageID),
			LastMsgJson: string(lmj), LastMsgType: int32(m.LastMessage.MsgType),
			LastMessageTime: m.LastMessageTime.UnixMilli(), UnreadCount: int32(m.UnreadCount), IsRead: m.IsRead,
			PeerNickname: m.UserModel.NickName, PeerAvatar: m.UserModel.Avatar,
		})
	}
	return rep, nil
}
