package content_service

import (
	"context"

	contentv1 "StarDreamerCyberNook/gen/content/v1"
	"StarDreamerCyberNook/models"
)

// TopItem 置顶项。
type TopItem struct {
	ArticleID uint
	IsAdmin   bool
}

// UserTopArticles 用户的置顶文章。
func UserTopArticles(userID uint) ([]TopItem, error) {
	c, err := client()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.UserTopArticles(ctx, &contentv1.UserTopArticlesRequest{UserId: uint64(userID)})
	if err != nil {
		return nil, err
	}
	out := make([]TopItem, 0, len(rep.GetList()))
	for _, t := range rep.GetList() {
		out = append(out, TopItem{ArticleID: uint(t.GetArticleId()), IsAdmin: t.GetIsAdmin()})
	}
	return out, nil
}

// AdminTopArticleIDs 所有管理员置顶的文章ID。
func AdminTopArticleIDs() ([]uint, error) {
	c, err := client()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.AdminTopArticleIDs(ctx, &contentv1.Empty{})
	if err != nil {
		return nil, err
	}
	out := make([]uint, 0, len(rep.GetIds()))
	for _, id := range rep.GetIds() {
		out = append(out, uint(id))
	}
	return out, nil
}

// ArticleDetail 文章详情(批量)。
type ArticleDetail struct {
	ArticleModel  models.ArticleModel
	CategoryTitle *string
	NickName      string
	Avatar        string
}

// GetArticlesByIDs 按 ID 批量取文章详情(保持入参顺序)。
func GetArticlesByIDs(ids []uint, publishedOnly bool) ([]ArticleDetail, error) {
	c, err := client()
	if err != nil {
		return nil, err
	}
	req := &contentv1.GetArticlesByIDsRequest{PublishedOnly: publishedOnly}
	for _, id := range ids {
		req.Ids = append(req.Ids, uint64(id))
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.GetArticlesByIDs(ctx, req)
	if err != nil {
		return nil, err
	}
	out := make([]ArticleDetail, 0, len(rep.GetList()))
	for _, it := range rep.GetList() {
		d := ArticleDetail{
			ArticleModel: fromProtoArticle(it.GetArticle()),
			NickName:     it.GetNickName(),
			Avatar:       it.GetAvatar(),
		}
		if it.CategoryTitle != nil {
			t := it.GetCategoryTitle()
			d.CategoryTitle = &t
		}
		out = append(out, d)
	}
	return out, nil
}
