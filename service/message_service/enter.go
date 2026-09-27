// Package message_service 是经 gRPC 调用独立 message 服务的客户端(站内信)。
// Insert* 保持原签名,调用点零改动;新增读接口供 message_api 使用。
package message_service

import (
	"context"
	"errors"
	"sync"
	"time"

	messagev1 "StarDreamerCyberNook/gen/message/v1"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/pkg/svc"
)

const messageTimeout = 10 * time.Second

var (
	cliOnce sync.Once
	cli     messagev1.MessageServiceClient
	cliErr  error
)

func client() (messagev1.MessageServiceClient, error) {
	cliOnce.Do(func() {
		var cfgs map[string][]string
		if global.Config != nil {
			cfgs = global.Config.Services
		}
		svc.Init(cfgs)
		conn, err := grpcx.Dial("message")
		if err != nil {
			cliErr = err
			return
		}
		cli = messagev1.NewMessageServiceClient(conn)
	})
	return cli, cliErr
}

func call(fn func(context.Context, messagev1.MessageServiceClient) error) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), messageTimeout)
	defer cancel()
	return fn(ctx, c)
}

// InsertCommentMessage 给文章作者发送评论消息。
func InsertCommentMessage(model models.CommentModel, RevUserID uint) error {
	return call(func(ctx context.Context, c messagev1.MessageServiceClient) error {
		_, err := c.InsertComment(ctx, &messagev1.InsertCommentRequest{
			RevUserId: uint64(RevUserID), ActionUserId: uint64(model.UserID),
			ArticleId: uint64(model.ArticleID), CommentId: uint64(model.ID), Content: model.Content,
		})
		return err
	})
}

// InsertReplyMessage 给被回复的评论作者发送回复消息。
func InsertReplyMessage(model models.CommentModel, RevUserID uint) error {
	return call(func(ctx context.Context, c messagev1.MessageServiceClient) error {
		_, err := c.InsertReply(ctx, &messagev1.InsertCommentRequest{
			RevUserId: uint64(RevUserID), ActionUserId: uint64(model.UserID),
			ArticleId: uint64(model.ArticleID), CommentId: uint64(model.ID), Content: model.Content,
		})
		return err
	})
}

// InsertArticleDiggMessage 给文章作者发送点赞消息。
func InsertArticleDiggMessage(model models.ArticleDiggModel) error {
	return call(func(ctx context.Context, c messagev1.MessageServiceClient) error {
		_, err := c.InsertArticleDigg(ctx, &messagev1.InsertDiggRequest{
			ActionUserId: uint64(model.UserID), ArticleId: uint64(model.ArticleID),
		})
		return err
	})
}

// InsertCommentDiggMessage TODO:未实现。
func InsertCommentDiggMessage(model models.CommentModel, RevUserID uint) error {
	return errors.New("TODO")
}

// InsertCollectMessage 给文章作者发送收藏消息。
func InsertCollectMessage(model models.UserArticleCollectModel) error {
	return call(func(ctx context.Context, c messagev1.MessageServiceClient) error {
		_, err := c.InsertCollect(ctx, &messagev1.InsertCollectRequest{
			ActionUserId: uint64(model.UserID), ArticleId: uint64(model.ArticleID),
		})
		return err
	})
}

// InsertSystemMessage 发送系统消息。
func InsertSystemMessage(message models.MessageModel) error {
	return call(func(ctx context.Context, c messagev1.MessageServiceClient) error {
		_, err := c.InsertSystem(ctx, &messagev1.SystemMessage{
			RevUserId: uint64(message.RevUserID), ActionUserId: uint64(message.ActionUserID),
			ActionUserNickname: message.ActionUserNickname, ActionUserAvatar: message.ActionUserAvatar,
			Title: message.Title, ArticleId: uint64(message.ArticleID), ArticleTitle: message.ArticleTitle,
			CommentId: uint64(message.CommentID), Content: message.Content,
			LinkTitle: message.LinkTitle, LinkHref: message.LinkHref,
		})
		return err
	})
}

// InsertAtMessage 发送@消息。
func InsertAtMessage(model models.UserModel, ReceverUserID uint) error {
	return call(func(ctx context.Context, c messagev1.MessageServiceClient) error {
		_, err := c.InsertAt(ctx, &messagev1.InsertAtRequest{
			RevUserId: uint64(ReceverUserID), ActionUserId: uint64(model.ID),
			ActionUserNickname: model.NickName, ActionUserAvatar: model.Avatar,
		})
		return err
	})
}
