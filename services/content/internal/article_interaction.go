package internal

import (
	"context"
	"errors"

	contentv1 "StarDreamerCyberNook/gen/content/v1"
	"StarDreamerCyberNook/models"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// ToggleArticleDigg 切换文章点赞(未点赞→点赞;已点赞→取消)。
func (s *Server) ToggleArticleDigg(_ context.Context, req *contentv1.ToggleArticleDiggRequest) (*contentv1.ToggleArticleDiggReply, error) {
	userID, articleID := uint(req.GetUserId()), uint(req.GetArticleId())

	var article models.ArticleModel
	if err := s.db.Take(&article, "status = ? and id = ?", models.StatusPublished, articleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "文章不存在")
		}
		return nil, err
	}

	var d models.ArticleDiggModel
	err := s.db.Take(&d, "user_id = ? and article_id = ?", userID, articleID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := s.db.Create(&models.ArticleDiggModel{UserID: userID, ArticleID: articleID}).Error; err != nil {
			return nil, err
		}
		return &contentv1.ToggleArticleDiggReply{Digged: true, Changed: true}, nil
	}
	if err != nil {
		return nil, err
	}
	tx := s.db.Delete(&d)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &contentv1.ToggleArticleDiggReply{Digged: false, Changed: tx.RowsAffected > 0}, nil
}

// GetArticleInteraction 查询用户对文章的点赞/收藏状态。
func (s *Server) GetArticleInteraction(_ context.Context, req *contentv1.GetArticleInteractionRequest) (*contentv1.GetArticleInteractionReply, error) {
	userID, articleID := uint(req.GetUserId()), uint(req.GetArticleId())

	var diggCount int64
	if err := s.db.Model(&models.ArticleDiggModel{}).
		Where("user_id = ? and article_id = ?", userID, articleID).
		Count(&diggCount).Error; err != nil {
		return nil, err
	}
	var collectCount int64
	if err := s.db.Model(&models.UserArticleCollectModel{}).
		Where("user_id = ? and article_id = ?", userID, articleID).
		Count(&collectCount).Error; err != nil {
		return nil, err
	}
	return &contentv1.GetArticleInteractionReply{Digged: diggCount > 0, Collected: collectCount > 0}, nil
}

// TopArticle 置顶文章;非管理员受 max_top 限制。
func (s *Server) TopArticle(_ context.Context, req *contentv1.TopArticleRequest) (*contentv1.Empty, error) {
	userID, articleID := uint(req.GetUserId()), uint(req.GetArticleId())

	var existing models.UserTopArticleModel
	if err := s.db.Where("user_id = ? AND article_id = ?", userID, articleID).First(&existing).Error; err == nil {
		return nil, status.Error(codes.AlreadyExists, "already top")
	}
	if !req.GetIsAdmin() {
		var count int64
		s.db.Model(&models.UserTopArticleModel{}).Where("user_id = ?", userID).Count(&count)
		if count > int64(req.GetMaxTop()) {
			return nil, status.Error(codes.ResourceExhausted, "top limit")
		}
	}
	if err := s.db.Create(&models.UserTopArticleModel{UserID: userID, ArticleID: articleID}).Error; err != nil {
		return nil, err
	}
	return &contentv1.Empty{}, nil
}

// CancelArticleTop 取消置顶。
func (s *Server) CancelArticleTop(_ context.Context, req *contentv1.CancelArticleTopRequest) (*contentv1.Empty, error) {
	tx := s.db.Where("user_id = ? AND article_id = ?", uint(req.GetUserId()), uint(req.GetArticleId())).
		Delete(&models.UserTopArticleModel{})
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "no top record")
	}
	return &contentv1.Empty{}, nil
}

// AdminCancelArticleTop 管理员强制取消指定用户的置顶。
func (s *Server) AdminCancelArticleTop(ctx context.Context, req *contentv1.CancelArticleTopRequest) (*contentv1.Empty, error) {
	return s.CancelArticleTop(ctx, req)
}
