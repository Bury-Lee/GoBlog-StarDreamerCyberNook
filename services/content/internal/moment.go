package internal

import (
	"context"
	"encoding/json"
	"errors"

	"StarDreamerCyberNook/common"
	contentv1 "StarDreamerCyberNook/gen/content/v1"
	"StarDreamerCyberNook/models"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func reduceUser(u models.UserModel) *contentv1.CommentUser {
	return &contentv1.CommentUser{
		Id:            uint64(u.ID),
		CreatedAt:     u.CreatedAt.UnixMilli(),
		UpdatedAt:     u.UpdatedAt.UnixMilli(),
		Nickname:      u.NickName,
		Avatar:        u.Avatar,
		Age:           int32(u.Age),
		LikeTags:      u.LikeTags,
		LastLoginTime: u.LastLoginTime.UnixMilli(),
	}
}

func toProtoMoment(m models.MomentModel) *contentv1.Moment {
	pm := &contentv1.Moment{
		Id:           uint64(m.ID),
		UserId:       uint64(m.UserID),
		User:         reduceUser(m.UserModel),
		Type:         int32(m.Type),
		Visibility:   int32(m.Visibility),
		Content:      m.Content,
		Images:       m.Images,
		LikeCount:    int32(m.LikeCount),
		CommentCount: int32(m.CommentCount),
		RepostCount:  int32(m.RepostCount),
		Status:       int32(m.Status),
		CreatedAt:    m.CreatedAt.UnixMilli(),
		UpdatedAt:    m.UpdatedAt.UnixMilli(),
	}
	if m.RepostFromID != nil {
		v := uint64(*m.RepostFromID)
		pm.RepostFromId = &v
	}
	if m.RepostFrom != nil {
		pm.RepostFrom = &contentv1.Moment{
			Id:           uint64(m.RepostFrom.ID),
			UserId:       uint64(m.RepostFrom.UserID),
			User:         reduceUser(m.RepostFrom.UserModel),
			Type:         int32(m.RepostFrom.Type),
			Visibility:   int32(m.RepostFrom.Visibility),
			Content:      m.RepostFrom.Content,
			Images:       m.RepostFrom.Images,
			LikeCount:    int32(m.RepostFrom.LikeCount),
			CommentCount: int32(m.RepostFrom.CommentCount),
			RepostCount:  int32(m.RepostFrom.RepostCount),
			Status:       int32(m.RepostFrom.Status),
			CreatedAt:    m.RepostFrom.CreatedAt.UnixMilli(),
			UpdatedAt:    m.RepostFrom.UpdatedAt.UnixMilli(),
		}
	}
	return pm
}

func (s *Server) canViewMoment(m models.MomentModel, viewerID uint, isAdmin bool) bool {
	if isAdmin || (viewerID != 0 && viewerID == m.UserID) {
		return true
	}
	if m.Status != models.StatusPublished {
		return false
	}
	switch m.Visibility {
	case models.MomentVisibilityPublic:
		return true
	case models.MomentVisibilityFriends:
		if viewerID == 0 {
			return false
		}
		var rel models.UserFollowModel
		return s.db.Take(&rel, "user_id = ? and focus_user_id = ? and friend = ?", viewerID, m.UserID, true).Error == nil
	default:
		return false
	}
}

func (s *Server) momentWithPreload(id uint) (models.MomentModel, error) {
	var m models.MomentModel
	err := s.db.Preload("UserModel").Preload("RepostFrom.UserModel").Take(&m, id).Error
	return m, err
}

// CreateMoment 发布动态/日记。
func (s *Server) CreateMoment(_ context.Context, req *contentv1.CreateMomentRequest) (*contentv1.MomentReply, error) {
	model := models.MomentModel{
		UserID:     uint(req.GetUserId()),
		Type:       models.MomentType(req.GetType()),
		Visibility: models.MomentVisibility(req.GetVisibility()),
		Content:    req.GetContent(),
		Images:     req.GetImages(),
		Status:     models.Status(req.GetStatus()),
	}
	if err := s.db.Create(&model).Error; err != nil {
		return nil, err
	}
	if m, err := s.momentWithPreload(model.ID); err == nil {
		model = m
	}
	return &contentv1.MomentReply{Moment: toProtoMoment(model)}, nil
}

// ListMoments 动态列表(按隐私规则过滤)。
func (s *Server) ListMoments(_ context.Context, req *contentv1.ListMomentsRequest) (*contentv1.ListMomentsReply, error) {
	where := s.db
	if req.GetUserId() != 0 {
		where = where.Where("user_id = ?", uint(req.GetUserId()))
		if uint(req.GetUserId()) != uint(req.GetViewerId()) && !req.GetIsAdmin() {
			var rel models.UserFollowModel
			if req.GetViewerId() != 0 && s.db.Take(&rel, "user_id = ? and focus_user_id = ? and friend = ?",
				uint(req.GetViewerId()), uint(req.GetUserId()), true).Error == nil {
				where = where.Where("status = ? and visibility in ?", models.StatusPublished,
					[]models.MomentVisibility{models.MomentVisibilityPublic, models.MomentVisibilityFriends})
			} else {
				where = where.Where("status = ? and visibility = ?", models.StatusPublished, models.MomentVisibilityPublic)
			}
		}
	} else {
		where = where.Where("status = ? and visibility = ?", models.StatusPublished, models.MomentVisibilityPublic)
	}
	if req.Type != nil {
		where = where.Where("type = ?", *req.Type)
	}

	opts := common.Options{
		PageInfo:      common.PageInfo{Page: int(req.GetPage()), Limit: int(req.GetLimit()), Key: req.GetKey(), Order: req.GetOrder(), EndId: uint(req.GetEndId())},
		Preloads:      []string{"UserModel", "RepostFrom.UserModel"},
		Where:         where,
		DefaultOrder:  "created_at desc",
		AllowedOrders: []string{"id", "created_at", "like_count", "comment_count"},
		CountCap:      common.DefaultCountCap,
	}
	list, count, capped, err := common.ListQuery(s.db, models.MomentModel{}, opts)
	if err != nil {
		return nil, err
	}
	rep := &contentv1.ListMomentsReply{Count: int64(count), Capped: capped}
	for _, m := range list {
		rep.List = append(rep.List, toProtoMoment(m))
	}
	return rep, nil
}

// GetMoment 动态详情。
func (s *Server) GetMoment(_ context.Context, req *contentv1.GetMomentRequest) (*contentv1.MomentReply, error) {
	m, err := s.momentWithPreload(uint(req.GetId()))
	if err != nil {
		return nil, status.Error(codes.NotFound, "动态不存在")
	}
	if !s.canViewMoment(m, uint(req.GetViewerId()), req.GetIsAdmin()) {
		return nil, status.Error(codes.NotFound, "动态不存在")
	}
	return &contentv1.MomentReply{Moment: toProtoMoment(m)}, nil
}

// UpdateMoment 更新本人动态。
func (s *Server) UpdateMoment(_ context.Context, req *contentv1.UpdateMomentRequest) (*contentv1.UpdateMomentReply, error) {
	var m models.MomentModel
	if err := s.db.Take(&m, uint(req.GetId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "动态不存在")
	}
	if uint(req.GetUserId()) != m.UserID {
		return nil, status.Error(codes.PermissionDenied, "没有权限修改该动态")
	}
	img, _ := json.Marshal(req.GetImages())
	if err := s.db.Model(&m).Updates(map[string]any{
		"content":    req.GetContent(),
		"images":     string(img),
		"type":       req.GetType(),
		"visibility": req.GetVisibility(),
		"status":     req.GetStatus(),
	}).Error; err != nil {
		return nil, err
	}
	return &contentv1.UpdateMomentReply{}, nil
}

// RemoveMoment 删除动态(本人/管理员),级联删除评论与点赞。
func (s *Server) RemoveMoment(_ context.Context, req *contentv1.RemoveMomentRequest) (*contentv1.RemoveMomentReply, error) {
	var m models.MomentModel
	if err := s.db.Take(&m, uint(req.GetId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "动态不存在")
	}
	if uint(req.GetUserId()) != m.UserID && !req.GetIsAdmin() {
		return nil, status.Error(codes.PermissionDenied, "没有权限删除该动态")
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("moment_comment_id in (?)",
			tx.Model(&models.MomentCommentModel{}).Select("id").Where("moment_id = ?", m.ID)).
			Delete(&models.MomentCommentDiggModel{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.MomentCommentModel{}, "moment_id = ?", m.ID).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.MomentDiggModel{}, "moment_id = ?", m.ID).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.MomentModel{}, "id = ?", m.ID).Error; err != nil {
			return err
		}
		if m.RepostFromID != nil {
			if err := tx.Model(&models.MomentModel{}).Where("id = ? and repost_count > 0", *m.RepostFromID).
				UpdateColumn("repost_count", gorm.Expr("repost_count - 1")).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &contentv1.RemoveMomentReply{}, nil
}

// MomentInteraction 当前用户对该动态的点赞状态。
func (s *Server) MomentInteraction(_ context.Context, req *contentv1.MomentInteractionRequest) (*contentv1.MomentInteractionReply, error) {
	digged := false
	if req.GetViewerId() != 0 {
		err := s.db.Take(&models.MomentDiggModel{}, "user_id = ? and moment_id = ?", uint(req.GetViewerId()), uint(req.GetId())).Error
		digged = err == nil
	}
	return &contentv1.MomentInteractionReply{Digged: digged}, nil
}

// ToggleMomentDigg 点赞/取消点赞(事务内维护计数)。
func (s *Server) ToggleMomentDigg(_ context.Context, req *contentv1.ToggleMomentDiggRequest) (*contentv1.ToggleMomentDiggReply, error) {
	var m models.MomentModel
	if err := s.db.Take(&m, uint(req.GetId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "动态不存在")
	}
	digged := false
	err := s.db.Transaction(func(tx *gorm.DB) error {
		takeErr := tx.Take(&models.MomentDiggModel{}, "user_id = ? and moment_id = ?", uint(req.GetUserId()), uint(req.GetId())).Error
		switch {
		case errors.Is(takeErr, gorm.ErrRecordNotFound):
			if e := tx.Create(&models.MomentDiggModel{UserID: uint(req.GetUserId()), MomentID: uint(req.GetId())}).Error; e != nil {
				return e
			}
			if e := tx.Model(&models.MomentModel{}).Where("id = ?", uint(req.GetId())).
				UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error; e != nil {
				return e
			}
			digged = true
		case takeErr != nil:
			return takeErr
		default:
			res := tx.Where("user_id = ? and moment_id = ?", uint(req.GetUserId()), uint(req.GetId())).Delete(&models.MomentDiggModel{})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected > 0 {
				if e := tx.Model(&models.MomentModel{}).Where("id = ? and like_count > 0", uint(req.GetId())).
					UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error; e != nil {
					return e
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var likeCount int
	s.db.Model(&models.MomentModel{}).Where("id = ?", uint(req.GetId())).Select("like_count").Scan(&likeCount)
	return &contentv1.ToggleMomentDiggReply{Digged: digged, LikeCount: int32(likeCount)}, nil
}

// RepostMoment 转发动态。
func (s *Server) RepostMoment(_ context.Context, req *contentv1.RepostMomentRequest) (*contentv1.MomentReply, error) {
	source, err := s.momentWithPreload(uint(req.GetId()))
	if err != nil {
		return nil, status.Error(codes.NotFound, "动态不存在")
	}
	if !s.canViewMoment(source, uint(req.GetViewerId()), req.GetIsAdmin()) {
		return nil, status.Error(codes.NotFound, "动态不存在")
	}

	model := models.MomentModel{
		UserID:       uint(req.GetUserId()),
		Type:         models.MomentTypeMoment,
		Visibility:   models.MomentVisibility(req.GetVisibility()),
		Content:      req.GetContent(),
		RepostFromID: &source.ID,
		Status:       models.StatusPublished,
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if e := tx.Create(&model).Error; e != nil {
			return e
		}
		return tx.Model(&models.MomentModel{}).Where("id = ?", source.ID).
			UpdateColumn("repost_count", gorm.Expr("repost_count + 1")).Error
	}); err != nil {
		return nil, err
	}

	model.RepostFrom = &source
	if m, e := s.momentWithPreload(model.ID); e == nil {
		model = m
	}
	return &contentv1.MomentReply{Moment: toProtoMoment(model)}, nil
}
