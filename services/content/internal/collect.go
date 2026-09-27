package internal

import (
	"context"
	"encoding/json"
	"errors"

	"StarDreamerCyberNook/common"
	contentv1 "StarDreamerCyberNook/gen/content/v1"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/utils/sql"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func toProtoFolder(c models.CollectModel) *contentv1.CollectFolder {
	return &contentv1.CollectFolder{
		Id: uint64(c.ID), UserId: uint64(c.UserID), Title: c.Title, Abstract: c.Abstract, Cover: c.Cover,
		IsPublic: c.IsPublic, IsDefault: c.IsDefault,
		CreatedAt: c.CreatedAt.UnixMilli(), UpdatedAt: c.UpdatedAt.UnixMilli(),
	}
}

// ToggleArticleCollect 收藏/取消收藏/移动收藏夹;返回动作(created|removed|moved)。
func (s *Server) ToggleArticleCollect(_ context.Context, req *contentv1.ToggleCollectRequest) (*contentv1.ToggleCollectReply, error) {
	userID, articleID, collectID := uint(req.GetUserId()), uint(req.GetArticleId()), uint(req.GetCollectId())

	var article models.ArticleModel
	if err := s.db.Take(&article, "status = ? and id = ?", models.StatusPublished, articleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "文章不存在")
		}
		return nil, err
	}

	// 目标收藏夹:0 用默认(不存在则建),否则须属于当前用户
	var folder models.CollectModel
	if collectID == 0 {
		err := s.db.Take(&folder, "user_id = ? and is_default = ?", userID, true).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			folder = models.CollectModel{Title: "默认收藏夹", UserID: userID, IsDefault: true}
			if err = s.db.Create(&folder).Error; err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
	} else {
		if err := s.db.Take(&folder, "id = ? and user_id = ?", collectID, userID).Error; err != nil {
			return nil, status.Error(codes.FailedPrecondition, "收藏夹不存在")
		}
	}

	var rec models.UserArticleCollectModel
	err := s.db.Take(&rec, "user_id = ? and article_id = ?", userID, articleID).Error
	if err == nil {
		if rec.CollectID == folder.ID { // 已收藏且同夹 → 取消
			if err := s.db.Where("user_id = ? and article_id = ?", userID, articleID).Delete(&rec).Error; err != nil {
				return nil, err
			}
			return &contentv1.ToggleCollectReply{Action: "removed"}, nil
		}
		// 已收藏但在别的夹 → 移动(总数不变)
		if err := s.db.Model(&models.UserArticleCollectModel{}).
			Where("user_id = ? and article_id = ?", userID, articleID).
			Update("collect_id", folder.ID).Error; err != nil {
			return nil, err
		}
		return &contentv1.ToggleCollectReply{Action: "moved"}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err := s.db.Create(&models.UserArticleCollectModel{
		UserID: userID, ArticleID: articleID, CollectID: folder.ID,
	}).Error; err != nil {
		return nil, err
	}
	return &contentv1.ToggleCollectReply{Action: "created"}, nil
}

// CreateCollectFolder 创建收藏夹。
func (s *Server) CreateCollectFolder(_ context.Context, req *contentv1.CreateFolderRequest) (*contentv1.Empty, error) {
	if err := s.db.Create(&models.CollectModel{
		Title: req.GetTitle(), UserID: uint(req.GetUserId()), Abstract: req.GetAbstract(), Cover: req.GetCover(),
	}).Error; err != nil {
		return nil, err
	}
	return &contentv1.Empty{}, nil
}

// UpdateCollectFolder 增量更新收藏夹(patch 为 JSON 列补丁)。
func (s *Server) UpdateCollectFolder(_ context.Context, req *contentv1.UpdateFolderRequest) (*contentv1.Empty, error) {
	var patch map[string]any
	if err := json.Unmarshal([]byte(req.GetPatch()), &patch); err != nil {
		return nil, status.Error(codes.InvalidArgument, "bad patch")
	}
	var m models.CollectModel
	if err := s.db.Take(&m, "user_id = ? and id = ?", uint(req.GetUserId()), uint(req.GetId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "收藏夹不存在")
	}
	if len(patch) > 0 {
		if err := s.db.Model(&m).Updates(patch).Error; err != nil {
			return nil, err
		}
	}
	return &contentv1.Empty{}, nil
}

