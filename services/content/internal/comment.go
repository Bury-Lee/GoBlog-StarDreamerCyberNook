package internal

import (
	"context"

	"StarDreamerCyberNook/common"
	contentv1 "StarDreamerCyberNook/gen/content/v1"
	"StarDreamerCyberNook/models"
	utils_other "StarDreamerCyberNook/utils/other"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func toProtoComment(c models.CommentModel) *contentv1.Comment {
	pc := &contentv1.Comment{
		Id:         uint64(c.ID),
		Content:    c.Content,
		UserId:     uint64(c.UserID),
		ArticleId:  uint64(c.ArticleID),
		Path:       c.ParentPath,
		DiggCount:  int32(c.DiggCount),
		CreatedAt:  c.CreatedAt.UnixMilli(),
		UpdatedAt:  c.UpdatedAt.UnixMilli(),
		User: &contentv1.CommentUser{
			Id:            uint64(c.UserModel.ID),
			CreatedAt:     c.UserModel.CreatedAt.UnixMilli(),
			UpdatedAt:     c.UserModel.UpdatedAt.UnixMilli(),
			Nickname:      c.UserModel.NickName,
			Avatar:        c.UserModel.Avatar,
			Age:           int32(c.UserModel.Age),
			LikeTags:      c.UserModel.LikeTags,
			LastLoginTime: c.UserModel.LastLoginTime.UnixMilli(),
		},
	}
	if c.RootParentID != nil {
		v := uint64(*c.RootParentID)
		pc.RootParentId = &v
	}
	return pc
}

// CreateComment 创建评论(校验文章可评论、父子关系、路径),返回需要通知的接收人。
func (s *Server) CreateComment(_ context.Context, req *contentv1.CreateCommentRequest) (*contentv1.CreateCommentReply, error) {
	var article models.ArticleModel
	if err := s.db.Take(&article, "id = ? and status = ?", uint(req.GetArticleId()), models.StatusPublished).Error; err != nil {
		return nil, status.Error(codes.NotFound, "文章不存在")
	}
	if !article.OpenComment {
		return nil, status.Error(codes.FailedPrecondition, "该文章已关闭评论")
	}

	model := models.CommentModel{
		Content:   req.GetContent(),
		UserID:    uint(req.GetUserId()),
		ArticleID: uint(req.GetArticleId()),
	}

	var replyRevUserIDs []uint
	if req.GetParentId() == 0 {
		model.RootParentID = nil
		model.ParentPath = ""
	} else {
		var parent models.CommentModel
		if err := s.db.Take(&parent, "id = ? and article_id = ?", uint(req.GetParentId()), uint(req.GetArticleId())).Error; err != nil {
			return nil, status.Error(codes.NotFound, "评论不存在")
		}
		model.ParentPath = utils_other.EncodePath(parent.ParentPath, parent.ID)
		if parent.UserID != uint(req.GetUserId()) {
			replyRevUserIDs = append(replyRevUserIDs, parent.UserID)
		}
		if parent.RootParentID == nil {
			model.RootParentID = &parent.ID
		} else {
			model.RootParentID = parent.RootParentID
			var root models.CommentModel
			if err := s.db.Take(&root, "id = ? and article_id = ?", *parent.RootParentID, uint(req.GetArticleId())).Error; err == nil {
				if root.UserID != parent.UserID && root.UserID != uint(req.GetUserId()) {
					replyRevUserIDs = append(replyRevUserIDs, root.UserID)
				}
			}
		}
	}

	if err := s.db.Create(&model).Error; err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(replyRevUserIDs))
	for _, id := range replyRevUserIDs {
		ids = append(ids, uint64(id))
	}
	return &contentv1.CreateCommentReply{
		Id:              uint64(model.ID),
		ReplyRevUserIds: ids,
		ArticleUserId:   uint64(article.UserID),
	}, nil
}

func commentListOptions(page, limit int, key, order string, endID uint, where *gorm.DB) common.Options {
	return common.Options{
		PageInfo:      common.PageInfo{Page: page, Limit: limit, Key: key, Order: order, EndId: endID},
		Preloads:      []string{"UserModel"},
		Where:         where,
		DefaultOrder:  "digg_count desc",
		AllowedOrders: []string{"id", "created_at", "digg_count"},
		CountCap:      common.DefaultCountCap,
	}
}

// ListComments 一级评论列表。
func (s *Server) ListComments(_ context.Context, req *contentv1.ListCommentsRequest) (*contentv1.ListCommentsReply, error) {
	var article models.ArticleModel
	if err := s.db.Take(&article, "id = ? and status = ?", uint(req.GetArticleId()), models.StatusPublished).Error; err != nil {
		return nil, status.Error(codes.NotFound, "文章不存在")
	}
	opts := commentListOptions(int(req.GetPage()), int(req.GetLimit()), req.GetKey(), req.GetOrder(), uint(req.GetEndId()),
		s.db.Where("article_id = ? and root_parent_id is null", uint(req.GetArticleId())))
	list, count, capped, err := common.ListQuery(s.db, models.CommentModel{}, opts)
	if err != nil {
		return nil, err
	}
	rep := &contentv1.ListCommentsReply{Count: int64(count), Capped: capped}
	for _, c := range list {
		rep.List = append(rep.List, toProtoComment(c))
	}
	return rep, nil
}

