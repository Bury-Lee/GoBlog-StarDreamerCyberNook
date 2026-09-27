package content_service

import (
	"context"
	"time"

	contentv1 "StarDreamerCyberNook/gen/content/v1"
	"StarDreamerCyberNook/models"
)

func fromProtoCommentUser(u *contentv1.CommentUser) models.UserModel {
	if u == nil {
		return models.UserModel{}
	}
	return models.UserModel{
		Model:         models.Model{ID: uint(u.GetId()), CreatedAt: time.UnixMilli(u.GetCreatedAt()), UpdatedAt: time.UnixMilli(u.GetUpdatedAt())},
		NickName:      u.GetNickname(),
		Avatar:        u.GetAvatar(),
		Age:           int(u.GetAge()),
		LikeTags:      u.GetLikeTags(),
		LastLoginTime: time.UnixMilli(u.GetLastLoginTime()),
	}
}

func fromProtoComment(c *contentv1.Comment) models.CommentModel {
	m := models.CommentModel{
		Model:      models.Model{ID: uint(c.GetId()), CreatedAt: time.UnixMilli(c.GetCreatedAt()), UpdatedAt: time.UnixMilli(c.GetUpdatedAt())},
		Content:    c.GetContent(),
		UserID:     uint(c.GetUserId()),
		ArticleID:  uint(c.GetArticleId()),
		ParentPath: c.GetPath(),
		DiggCount:  int(c.GetDiggCount()),
		UserModel:  fromProtoCommentUser(c.GetUser()),
	}
	if c.RootParentId != nil {
		v := uint(*c.RootParentId)
		m.RootParentID = &v
	}
	return m
}

func mapComments(rep *contentv1.ListCommentsReply) []models.CommentModel {
	list := make([]models.CommentModel, 0, len(rep.GetList()))
	for _, c := range rep.GetList() {
		list = append(list, fromProtoComment(c))
	}
	return list
}

// CreateCommentResult 创建评论结果。
type CreateCommentResult struct {
	ID              uint
	ReplyRevUserIDs []uint
	ArticleUserID   uint
}

// CreateComment 创建评论。
func CreateComment(userID, articleID, parentID uint, content string) (CreateCommentResult, error) {
	c, err := client()
	if err != nil {
		return CreateCommentResult{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.CreateComment(ctx, &contentv1.CreateCommentRequest{
		UserId: uint64(userID), ArticleId: uint64(articleID), ParentId: uint64(parentID), Content: content,
	})
	if err != nil {
		return CreateCommentResult{}, err
	}
	ids := make([]uint, 0, len(rep.GetReplyRevUserIds()))
	for _, id := range rep.GetReplyRevUserIds() {
		ids = append(ids, uint(id))
	}
	return CreateCommentResult{ID: uint(rep.GetId()), ReplyRevUserIDs: ids, ArticleUserID: uint(rep.GetArticleUserId())}, nil
}

// ListComments 一级评论列表。
func ListComments(articleID uint, page, limit int, key, order string, endID uint) ([]models.CommentModel, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.ListComments(ctx, &contentv1.ListCommentsRequest{
		ArticleId: uint64(articleID), Page: int32(page), Limit: int32(limit), Key: key, Order: order, EndId: uint64(endID),
	})
	if err != nil {
		return nil, 0, false, err
	}
	return mapComments(rep), int(rep.GetCount()), rep.GetCapped(), nil
}

// ListChildComments 子评论列表。
func ListChildComments(rootID uint, page, limit int, key, order string, endID uint) ([]models.CommentModel, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.ListChildComments(ctx, &contentv1.ListCommentsRequest{
		RootId: uint64(rootID), Page: int32(page), Limit: int32(limit), Key: key, Order: order, EndId: uint64(endID),
	})
	if err != nil {
		return nil, 0, false, err
	}
	return mapComments(rep), int(rep.GetCount()), rep.GetCapped(), nil
}

// DeleteCommentResult 删除评论结果。
type DeleteCommentResult struct {
	ArticleID uint
	Delta     int
	Ok        bool
}

// DeleteComment 删除评论。
func DeleteComment(id, userID uint, isAdmin bool) (DeleteCommentResult, error) {
	c, err := client()
	if err != nil {
		return DeleteCommentResult{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.DeleteComment(ctx, &contentv1.DeleteCommentRequest{Id: uint64(id), UserId: uint64(userID), IsAdmin: isAdmin})
	if err != nil {
		return DeleteCommentResult{}, err
	}
	return DeleteCommentResult{ArticleID: uint(rep.GetArticleId()), Delta: int(rep.GetDelta()), Ok: rep.GetOk()}, nil
}

// ToggleDiggResult 评论点赞结果。
type ToggleDiggResult struct {
	Digged        bool
	BaseDiggCount int
}

// ToggleCommentDigg 切换评论点赞。
func ToggleCommentDigg(commentID, userID uint) (ToggleDiggResult, error) {
	c, err := client()
	if err != nil {
		return ToggleDiggResult{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.ToggleCommentDigg(ctx, &contentv1.ToggleCommentDiggRequest{CommentId: uint64(commentID), UserId: uint64(userID)})
	if err != nil {
		return ToggleDiggResult{}, err
	}
	return ToggleDiggResult{Digged: rep.GetDigged(), BaseDiggCount: int(rep.GetBaseDiggCount())}, nil
}

// CommentDiggIDs 查询用户在一组评论上的点赞状态。
func CommentDiggIDs(userID uint, ids []uint) ([]uint, error) {
	c, err := client()
	if err != nil {
		return nil, err
	}
	req := &contentv1.CommentDiggIDsRequest{UserId: uint64(userID)}
	for _, id := range ids {
		req.CommentIds = append(req.CommentIds, uint64(id))
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.CommentDiggIDs(ctx, req)
	if err != nil {
		return nil, err
	}
	out := make([]uint, 0, len(rep.GetDiggedIds()))
	for _, id := range rep.GetDiggedIds() {
		out = append(out, uint(id))
	}
	return out, nil
}
