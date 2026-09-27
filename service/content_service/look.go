package content_service

import (
	"context"
	"time"

	contentv1 "StarDreamerCyberNook/gen/content/v1"
)

// RecordArticleLook 记录一次浏览;created=false 表示当日已读过。
func RecordArticleLook(userID, articleID uint) (created bool, err error) {
	c, err := client()
	if err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.RecordArticleLook(ctx, &contentv1.RecordArticleLookRequest{
		UserId: uint64(userID), ArticleId: uint64(articleID),
	})
	if err != nil {
		return false, mapErr(err)
	}
	return rep.GetCreated(), nil
}

// LookItem 浏览足迹条目。
type LookItem struct {
	ID        uint
	LookDate  time.Time
	Title     string
	Cover     string
	Nickname  string
	Avatar    string
	UserID    uint
	ArticleID uint
}

// ListArticleLook 浏览足迹列表。
func ListArticleLook(userID, viewerID uint, viewerLogged bool, page, limit int, order string, endID uint) ([]LookItem, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.ListArticleLook(ctx, &contentv1.ListArticleLookRequest{
		UserId: uint64(userID), ViewerId: uint64(viewerID), ViewerLogged: viewerLogged,
		Page: int32(page), Limit: int32(limit), Order: order, EndId: uint64(endID),
	})
	if err != nil {
		return nil, 0, false, mapErr(err)
	}
	out := make([]LookItem, 0, len(rep.GetList()))
	for _, it := range rep.GetList() {
		out = append(out, LookItem{
			ID: uint(it.GetId()), LookDate: time.UnixMilli(it.GetLookDate()),
			Title: it.GetTitle(), Cover: it.GetCover(), Nickname: it.GetNickname(), Avatar: it.GetAvatar(),
			UserID: uint(it.GetUserId()), ArticleID: uint(it.GetArticleId()),
		})
	}
	return out, int(rep.GetCount()), rep.GetCapped(), nil
}

// RemoveArticleLook 删除浏览历史,返回删除条数。
func RemoveArticleLook(userID uint, ids []uint) (int, error) {
	c, err := client()
	if err != nil {
		return 0, err
	}
	req := &contentv1.RemoveArticleLookRequest{UserId: uint64(userID)}
	for _, id := range ids {
		req.Ids = append(req.Ids, uint64(id))
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.RemoveArticleLook(ctx, req)
	if err != nil {
		return 0, err
	}
	return int(rep.GetDeleted()), nil
}
