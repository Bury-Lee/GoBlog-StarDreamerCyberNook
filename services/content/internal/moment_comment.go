package internal

import (
	"context"
	"errors"

	"StarDreamerCyberNook/common"
	contentv1 "StarDreamerCyberNook/gen/content/v1"
	"StarDreamerCyberNook/models"
	utils_other "StarDreamerCyberNook/utils/other"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func toProtoMomentComment(c models.MomentCommentModel) *contentv1.MomentComment {
	pc := &contentv1.MomentComment{
		Id:        uint64(c.ID),
		MomentId:  uint64(c.MomentID),
		UserId:    uint64(c.UserID),
		User:      reduceUser(c.UserModel),
		Content:   c.Content,
		Path:      c.ParentPath,
		DiggCount: int32(c.DiggCount),
		CreatedAt: c.CreatedAt.UnixMilli(),
		UpdatedAt: c.UpdatedAt.UnixMilli(),
	}
	if c.RootParentID != nil {
		v := uint64(*c.RootParentID)
		pc.RootParentId = &v
	}
	return pc
}

// CreateMomentComment 发表动态评论/回复(校验可见性,事务内 +1 评论数)。
func (s *Server) CreateMomentComment(_ context.Context, req *contentv1.CreateMomentCommentRequest) (*contentv1.MomentCommentReply, error) {
	var m models.MomentModel
	if err := s.db.Take(&m, uint(req.GetMomentId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "动态不存在")
	}
	if !s.canViewMoment(m, uint(req.GetViewerId()), req.GetIsAdmin()) {
		return nil, status.Error(codes.NotFound, "动态不存在")
	}

	model := models.MomentCommentModel{
		MomentID: uint(req.GetMomentId()),
		UserID:   uint(req.GetUserId()),
		Content:  req.GetContent(),
	}
	if req.GetParentId() != 0 {
		var parent models.MomentCommentModel
		if err := s.db.Take(&parent, "id = ? and moment_id = ?", uint(req.GetParentId()), uint(req.GetMomentId())).Error; err != nil {
			return nil, status.Error(codes.NotFound, "评论不存在")
		}
		model.ParentPath = utils_other.EncodePath(parent.ParentPath, parent.ID)
		if parent.RootParentID == nil {
			model.RootParentID = &parent.ID
		} else {
			model.RootParentID = parent.RootParentID
		}
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if e := tx.Create(&model).Error; e != nil {
			return e
		}
		return tx.Model(&models.MomentModel{}).Where("id = ?", uint(req.GetMomentId())).
			UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
	}); err != nil {
		return nil, err
	}
	return &contentv1.MomentCommentReply{Comment: toProtoMomentComment(model)}, nil
}

func momentCommentOptions(page, limit int, key, order string, endID uint, where *gorm.DB) common.Options {
	return common.Options{
		PageInfo:      common.PageInfo{Page: page, Limit: limit, Key: key, Order: order, EndId: endID},
		Preloads:      []string{"UserModel"},
		Where:         where,
		DefaultOrder:  "created_at desc",
		AllowedOrders: []string{"id", "created_at", "digg_count"},
		CountCap:      common.DefaultCountCap,
	}
}

// ListMomentComments 一级评论列表。
func (s *Server) ListMomentComments(_ context.Context, req *contentv1.ListMomentCommentsRequest) (*contentv1.ListMomentCommentsReply, error) {
	var m models.MomentModel
	if err := s.db.Take(&m, uint(req.GetMomentId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "动态不存在")
	}
	if !s.canViewMoment(m, uint(req.GetViewerId()), req.GetIsAdmin()) {
		return nil, status.Error(codes.NotFound, "动态不存在")
	}
	opts := momentCommentOptions(int(req.GetPage()), int(req.GetLimit()), req.GetKey(), req.GetOrder(), uint(req.GetEndId()),
		s.db.Where("moment_id = ? and root_parent_id is null", uint(req.GetMomentId())))
	list, count, capped, err := common.ListQuery(s.db, models.MomentCommentModel{}, opts)
	if err != nil {
		return nil, err
	}
	rep := &contentv1.ListMomentCommentsReply{Count: int64(count), Capped: capped}
	for _, c := range list {
		rep.List = append(rep.List, toProtoMomentComment(c))
	}
	return rep, nil
}

// ListMomentChildComments 某根评论下的子评论列表。
func (s *Server) ListMomentChildComments(_ context.Context, req *contentv1.ListMomentChildCommentsRequest) (*contentv1.ListMomentCommentsReply, error) {
	var root models.MomentCommentModel
	if err := s.db.Take(&root, uint(req.GetRootId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "评论不存在")
	}
	path := utils_other.EncodePath(root.ParentPath, root.ID)
	opts := momentCommentOptions(int(req.GetPage()), int(req.GetLimit()), req.GetKey(), req.GetOrder(), uint(req.GetEndId()),
		s.db.Where("moment_id = ? and (parent_path = ? or parent_path like ?)", root.MomentID, path, path+"/%"))
	list, count, capped, err := common.ListQuery(s.db, models.MomentCommentModel{}, opts)
	if err != nil {
		return nil, err
	}
	rep := &contentv1.ListMomentCommentsReply{Count: int64(count), Capped: capped}
	for _, c := range list {
		rep.List = append(rep.List, toProtoMomentComment(c))
	}
	return rep, nil
}

