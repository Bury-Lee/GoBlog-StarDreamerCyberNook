package internal

import (
	"context"

	"StarDreamerCyberNook/common"
	contentv1 "StarDreamerCyberNook/gen/content/v1"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/utils/sql"
)

// UserTopArticles 用户的置顶文章(含是否为管理员置顶)。
func (s *Server) UserTopArticles(_ context.Context, req *contentv1.UserTopArticlesRequest) (*contentv1.UserTopArticlesReply, error) {
	var tops []models.UserTopArticleModel
	if err := s.db.Preload("UserModel").Order("created_at desc").
		Find(&tops, "user_id = ?", uint(req.GetUserId())).Error; err != nil {
		return nil, err
	}
	rep := &contentv1.UserTopArticlesReply{}
	for _, t := range tops {
		rep.List = append(rep.List, &contentv1.UserTopArticle{
			ArticleId: uint64(t.ArticleID), IsAdmin: t.UserModel.Role == enum.AdminRole,
		})
	}
	return rep, nil
}

// AdminTopArticleIDs 所有管理员置顶的文章ID。
func (s *Server) AdminTopArticleIDs(_ context.Context, _ *contentv1.Empty) (*contentv1.ArticleIDsReply, error) {
	rep := &contentv1.ArticleIDsReply{}
	var adminIDs []uint
	if err := s.db.Model(models.UserModel{}).Where("role = ?", enum.AdminRole).Select("id").Scan(&adminIDs).Error; err != nil {
		return nil, err
	}
	if len(adminIDs) == 0 {
		return rep, nil
	}
	var ids []uint
	if err := s.db.Model(models.UserTopArticleModel{}).Where("user_id in ?", adminIDs).Select("article_id").Scan(&ids).Error; err != nil {
		return nil, err
	}
	rep.Ids = make([]uint64, 0, len(ids))
	for _, id := range ids {
		rep.Ids = append(rep.Ids, uint64(id))
	}
	return rep, nil
}

// GetArticlesByIDs 按 ID 批量取文章详情(保持入参顺序,含分类标题与作者昵称/头像)。
func (s *Server) GetArticlesByIDs(_ context.Context, req *contentv1.GetArticlesByIDsRequest) (*contentv1.GetArticlesByIDsReply, error) {
	ids := make([]uint, 0, len(req.GetIds()))
	for _, id := range req.GetIds() {
		ids = append(ids, uint(id))
	}
	rep := &contentv1.GetArticlesByIDsReply{}
	if len(ids) == 0 {
		return rep, nil
	}
	where := s.db.Where("id in ?", ids)
	if req.GetPublishedOnly() {
		where = where.Where("status = ?", models.StatusPublished)
	}
	list, _, _, err := common.ListQuery[models.ArticleModel](s.db, models.ArticleModel{}, common.Options{
		Where:        where,
		Preloads:     []string{"CategoryModel", "UserModel"},
		DefaultOrder: sql.ConvertSliceOrderSqlBy(s.driver, ids),
		CountCap:     common.DefaultCountCap,
	})
	if err != nil {
		return nil, err
	}
	for _, m := range list {
		item := &contentv1.ArticleDetailItem{
			Article:  toProtoArticle(m),
			NickName: m.UserModel.NickName,
			Avatar:   m.UserModel.Avatar,
		}
		if m.CategoryModel != nil {
			t := m.CategoryModel.Title
			item.CategoryTitle = &t
		}
		rep.List = append(rep.List, item)
	}
	return rep, nil
}
