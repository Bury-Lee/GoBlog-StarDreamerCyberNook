package internal

import (
	"context"
	"errors"
	"time"

	"StarDreamerCyberNook/common"
	contentv1 "StarDreamerCyberNook/gen/content/v1"
	"StarDreamerCyberNook/models"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// RecordArticleLook 记录一次浏览;当日已读过则不重复创建。
func (s *Server) RecordArticleLook(_ context.Context, req *contentv1.RecordArticleLookRequest) (*contentv1.RecordArticleLookReply, error) {
	userID, articleID := uint(req.GetUserId()), uint(req.GetArticleId())

	var article models.ArticleModel
	if err := s.db.Take(&article, "status = ? and id = ?", models.StatusPublished, articleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "文章不存在")
		}
		return nil, err
	}

	var history models.UserArticleHistoryModel
	err := s.db.Take(&history, "user_id = ? and article_id = ? and created_at >= ?",
		userID, articleID, time.Now().Format("2006-01-02")+" 00:00:00").Error
	if err == nil {
		return &contentv1.RecordArticleLookReply{Created: false, ArticleTitle: article.Title}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err := s.db.Create(&models.UserArticleHistoryModel{
		UserID: userID, ArticleName: article.Title, ArticleID: article.ID,
	}).Error; err != nil {
		return nil, err
	}
	return &contentv1.RecordArticleLookReply{Created: true, ArticleTitle: article.Title}, nil
}

// ListArticleLook 浏览足迹列表(非本人访问遵循 OpenHistory 隐私)。
func (s *Server) ListArticleLook(_ context.Context, req *contentv1.ListArticleLookRequest) (*contentv1.ListArticleLookReply, error) {
	userID := uint(req.GetUserId())
	if !(req.GetViewerLogged() && uint(req.GetViewerId()) == userID) {
		var u models.UserModel
		if err := s.db.Take(&u, userID).Error; err != nil {
			return nil, status.Error(codes.NotFound, "用户不存在")
		}
		var conf models.UserConfModel
		if err := s.db.Take(&conf, "user_id = ?", userID).Error; err != nil || !conf.OpenHistory {
			return nil, status.Error(codes.PermissionDenied, "用户未公开浏览记录")
		}
	}

	opts := common.Options{
		PageInfo:      common.PageInfo{Page: int(req.GetPage()), Limit: int(req.GetLimit()), Order: req.GetOrder(), EndId: uint(req.GetEndId())},
		Likes:         []string{"article_name"},
		Preloads:      []string{"UserModel", "ArticleModel"},
		AllowedOrders: []string{"id", "created_at"},
		CountCap:      common.DefaultCountCap,
	}
	list, count, capped, err := common.ListQuery[models.UserArticleHistoryModel](s.db, models.UserArticleHistoryModel{UserID: userID}, opts)
	if err != nil {
		return nil, err
	}
	rep := &contentv1.ListArticleLookReply{Count: int64(count), Capped: capped}
	for _, m := range list {
		rep.List = append(rep.List, &contentv1.ArticleLookItem{
			Id: uint64(m.ID), LookDate: m.CreatedAt.UnixMilli(),
			Title: m.ArticleModel.Title, Cover: m.ArticleModel.Cover,
			Nickname: m.UserModel.NickName, Avatar: m.UserModel.Avatar,
			UserId: uint64(m.UserID), ArticleId: uint64(m.ArticleID),
		})
	}
	return rep, nil
}

// RemoveArticleLook 删除浏览历史。
func (s *Server) RemoveArticleLook(_ context.Context, req *contentv1.RemoveArticleLookRequest) (*contentv1.RemoveArticleLookReply, error) {
	ids := make([]uint, 0, len(req.GetIds()))
	for _, id := range req.GetIds() {
		ids = append(ids, uint(id))
	}
	if len(ids) == 0 {
		return &contentv1.RemoveArticleLookReply{}, nil
	}
	tx := s.db.Where("user_id = ? and id in ?", uint(req.GetUserId()), ids).Delete(&models.UserArticleHistoryModel{})
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &contentv1.RemoveArticleLookReply{Deleted: int32(tx.RowsAffected)}, nil
}
