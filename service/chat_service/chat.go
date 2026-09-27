// Package chat_service 是经 gRPC 调用独立 chat 服务的客户端(私信域)。
package chat_service

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	chatv1 "StarDreamerCyberNook/gen/chat/v1"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/pkg/svc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrNotFriend 领域错误:非好友不能发送私信。
var ErrNotFriend = errors.New("chat: not friend")

const chatTimeout = 10 * time.Second

var (
	cliOnce sync.Once
	cli     chatv1.ChatServiceClient
	cliErr  error
)

func client() (chatv1.ChatServiceClient, error) {
	cliOnce.Do(func() {
		var cfgs map[string][]string
		if global.Config != nil {
			cfgs = global.Config.Services
		}
		svc.Init(cfgs)
		conn, err := grpcx.Dial("chat")
		if err != nil {
			cliErr = err
			return
		}
		cli = chatv1.NewChatServiceClient(conn)
	})
	return cli, cliErr
}

// SendMessage 发送私信(消息内容以 JSON 透传)。
func SendMessage(sendUserID, revUserID uint, msg models.ChatMsg, msgType models.ChatMessageType) error {
	c, err := client()
	if err != nil {
		return err
	}
	mj, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), chatTimeout)
	defer cancel()
	_, err = c.SendMessage(ctx, &chatv1.SendMessageRequest{
		SendUserId: uint64(sendUserID), RevUserId: uint64(revUserID), MsgJson: string(mj), MsgType: int32(msgType),
	})
	if status.Code(err) == codes.PermissionDenied {
		return ErrNotFriend
	}
	return err
}

// ListMessages 聊天记录(服务端会将会话标记已读)。
func ListMessages(me, peer uint, page, limit int, order string, endID uint) ([]models.ChatModel, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), chatTimeout)
	defer cancel()
	rep, err := c.ListMessages(ctx, &chatv1.ListMessagesRequest{
		UserId: uint64(me), PeerId: uint64(peer), Page: int32(page), Limit: int32(limit), Order: order, EndId: uint64(endID),
	})
	if err != nil {
		return nil, 0, false, err
	}
	out := make([]models.ChatModel, 0, len(rep.GetList()))
	for _, m := range rep.GetList() {
		var msg models.ChatMsg
		if m.GetMsgJson() != "" {
			_ = json.Unmarshal([]byte(m.GetMsgJson()), &msg)
		}
		cm := models.ChatModel{
			SendUserID: uint(m.GetSendUserId()), RevUserID: uint(m.GetRevUserId()),
			MsgType: models.ChatMessageType(m.GetMsgType()), Msg: msg,
			SendUserModel: models.UserModel{NickName: m.GetSendNickname(), Avatar: m.GetSendAvatar()},
			RevUserModel:  models.UserModel{NickName: m.GetRevNickname(), Avatar: m.GetRevAvatar()},
		}
		cm.ID = uint(m.GetId())
		cm.CreatedAt = time.UnixMilli(m.GetCreatedAt())
		cm.UpdatedAt = time.UnixMilli(m.GetUpdatedAt())
		out = append(out, cm)
	}
	return out, int(rep.GetCount()), rep.GetCapped(), nil
}

// ListSessions 我的会话列表。
func ListSessions(userID uint, page, limit int, order string, endID uint) ([]models.SessionModel, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), chatTimeout)
	defer cancel()
	rep, err := c.ListSessions(ctx, &chatv1.ListSessionsRequest{
		UserId: uint64(userID), Page: int32(page), Limit: int32(limit), Order: order, EndId: uint64(endID),
	})
	if err != nil {
		return nil, 0, false, err
	}
	out := make([]models.SessionModel, 0, len(rep.GetList()))
	for _, s := range rep.GetList() {
		var msg models.ChatMsg
		if s.GetLastMsgJson() != "" {
			_ = json.Unmarshal([]byte(s.GetLastMsgJson()), &msg)
		}
		m := models.SessionModel{
			UniqueID: s.GetUniqueId(), UserID: uint(s.GetUserId()),
			LastMessageID: uint(s.GetLastMessageId()),
			LastMessage:   models.ChatModel{Msg: msg, MsgType: models.ChatMessageType(s.GetLastMsgType())},
			UnreadCount:   int(s.GetUnreadCount()), IsRead: s.GetIsRead(),
			UserModel: models.UserModel{NickName: s.GetPeerNickname(), Avatar: s.GetPeerAvatar()},
		}
		m.ID = uint(s.GetId())
		m.LastMessageTime = time.UnixMilli(s.GetLastMessageTime())
		out = append(out, m)
	}
	return out, int(rep.GetCount()), rep.GetCapped(), nil
}
