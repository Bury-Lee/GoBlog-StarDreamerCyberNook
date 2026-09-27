package content_service

import (
	"context"
	"time"

	contentv1 "StarDreamerCyberNook/gen/content/v1"
	"StarDreamerCyberNook/models"
)

func fromProtoMoment(pm *contentv1.Moment) models.MomentModel {
	m := models.MomentModel{
		Model:        models.Model{ID: uint(pm.GetId()), CreatedAt: time.UnixMilli(pm.GetCreatedAt()), UpdatedAt: time.UnixMilli(pm.GetUpdatedAt())},
		UserID:       uint(pm.GetUserId()),
		UserModel:    fromProtoCommentUser(pm.GetUser()),
		Type:         models.MomentType(pm.GetType()),
		Visibility:   models.MomentVisibility(pm.GetVisibility()),
		Content:      pm.GetContent(),
		Images:       pm.GetImages(),
		LikeCount:    int(pm.GetLikeCount()),
		CommentCount: int(pm.GetCommentCount()),
		RepostCount:  int(pm.GetRepostCount()),
		Status:       models.Status(pm.GetStatus()),
	}
	if pm.RepostFromId != nil {
		v := uint(*pm.RepostFromId)
		m.RepostFromID = &v
	}
	if rf := pm.GetRepostFrom(); rf != nil {
		sub := fromProtoMoment(rf)
		m.RepostFrom = &sub
	}
	return m
}

// CreateMoment 发布动态/日记。
func CreateMoment(userID uint, typ models.MomentType, vis models.MomentVisibility, content string, images []string, st models.Status) (models.MomentModel, error) {
	c, err := client()
	if err != nil {
		return models.MomentModel{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.CreateMoment(ctx, &contentv1.CreateMomentRequest{
		UserId: uint64(userID), Type: int32(typ), Visibility: int32(vis), Content: content, Images: images, Status: int32(st),
	})
	if err != nil {
		return models.MomentModel{}, err
	}
	return fromProtoMoment(rep.GetMoment()), nil
}

// MomentListQuery 动态列表查询(隐私过滤在服务端)。
type MomentListQuery struct {
	ViewerID uint
	IsAdmin  bool
	UserID   uint
	Type     *models.MomentType
	Page     int
	Limit    int
	Key      string
	Order    string
	EndID    uint
}

// ListMoments 动态列表。
func ListMoments(q MomentListQuery) ([]models.MomentModel, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	req := &contentv1.ListMomentsRequest{
		ViewerId: uint64(q.ViewerID), IsAdmin: q.IsAdmin, UserId: uint64(q.UserID),
		Page: int32(q.Page), Limit: int32(q.Limit), Key: q.Key, Order: q.Order, EndId: uint64(q.EndID),
	}
	if q.Type != nil {
		v := int32(*q.Type)
		req.Type = &v
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.ListMoments(ctx, req)
	if err != nil {
		return nil, 0, false, err
	}
	list := make([]models.MomentModel, 0, len(rep.GetList()))
	for _, pm := range rep.GetList() {
		list = append(list, fromProtoMoment(pm))
	}
	return list, int(rep.GetCount()), rep.GetCapped(), nil
}

// GetMoment 动态详情。
func GetMoment(id, viewerID uint, isAdmin bool) (models.MomentModel, error) {
	c, err := client()
	if err != nil {
		return models.MomentModel{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.GetMoment(ctx, &contentv1.GetMomentRequest{Id: uint64(id), ViewerId: uint64(viewerID), IsAdmin: isAdmin})
	if err != nil {
		return models.MomentModel{}, err
	}
	return fromProtoMoment(rep.GetMoment()), nil
}

// UpdateMoment 更新本人动态。
func UpdateMoment(id, userID uint, content string, images []string, typ models.MomentType, vis models.MomentVisibility, st models.Status) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	_, err = c.UpdateMoment(ctx, &contentv1.UpdateMomentRequest{
		Id: uint64(id), UserId: uint64(userID), Content: content, Images: images,
		Type: int32(typ), Visibility: int32(vis), Status: int32(st),
	})
	return err
}

// RemoveMoment 删除动态。
func RemoveMoment(id, userID uint, isAdmin bool) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	_, err = c.RemoveMoment(ctx, &contentv1.RemoveMomentRequest{Id: uint64(id), UserId: uint64(userID), IsAdmin: isAdmin})
	return err
}

// MomentInteraction 当前用户对该动态的点赞状态。
func MomentInteraction(id, viewerID uint) (bool, error) {
	c, err := client()
	if err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.MomentInteraction(ctx, &contentv1.MomentInteractionRequest{Id: uint64(id), ViewerId: uint64(viewerID)})
	if err != nil {
		return false, err
	}
	return rep.GetDigged(), nil
}

// ToggleMomentDigg 点赞/取消动态,返回(是否已点赞, 点赞数)。
func ToggleMomentDigg(id, userID uint) (bool, int, error) {
	c, err := client()
	if err != nil {
		return false, 0, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.ToggleMomentDigg(ctx, &contentv1.ToggleMomentDiggRequest{Id: uint64(id), UserId: uint64(userID)})
	if err != nil {
		return false, 0, err
	}
	return rep.GetDigged(), int(rep.GetLikeCount()), nil
}

// RepostMoment 转发动态。
func RepostMoment(id, userID, viewerID uint, isAdmin bool, content string, vis models.MomentVisibility) (models.MomentModel, error) {
	c, err := client()
	if err != nil {
		return models.MomentModel{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.RepostMoment(ctx, &contentv1.RepostMomentRequest{
		Id: uint64(id), UserId: uint64(userID), ViewerId: uint64(viewerID), IsAdmin: isAdmin,
		Content: content, Visibility: int32(vis),
	})
	if err != nil {
		return models.MomentModel{}, err
	}
	return fromProtoMoment(rep.GetMoment()), nil
}
