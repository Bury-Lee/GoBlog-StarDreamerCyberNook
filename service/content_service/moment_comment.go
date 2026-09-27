package content_service

import (
	"context"
	"time"

	contentv1 "StarDreamerCyberNook/gen/content/v1"
	"StarDreamerCyberNook/models"
)

func fromProtoMomentComment(c *contentv1.MomentComment) models.MomentCommentModel {
	m := models.MomentCommentModel{
		Model:      models.Model{ID: uint(c.GetId()), CreatedAt: time.UnixMilli(c.GetCreatedAt()), UpdatedAt: time.UnixMilli(c.GetUpdatedAt())},
		MomentID:   uint(c.GetMomentId()),
		UserID:     uint(c.GetUserId()),
		UserModel:  fromProtoCommentUser(c.GetUser()),
		Content:    c.GetContent(),
		ParentPath: c.GetPath(),
		DiggCount:  int(c.GetDiggCount()),
	}
	if c.RootParentId != nil {
		v := uint(*c.RootParentId)
		m.RootParentID = &v
	}
	return m
}

// CreateMomentComment 发表动态评论/回复。
func CreateMomentComment(userID, momentID, parentID uint, content string, viewerID uint, isAdmin bool) (models.MomentCommentModel, error) {
	c, err := client()
	if err != nil {
		return models.MomentCommentModel{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.CreateMomentComment(ctx, &contentv1.CreateMomentCommentRequest{
		UserId: uint64(userID), MomentId: uint64(momentID), ParentId: uint64(parentID),
		Content: content, ViewerId: uint64(viewerID), IsAdmin: isAdmin,
	})
	if err != nil {
		return models.MomentCommentModel{}, err
	}
	return fromProtoMomentComment(rep.GetComment()), nil
}

func mapMomentComments(rep *contentv1.ListMomentCommentsReply) []models.MomentCommentModel {
	list := make([]models.MomentCommentModel, 0, len(rep.GetList()))
	for _, c := range rep.GetList() {
		list = append(list, fromProtoMomentComment(c))
	}
	return list
}

// ListMomentComments 一级评论列表。
func ListMomentComments(momentID, viewerID uint, isAdmin bool, page, limit int, key, order string, endID uint) ([]models.MomentCommentModel, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.ListMomentComments(ctx, &contentv1.ListMomentCommentsRequest{
		MomentId: uint64(momentID), ViewerId: uint64(viewerID), IsAdmin: isAdmin,
		Page: int32(page), Limit: int32(limit), Key: key, Order: order, EndId: uint64(endID),
	})
	if err != nil {
		return nil, 0, false, err
	}
	return mapMomentComments(rep), int(rep.GetCount()), rep.GetCapped(), nil
}

// ListMomentChildComments 子评论列表。
func ListMomentChildComments(rootID uint, page, limit int, key, order string, endID uint) ([]models.MomentCommentModel, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.ListMomentChildComments(ctx, &contentv1.ListMomentChildCommentsRequest{
		RootId: uint64(rootID), Page: int32(page), Limit: int32(limit), Key: key, Order: order, EndId: uint64(endID),
	})
	if err != nil {
		return nil, 0, false, err
	}
	return mapMomentComments(rep), int(rep.GetCount()), rep.GetCapped(), nil
}

// DeleteMomentComment 删除动态评论。
func DeleteMomentComment(id, userID uint, isAdmin bool) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	_, err = c.DeleteMomentComment(ctx, &contentv1.DeleteMomentCommentRequest{Id: uint64(id), UserId: uint64(userID), IsAdmin: isAdmin})
	return err
}

// ToggleMomentCommentDigg 动态评论点赞/取消,返回(是否已点赞, 点赞数)。
func ToggleMomentCommentDigg(id, userID uint) (bool, int, error) {
	c, err := client()
	if err != nil {
		return false, 0, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.ToggleMomentCommentDigg(ctx, &contentv1.ToggleMomentCommentDiggRequest{Id: uint64(id), UserId: uint64(userID)})
	if err != nil {
		return false, 0, err
	}
	return rep.GetDigged(), int(rep.GetDiggCount()), nil
}