// RemoveCollectFolders 删除收藏夹并级联删除收藏记录,返回各文章收藏数的回退增量。
func (s *Server) RemoveCollectFolders(_ context.Context, req *contentv1.RemoveFoldersRequest) (*contentv1.RemoveFoldersReply, error) {
	rep := &contentv1.RemoveFoldersReply{ArticleDeltas: map[uint64]int32{}}
	ids := make([]uint, 0, len(req.GetIds()))
	for _, id := range req.GetIds() {
		ids = append(ids, uint(id))
	}
	if len(ids) == 0 {
		return rep, nil
	}

	query := s.db.Model(&models.CollectModel{}).Where("id IN ? AND is_default = ?", ids, false)
	if !req.GetIsAdmin() {
		query = query.Where("user_id = ?", uint(req.GetUserId()))
	}
	var folderIDs []uint
	if err := query.Pluck("id", &folderIDs).Error; err != nil {
		return nil, err
	}
	if len(folderIDs) == 0 {
		return rep, nil
	}

	var records []models.UserArticleCollectModel
	if err := s.db.Where("collect_id IN ?", folderIDs).Find(&records).Error; err != nil {
		return nil, err
	}
	deltas := map[uint]int{}
	for _, r := range records {
		deltas[r.ArticleID]--
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if len(records) > 0 {
			if err := tx.Session(&gorm.Session{SkipHooks: true}).
				Where("collect_id IN ?", folderIDs).
				Delete(&models.UserArticleCollectModel{}).Error; err != nil {
				return err
			}
		}
		return tx.Where("id IN ?", folderIDs).Delete(&models.CollectModel{}).Error
	})
	if err != nil {
		return nil, err
	}
	for aid, d := range deltas {
		rep.ArticleDeltas[uint64(aid)] = int32(d)
	}
	rep.Deleted = int32(len(folderIDs))
	return rep, nil
}

// ListCollectFolders 收藏夹列表(include_private=false 时仅公开)。
func (s *Server) ListCollectFolders(_ context.Context, req *contentv1.ListFoldersRequest) (*contentv1.ListFoldersReply, error) {
	where := s.db.Where("user_id = ?", uint(req.GetUserId()))
	if !req.GetIncludePrivate() {
		where = where.Where("is_public = ?", true)
	}
	opts := common.Options{
		PageInfo:      common.PageInfo{Page: int(req.GetPage()), Limit: int(req.GetLimit()), Order: req.GetOrder(), EndId: uint(req.GetEndId())},
		Where:         where,
		Likes:         []string{"title", "abstract"},
		DefaultOrder:  "created_at desc",
		AllowedOrders: []string{"id", "created_at"},
		CountCap:      common.DefaultCountCap,
	}
	list, count, capped, err := common.ListQuery[models.CollectModel](s.db, models.CollectModel{}, opts)
	if err != nil {
		return nil, err
	}
	rep := &contentv1.ListFoldersReply{Count: int64(count), Capped: capped}
	for _, m := range list {
		rep.List = append(rep.List, toProtoFolder(m))
	}
	return rep, nil
}

// GetCollectFolder 收藏夹详情 + 内文章数。
func (s *Server) GetCollectFolder(_ context.Context, req *contentv1.GetFolderRequest) (*contentv1.CollectFolderReply, error) {
	var m models.CollectModel
	if err := s.db.Take(&m, uint(req.GetId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "收藏夹不存在")
	}
	var count int64
	s.db.Model(&models.UserArticleCollectModel{}).Where("collect_id = ?", m.ID).Count(&count)
	return &contentv1.CollectFolderReply{Folder: toProtoFolder(m), ArticleCount: count}, nil
}

// ListCollectArticles 收藏夹内文章(按收藏时间倒序;仅已发布)。
func (s *Server) ListCollectArticles(_ context.Context, req *contentv1.ListCollectArticlesRequest) (*contentv1.ArticleListReply, error) {
	var records []models.UserArticleCollectModel
	if err := s.db.Where("collect_id = ?", uint(req.GetFolderId())).
		Order("created_at desc").Find(&records).Error; err != nil {
		return nil, err
	}
	rep := &contentv1.ArticleListReply{}
	if len(records) == 0 {
		return rep, nil
	}
	ids := make([]uint, 0, len(records))
	for _, r := range records {
		ids = append(ids, r.ArticleID)
	}
	opts := common.Options{
		PageInfo:      common.PageInfo{Page: int(req.GetPage()), Limit: int(req.GetLimit()), Order: req.GetOrder(), EndId: uint(req.GetEndId())},
		Where:         s.db.Where("id in ? and status = ?", ids, models.StatusPublished),
		Likes:         []string{"title", "abstract"},
		DefaultOrder:  sql.ConvertSliceOrderSqlBy(s.driver, ids),
		AllowedOrders: []string{"id", "created_at", "look_count", "digg_count", "comment_count", "collect_count"},
		CountCap:      common.DefaultCountCap,
	}
	list, count, capped, err := common.ListQuery[models.ArticleModel](s.db, models.ArticleModel{}, opts)
	if err != nil {
		return nil, err
	}
	rep.Count = int64(count)
	rep.Capped = capped
	for _, m := range list {
		rep.List = append(rep.List, toProtoArticle(m))
	}
	return rep, nil
}