// ListChildComments 某根评论下的子评论列表。
func (s *Server) ListChildComments(_ context.Context, req *contentv1.ListCommentsRequest) (*contentv1.ListCommentsReply, error) {
	var comments models.CommentModel
	if err := s.db.Take(&comments, "id = ?", uint(req.GetRootId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "评论不存在")
	}
	var article models.ArticleModel
	if err := s.db.Take(&article, "id = ? and status = ?", comments.ArticleID, models.StatusPublished).Error; err != nil {
		return nil, status.Error(codes.NotFound, "文章不存在")
	}
	path := utils_other.EncodePath(comments.ParentPath, comments.ID)
	opts := commentListOptions(int(req.GetPage()), int(req.GetLimit()), req.GetKey(), req.GetOrder(), uint(req.GetEndId()),
		s.db.Where("article_id = ? and (parent_path = ? or parent_path like ?)",
			comments.ArticleID, path, path+"/%"))
	list, count, capped, err := common.ListQuery(s.db, models.CommentModel{}, opts)
	if err != nil {
		return nil, err
	}
	rep := &contentv1.ListCommentsReply{Count: int64(count), Capped: capped}
	for _, c := range list {
		rep.List = append(rep.List, toProtoComment(c))
	}
	return rep, nil
}

// DeleteComment 删除评论(一级连带子评论),返回文章ID与评论数减量。
func (s *Server) DeleteComment(_ context.Context, req *contentv1.DeleteCommentRequest) (*contentv1.DeleteCommentReply, error) {
	var comment models.CommentModel
	if err := s.db.Take(&comment, "id = ?", uint(req.GetId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "评论不存在")
	}
	var article models.ArticleModel
	if err := s.db.Take(&article, "id = ? and status = ?", comment.ArticleID, models.StatusPublished).Error; err != nil {
		return nil, status.Error(codes.NotFound, "文章不存在")
	}
	if uint(req.GetUserId()) != comment.UserID && !req.GetIsAdmin() {
		return nil, status.Error(codes.PermissionDenied, "没有权限删除评论")
	}

	var delta int32
	if comment.RootParentID == nil {
		var childCount int64
		if err := s.db.Model(&models.CommentModel{}).Where("root_parent_id = ?", comment.ID).Count(&childCount).Error; err != nil {
			return nil, err
		}
		if err := s.db.Delete(&models.CommentModel{}, "root_parent_id = ?", comment.ID).Error; err != nil {
			return nil, err
		}
		if err := s.db.Delete(&comment).Error; err != nil {
			return nil, err
		}
		delta = int32(childCount + 1)
	} else {
		if err := s.db.Delete(&comment).Error; err != nil {
			return nil, err
		}
		delta = 1
	}
	return &contentv1.DeleteCommentReply{ArticleId: uint64(comment.ArticleID), Delta: delta, Ok: true}, nil
}

// ToggleCommentDigg 切换评论点赞,返回最新状态与基准点赞数。
func (s *Server) ToggleCommentDigg(_ context.Context, req *contentv1.ToggleCommentDiggRequest) (*contentv1.ToggleCommentDiggReply, error) {
	var comment models.CommentModel
	if err := s.db.Take(&comment, "id = ?", uint(req.GetCommentId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "评论不存在")
	}

	digged := false
	var d models.CommentDiggModel
	err := s.db.Take(&d, "user_id = ? and comment_id = ?", uint(req.GetUserId()), uint(req.GetCommentId())).Error
	if err == gorm.ErrRecordNotFound {
		if err := s.db.Create(&models.CommentDiggModel{UserID: uint(req.GetUserId()), CommentID: uint(req.GetCommentId())}).Error; err != nil {
			return nil, err
		}
		digged = true
	} else {
		if err := s.db.Delete(&models.CommentDiggModel{}, "user_id = ? and comment_id = ?", uint(req.GetUserId()), uint(req.GetCommentId())).Error; err != nil {
			return nil, err
		}
	}
	return &contentv1.ToggleCommentDiggReply{Digged: digged, BaseDiggCount: int32(comment.DiggCount)}, nil
}

// CommentDiggIDs 查询用户在一组评论上的点赞状态。
func (s *Server) CommentDiggIDs(_ context.Context, req *contentv1.CommentDiggIDsRequest) (*contentv1.CommentDiggIDsReply, error) {
	ids := make([]uint, 0, len(req.GetCommentIds()))
	for _, id := range req.GetCommentIds() {
		ids = append(ids, uint(id))
	}
	digged := make([]uint64, 0)
	if len(ids) > 0 {
		var tmp []uint
		if err := s.db.Model(&models.CommentDiggModel{}).
			Where("user_id = ? and comment_id in ?", uint(req.GetUserId()), ids).
			Pluck("comment_id", &tmp).Error; err != nil {
			return nil, err
		}
		for _, id := range tmp {
			digged = append(digged, uint64(id))
		}
	}
	return &contentv1.CommentDiggIDsReply{DiggedIds: digged}, nil
}
