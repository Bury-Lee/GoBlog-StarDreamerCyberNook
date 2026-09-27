package message_service

import (
	"context"
	"time"

	messagev1 "StarDreamerCyberNook/gen/message/v1"
	"StarDreamerCyberNook/models"
)

// MessageItem 消息条目。
type MessageItem struct {
	ID                 uint
	Type               models.MessageType
	RevUserID          uint
	ActionUserID       uint
	ActionUserNickname string
	ActionUserAvatar   string
	Title              string
	Content            string
	ArticleID          uint
	ArticleTitle       string
	CommentID          uint
	LinkTitle          string
	LinkHref           string
	IsRead             bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func fromProtoMessage(m *messagev1.Message) MessageItem {
	return MessageItem{
		ID: uint(m.GetId()), Type: models.MessageType(m.GetType()),
		RevUserID: uint(m.GetRevUserId()), ActionUserID: uint(m.GetActionUserId()),
		ActionUserNickname: m.GetActionUserNickname(), ActionUserAvatar: m.GetActionUserAvatar(),
		Title: m.GetTitle(), Content: m.GetContent(),
		ArticleID: uint(m.GetArticleId()), ArticleTitle: m.GetArticleTitle(),
		CommentID: uint(m.GetCommentId()), LinkTitle: m.GetLinkTitle(), LinkHref: m.GetLinkHref(),
		IsRead: m.GetIsRead(), CreatedAt: time.UnixMilli(m.GetCreatedAt()), UpdatedAt: time.UnixMilli(m.GetUpdatedAt()),
	}
}

// List 消息列表。
func List(userID uint, typ models.MessageType, page, limit int, order string, endID uint) ([]MessageItem, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), messageTimeout)
	defer cancel()
	rep, err := c.List(ctx, &messagev1.ListRequest{
		UserId: uint64(userID), Type: int32(typ), Page: int32(page), Limit: int32(limit), Order: order, EndId: uint64(endID),
	})
	if err != nil {
		return nil, 0, false, err
	}
	list := make([]MessageItem, 0, len(rep.GetList()))
	for _, m := range rep.GetList() {
		list = append(list, fromProtoMessage(m))
	}
	return list, int(rep.GetCount()), rep.GetCapped(), nil
}

// UnreadCounts 各类型未读数。
func UnreadCounts(userID uint) (map[models.MessageType]int64, error) {
	c, err := client()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), messageTimeout)
	defer cancel()
	rep, err := c.UnreadCounts(ctx, &messagev1.UnreadRequest{UserId: uint64(userID)})
	if err != nil {
		return nil, err
	}
	out := make(map[models.MessageType]int64, len(rep.GetCounts()))
	for t, n := range rep.GetCounts() {
		out[models.MessageType(t)] = n
	}
	return out, nil
}

// Remove 删除消息。
func Remove(userID uint, ids []uint) error {
	c, err := client()
	if err != nil {
		return err
	}
	req := &messagev1.BatchRequest{UserId: uint64(userID)}
	for _, id := range ids {
		req.Ids = append(req.Ids, uint64(id))
	}
	ctx, cancel := context.WithTimeout(context.Background(), messageTimeout)
	defer cancel()
	_, err = c.Remove(ctx, req)
	return err
}

// OneKeyRead 一键已读。
func OneKeyRead(userID uint, comment, diggCollect, private, system bool) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), messageTimeout)
	defer cancel()
	_, err = c.OneKeyRead(ctx, &messagev1.OneKeyReadRequest{
		UserId: uint64(userID), CommentMessage: comment,
		DiggAndCollectMessage: diggCollect, PrivateMessage: private, SystemMessage: system,
	})
	return err
}

// Conf 用户消息配置。
type Conf struct {
	UserID             uint
	OpenCommentMessage bool
	OpenReplyMessage   bool
	OpenDiggMessage    bool
	OpenCollectMessage bool
	OpenPrivateMessage bool
}

// GetConf 读取用户消息配置。
func GetConf(userID uint) (Conf, error) {
	c, err := client()
	if err != nil {
		return Conf{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), messageTimeout)
	defer cancel()
	rep, err := c.GetConf(ctx, &messagev1.ConfUserRequest{UserId: uint64(userID)})
	if err != nil {
		return Conf{}, err
	}
	return Conf{
		UserID: uint(rep.GetUserId()), OpenCommentMessage: rep.GetOpenCommentMessage(),
		OpenReplyMessage: rep.GetOpenReplyMessage(), OpenDiggMessage: rep.GetOpenDiggMessage(),
		OpenCollectMessage: rep.GetOpenCollectMessage(), OpenPrivateMessage: rep.GetOpenPrivateMessage(),
	}, nil
}

// UpdateConf 更新用户消息配置(指针为 nil 的字段不更新)。
func UpdateConf(userID uint, comment, reply, digg, collect, private *bool) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), messageTimeout)
	defer cancel()
	_, err = c.UpdateConf(ctx, &messagev1.UpdateConfRequest{
		UserId: uint64(userID), OpenCommentMessage: comment, OpenReplyMessage: reply,
		OpenDiggMessage: digg, OpenCollectMessage: collect, OpenPrivateMessage: private,
	})
	return err
}