// DeleteMomentComment 删除动态评论(本人/管理员),一级连带子评论与其点赞。
func (s *Server) DeleteMomentComment(_ context.Context, req *contentv1.DeleteMomentCommentRequest) (*contentv1.DeleteMomentCommentReply, error) {
	var comment models.MomentCommentModel
	if err := s.db.Take(&comment, uint(req.GetId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "评论不存在")
	}
	if uint(req.GetUserId()) != comment.UserID && !req.GetIsAdmin() {
		return nil, status.Error(codes.PermissionDenied, "没有权限删除该评论")
	}

	if comment.RootParentID == nil {
		var childCount int64
		if err := s.db.Model(&models.MomentCommentModel{}).Where("root_parent_id = ?", comment.ID).Count(&childCount).Error; err != nil {
			return nil, err
		}
		if err := s.db.Transaction(func(tx *gorm.DB) error {
			childIDs := tx.Model(&models.MomentCommentModel{}).Select("id").Where("root_parent_id = ?", comment.ID)
			if e := tx.Where("moment_comment_id = ? or moment_comment_id in (?)", comment.ID, childIDs).
				Delete(&models.MomentCommentDiggModel{}).Error; e != nil {
				return e
			}
			if e := tx.Delete(&models.MomentCommentModel{}, "root_parent_id = ?", comment.ID).Error; e != nil {
				return e
			}
			if e := tx.Delete(&comment).Error; e != nil {
				return e
			}
			return tx.Model(&models.MomentModel{}).Where("id = ? and comment_count >= ?", comment.MomentID, childCount+1).
				UpdateColumn("comment_count", gorm.Expr("comment_count - ?", childCount+1)).Error
		}); err != nil {
			return nil, err
		}
	} else {
		if err := s.db.Transaction(func(tx *gorm.DB) error {
			res := tx.Delete(&comment)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return nil
			}
			if e := tx.Where("moment_comment_id = ?", comment.ID).Delete(&models.MomentCommentDiggModel{}).Error; e != nil {
				return e
			}
			return tx.Model(&models.MomentModel{}).Where("id = ? and comment_count > 0", comment.MomentID).
				UpdateColumn("comment_count", gorm.Expr("comment_count - 1")).Error
		}); err != nil {
			return nil, err
		}
	}
	return &contentv1.DeleteMomentCommentReply{}, nil
}

// ToggleMomentCommentDigg 动态评论点赞/取消(事务内维护计数)。
func (s *Server) ToggleMomentCommentDigg(_ context.Context, req *contentv1.ToggleMomentCommentDiggRequest) (*contentv1.ToggleMomentCommentDiggReply, error) {
	var comment models.MomentCommentModel
	if err := s.db.Take(&comment, uint(req.GetId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "评论不存在")
	}

	digged := false
	err := s.db.Transaction(func(tx *gorm.DB) error {
		takeErr := tx.Take(&models.MomentCommentDiggModel{}, "user_id = ? and moment_comment_id = ?", uint(req.GetUserId()), uint(req.GetId())).Error
		switch {
		case errors.Is(takeErr, gorm.ErrRecordNotFound):
			if e := tx.Create(&models.MomentCommentDiggModel{UserID: uint(req.GetUserId()), MomentCommentID: uint(req.GetId())}).Error; e != nil {
				return e
			}
			if e := tx.Model(&models.MomentCommentModel{}).Where("id = ?", uint(req.GetId())).
				UpdateColumn("digg_count", gorm.Expr("digg_count + 1")).Error; e != nil {
				return e
			}
			digged = true
		case takeErr != nil:
			return takeErr
		default:
			res := tx.Where("user_id = ? and moment_comment_id = ?", uint(req.GetUserId()), uint(req.GetId())).Delete(&models.MomentCommentDiggModel{})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected > 0 {
				if e := tx.Model(&models.MomentCommentModel{}).Where("id = ? and digg_count > 0", uint(req.GetId())).
					UpdateColumn("digg_count", gorm.Expr("digg_count - 1")).Error; e != nil {
					return e
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var diggCount int
	s.db.Model(&models.MomentCommentModel{}).Where("id = ?", uint(req.GetId())).Select("digg_count").Scan(&diggCount)
	return &contentv1.ToggleMomentCommentDiggReply{Digged: digged, DiggCount: int32(diggCount)}, nil
}
