package internal

import (
	"context"

	"StarDreamerCyberNook/common"
	contentv1 "StarDreamerCyberNook/gen/content/v1"
	"StarDreamerCyberNook/models"
)

// ListArticlesByStatus 按状态查询文章(可选用户/ID 过滤,分页)。
func (s *Server) ListArticlesByStatus(_ context.Context, req *contentv1.ListArticlesByStatusRequest) (*contentv1.ArticleListReply, error) {
	where := s.db.Where("status = ?", req.GetStatus())
	if req.GetUserId() != 0 {
		where = where.Where("user_id = ?", uint(req.GetUserId()))
	}
	if len(req.GetIds()) > 0 {
		ids := make([]uint, 0, len(req.GetIds()))
		for _, id := range req.GetIds() {
			ids = append(ids, uint(id))
		}
		where = where.Where("id in ?", ids)
	}
	opts := common.Options{
		PageInfo:      common.PageInfo{Page: int(req.GetPage()), Limit: int(req.GetLimit()), Order: req.GetOrder(), EndId: uint(req.GetEndId())},
		Where:         where,
		Likes:         []string{"title"},
		Preloads:      []string{"UserModel"},
		AllowedOrders: []string{"id", "created_at", "status"},
		CountCap:      common.DefaultCountCap,
	}
	list, count, capped, err := common.ListQuery[models.ArticleModel](s.db, models.ArticleModel{}, opts)
	if err != nil {
		return nil, err
	}
	rep := &contentv1.ArticleListReply{Count: int64(count), Capped: capped}
	for _, m := range list {
		rep.List = append(rep.List, toProtoArticle(m))
	}
	return rep, nil
}

// PendingBatch 游标分批取待审核文章(供定时任务)。
func (s *Server) PendingBatch(_ context.Context, req *contentv1.PendingBatchRequest) (*contentv1.ArticleListReply, error) {
	var list []models.ArticleModel
	if err := s.db.Where("status = ? and id > ?", models.StatusPending, uint(req.GetAfterId())).
		Order("id asc").Limit(int(req.GetLimit())).Find(&list).Error; err != nil {
		return nil, err
	}
	rep := &contentv1.ArticleListReply{}
	for _, m := range list {
		rep.List = append(rep.List, toProtoArticle(m))
	}
	return rep, nil
}

// CountArticlesByStatus 统计某状态文章数。
func (s *Server) CountArticlesByStatus(_ context.Context, req *contentv1.CountArticlesRequest) (*contentv1.CountArticlesReply, error) {
	var count int64
	if err := s.db.Model(&models.ArticleModel{}).Where("status = ?", req.GetStatus()).Count(&count).Error; err != nil {
		return nil, err
	}
	return &contentv1.CountArticlesReply{Count: count}, nil
}

// SetArticleStatus 更新文章状态。
func (s *Server) SetArticleStatus(_ context.Context, req *contentv1.SetArticleStatusRequest) (*contentv1.Empty, error) {
	if err := s.db.Model(&models.ArticleModel{}).Where("id = ?", uint(req.GetId())).
		Update("status", req.GetStatus()).Error; err != nil {
		return nil, err
	}
	return &contentv1.Empty{}, nil
}

// SetArticleAddition 写入/更新文章扩展附录(AI 点评等)。
func (s *Server) SetArticleAddition(_ context.Context, req *contentv1.SetArticleAdditionRequest) (*contentv1.Empty, error) {
	err := s.db.Where(models.ArticleAddition{ArticleID: uint(req.GetArticleId())}).
		Assign(map[string]any{
			"ai_quality":  req.GetAiQuality(),
			"ai_abstract": req.GetAiAbstract(),
			"ai_model":    req.GetAiModel(),
		}).
		FirstOrCreate(&models.ArticleAddition{}).Error
	if err != nil {
		return nil, err
	}
	return &contentv1.Empty{}, nil
}
