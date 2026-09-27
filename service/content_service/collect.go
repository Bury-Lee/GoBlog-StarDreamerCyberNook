package content_service

import (
	"context"
	"time"

	contentv1 "StarDreamerCyberNook/gen/content/v1"
	"StarDreamerCyberNook/models"
)

func fromProtoFolder(f *contentv1.CollectFolder) models.CollectModel {
	m := models.CollectModel{
		Title: f.GetTitle(), UserID: uint(f.GetUserId()), Abstract: f.GetAbstract(),
		Cover: f.GetCover(), IsPublic: f.GetIsPublic(), IsDefault: f.GetIsDefault(),
	}
	m.ID = uint(f.GetId())
	m.CreatedAt = time.UnixMilli(f.GetCreatedAt())
	m.UpdatedAt = time.UnixMilli(f.GetUpdatedAt())
	return m
}

// ToggleArticleCollect 收藏/取消/移动,返回动作 created|removed|moved。
func ToggleArticleCollect(userID, articleID, collectID uint) (string, error) {
	c, err := client()
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.ToggleArticleCollect(ctx, &contentv1.ToggleCollectRequest{
		UserId: uint64(userID), ArticleId: uint64(articleID), CollectId: uint64(collectID),
	})
	if err != nil {
		return "", mapErr(err)
	}
	return rep.GetAction(), nil
}

// CreateCollectFolder 创建收藏夹。
func CreateCollectFolder(userID uint, title, abstract, cover string) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	_, err = c.CreateCollectFolder(ctx, &contentv1.CreateFolderRequest{
		UserId: uint64(userID), Title: title, Abstract: abstract, Cover: cover,
	})
	return err
}

// UpdateCollectFolder 增量更新收藏夹(patch 为 JSON 列补丁)。
func UpdateCollectFolder(userID, id uint, patch string) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	_, err = c.UpdateCollectFolder(ctx, &contentv1.UpdateFolderRequest{UserId: uint64(userID), Id: uint64(id), Patch: patch})
	return mapErr(err)
}

// RemoveCollectFolders 删除收藏夹,返回 文章ID -> 收藏数回退增量(负数) 与实际删除的夹数量。
func RemoveCollectFolders(userID uint, isAdmin bool, ids []uint) (map[uint]int, int, error) {
	c, err := client()
	if err != nil {
		return nil, 0, err
	}
	req := &contentv1.RemoveFoldersRequest{UserId: uint64(userID), IsAdmin: isAdmin}
	for _, id := range ids {
		req.Ids = append(req.Ids, uint64(id))
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.RemoveCollectFolders(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	out := make(map[uint]int, len(rep.GetArticleDeltas()))
	for aid, d := range rep.GetArticleDeltas() {
		out[uint(aid)] = int(d)
	}
	return out, int(rep.GetDeleted()), nil
}

// ListCollectFolders 收藏夹列表(includePrivate=false 时仅公开)。
func ListCollectFolders(userID uint, includePrivate bool, page, limit int, key, order string, endID uint) ([]models.CollectModel, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.ListCollectFolders(ctx, &contentv1.ListFoldersRequest{
		UserId: uint64(userID), IncludePrivate: includePrivate, Key: key,
		Page: int32(page), Limit: int32(limit), Order: order, EndId: uint64(endID),
	})
	if err != nil {
		return nil, 0, false, err
	}
	out := make([]models.CollectModel, 0, len(rep.GetList()))
	for _, f := range rep.GetList() {
		out = append(out, fromProtoFolder(f))
	}
	return out, int(rep.GetCount()), rep.GetCapped(), nil
}

// GetCollectFolder 收藏夹详情 + 内文章数。
func GetCollectFolder(id uint) (models.CollectModel, int64, error) {
	c, err := client()
	if err != nil {
		return models.CollectModel{}, 0, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.GetCollectFolder(ctx, &contentv1.GetFolderRequest{Id: uint64(id)})
	if err != nil {
		return models.CollectModel{}, 0, mapErr(err)
	}
	return fromProtoFolder(rep.GetFolder()), rep.GetArticleCount(), nil
}

// ListCollectArticles 收藏夹内文章(按收藏时间倒序)。
func ListCollectArticles(folderID uint, page, limit int, key, order string, endID uint) ([]models.ArticleModel, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.ListCollectArticles(ctx, &contentv1.ListCollectArticlesRequest{
		FolderId: uint64(folderID), Key: key, Page: int32(page), Limit: int32(limit), Order: order, EndId: uint64(endID),
	})
	if err != nil {
		return nil, 0, false, err
	}
	out := make([]models.ArticleModel, 0, len(rep.GetList()))
	for _, a := range rep.GetList() {
		out = append(out, fromProtoArticle(a))
	}
	return out, int(rep.GetCount()), rep.GetCapped(), nil
}
