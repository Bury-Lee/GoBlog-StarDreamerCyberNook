package content_service

import (
	"context"

	contentv1 "StarDreamerCyberNook/gen/content/v1"
	"StarDreamerCyberNook/models"
)

func fromProtoArticleList(rep *contentv1.ArticleListReply) []models.ArticleModel {
	out := make([]models.ArticleModel, 0, len(rep.GetList()))
	for _, a := range rep.GetList() {
		out = append(out, fromProtoArticle(a))
	}
	return out
}

// ListArticlesByStatus 按状态查询文章(可选用户/ID 过滤,分页)。
func ListArticlesByStatus(status models.Status, userID uint, ids []uint, key string, page, limit int, order string, endID uint) ([]models.ArticleModel, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	req := &contentv1.ListArticlesByStatusRequest{
		Status: int32(status), UserId: uint64(userID), Key: key,
		Page: int32(page), Limit: int32(limit), Order: order, EndId: uint64(endID),
	}
	for _, id := range ids {
		req.Ids = append(req.Ids, uint64(id))
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.ListArticlesByStatus(ctx, req)
	if err != nil {
		return nil, 0, false, mapErr(err)
	}
	return fromProtoArticleList(rep), int(rep.GetCount()), rep.GetCapped(), nil
}

// PendingBatch 游标分批取待审核文章(供定时任务)。
func PendingBatch(afterID uint, limit int) ([]models.ArticleModel, error) {
	c, err := client()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.PendingBatch(ctx, &contentv1.PendingBatchRequest{AfterId: uint64(afterID), Limit: int32(limit)})
	if err != nil {
		return nil, err
	}
	return fromProtoArticleList(rep), nil
}

// CountArticlesByStatus 统计某状态文章数。
func CountArticlesByStatus(status models.Status) (int64, error) {
	c, err := client()
	if err != nil {
		return 0, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.CountArticlesByStatus(ctx, &contentv1.CountArticlesRequest{Status: int32(status)})
	if err != nil {
		return 0, err
	}
	return rep.GetCount(), nil
}

// SetArticleStatus 更新文章状态。
func SetArticleStatus(id uint, status models.Status) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	_, err = c.SetArticleStatus(ctx, &contentv1.SetArticleStatusRequest{Id: uint64(id), Status: int32(status)})
	return err
}

// SetArticleAddition 写入/更新文章扩展附录(AI 点评等)。
func SetArticleAddition(articleID uint, aiQuality, aiAbstract, aiModel string) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	_, err = c.SetArticleAddition(ctx, &contentv1.SetArticleAdditionRequest{
		ArticleId: uint64(articleID), AiQuality: aiQuality, AiAbstract: aiAbstract, AiModel: aiModel,
	})
	return err
}
