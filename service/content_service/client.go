// Package content_service 是经 gRPC 调用独立 content 服务的客户端(内容域)。
package content_service

import (
	"context"
	"sync"
	"time"

	contentv1 "StarDreamerCyberNook/gen/content/v1"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/pkg/svc"
)

const contentTimeout = 10 * time.Second

var (
	cliOnce sync.Once
	cli     contentv1.ContentServiceClient
	cliErr  error
)

// client 懒加载到 content 服务的 gRPC 连接(服务键 "content")。
func client() (contentv1.ContentServiceClient, error) {
	cliOnce.Do(func() {
		var cfgs map[string][]string
		if global.Config != nil {
			cfgs = global.Config.Services
		}
		svc.Init(cfgs)
		conn, err := grpcx.Dial("content")
		if err != nil {
			cliErr = err
			return
		}
		cli = contentv1.NewContentServiceClient(conn)
	})
	return cli, cliErr
}

func fromProtoArticle(a *contentv1.Article) models.ArticleModel {
	m := models.ArticleModel{}
	m.ID = uint(a.GetId())
	m.Title = a.GetTitle()
	m.Abstract = a.GetAbstract()
	m.Content = a.GetContent()
	if a.CategoryId != nil {
		cid := uint(*a.CategoryId)
		m.CategoryID = &cid
	}
	m.TagList = a.GetTagList()
	m.Cover = a.GetCover()
	m.UserID = uint(a.GetUserId())
	m.LookCount = int(a.GetLookCount())
	m.DiggCount = int(a.GetDiggCount())
	m.CommentCount = int(a.GetCommentCount())
	m.CollectCount = int(a.GetCollectCount())
	m.OpenComment = a.GetOpenComment()
	m.Status = models.Status(a.GetStatus())
	m.CreatedAt = time.UnixMilli(a.GetCreatedAt())
	m.UpdatedAt = time.UnixMilli(a.GetUpdatedAt())
	return m
}

// ListReq 文章列表请求(可选项用指针区分"不筛选")。
type ListReq struct {
	UserID     uint
	CategoryID *uint
	Status     *models.Status
	Order      string
	Key        string
	Page       int
	Limit      int
	EndID      uint
	TopIDs     []uint
}

// ListItem 列表条目。
type ListItem struct {
	ArticleModel  models.ArticleModel
	UserNickName  string
	Avatar        string
	CategoryTitle string
}

// ListArticles 文章列表。
func ListArticles(req ListReq) ([]ListItem, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	pr := &contentv1.ListArticlesRequest{
		UserId: uint64(req.UserID),
		Order:  req.Order,
		Key:    req.Key,
		Page:   int32(req.Page),
		Limit:  int32(req.Limit),
		EndId:  uint64(req.EndID),
	}
	if req.CategoryID != nil {
		v := uint64(*req.CategoryID)
		pr.CategoryId = &v
	}
	if req.Status != nil {
		v := int32(*req.Status)
		pr.Status = &v
	}
	for _, id := range req.TopIDs {
		pr.TopIds = append(pr.TopIds, uint64(id))
	}

	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.ListArticles(ctx, pr)
	if err != nil {
		return nil, 0, false, err
	}

	items := make([]ListItem, 0, len(rep.GetList()))
	for _, it := range rep.GetList() {
		items = append(items, ListItem{
			ArticleModel:  fromProtoArticle(it.GetArticle()),
			UserNickName:  it.GetUserNickName(),
			Avatar:        it.GetAvatar(),
			CategoryTitle: it.GetCategoryTitle(),
		})
	}
	return items, int(rep.GetCount()), rep.GetCapped(), nil
}

// Detail 文章详情。
type Detail struct {
	ArticleModel  models.ArticleModel
	CategoryTitle *string
	UserName      string
	NickName      string
	UserAvatar    string
}

// GetArticle 文章详情。
func GetArticle(id uint) (*Detail, error) {
	c, err := client()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.GetArticle(ctx, &contentv1.GetArticleRequest{Id: uint64(id)})
	if err != nil {
		return nil, err
	}

	m := fromProtoArticle(rep.GetArticle())
	if ad := rep.GetAddition(); ad != nil {
		m.ArticleAddition = &models.ArticleAddition{
			AdminComment: ad.GetAdminComment(),
			AIQuality:    ad.GetAiQuality(),
			AIAbstract:   ad.GetAiAbstract(),
			AIModel:      ad.GetAiModel(),
		}
	}

	d := &Detail{
		ArticleModel: m,
		UserName:     rep.GetUserName(),
		NickName:     rep.GetNickName(),
		UserAvatar:   rep.GetUserAvatar(),
	}
	if rep.GetCategoryTitle() != "" {
		t := rep.GetCategoryTitle()
		d.CategoryTitle = &t
	}
	return d, nil
}

// RemoveResult 删除结果。
type RemoveResult struct {
	Deleted int
	Titles  []string
}

// RemoveArticles 批量删除文章;ownerID 非空时校验归属。
func RemoveArticles(ids []uint, ownerID *uint) (RemoveResult, error) {
	c, err := client()
	if err != nil {
		return RemoveResult{}, err
	}
	pr := &contentv1.RemoveArticlesRequest{}
	for _, id := range ids {
		pr.Ids = append(pr.Ids, uint64(id))
	}
	if ownerID != nil {
		v := uint64(*ownerID)
		pr.OwnerId = &v
	}

	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.RemoveArticles(ctx, pr)
	if err != nil {
		return RemoveResult{}, err
	}
	return RemoveResult{Deleted: int(rep.GetDeleted()), Titles: rep.GetTitles()}, nil
}

// CreateReq 创建文章请求。
type CreateReq struct {
	UserID      uint
	Title       string
	Abstract    string
	Content     string
	CategoryID  *uint
	TagList     []string
	Cover       string
	OpenComment bool
	Status      models.Status
	AIQuality   string
	AIAbstract  string
	AIModel     string
}

// CreateArticle 创建文章,返回文章ID。
func CreateArticle(req CreateReq) (uint, error) {
	c, err := client()
	if err != nil {
		return 0, err
	}
	pr := &contentv1.CreateArticleRequest{
		UserId:      uint64(req.UserID),
		Title:       req.Title,
		Abstract:    req.Abstract,
		Content:     req.Content,
		TagList:     req.TagList,
		Cover:       req.Cover,
		OpenComment: req.OpenComment,
		Status:      int32(req.Status),
		AiQuality:   req.AIQuality,
		AiAbstract:  req.AIAbstract,
		AiModel:     req.AIModel,
	}
	if req.CategoryID != nil {
		v := uint64(*req.CategoryID)
		pr.CategoryId = &v
	}

	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.CreateArticle(ctx, pr)
	if err != nil {
		return 0, err
	}
	return uint(rep.GetId()), nil
}

// UpdateReq 文章更新补丁:指针为 nil 的字段不更新。
type UpdateReq struct {
	ID          uint
	OwnerID     uint
	Title       *string
	Abstract    *string
	Content     *string
	CategoryID  *uint // 非 nil 时:0 表示清空分类
	OpenComment *bool
	Cover       *string
	Status      *models.Status
	TagList     *[]string
	AIQuality   string
	AIAbstract  string
	AIModel     string
}

// UpdateArticle 增量更新文章。
func UpdateArticle(req UpdateReq) error {
	c, err := client()
	if err != nil {
		return err
	}
	pr := &contentv1.UpdateArticleRequest{
		Id:         uint64(req.ID),
		OwnerId:    uint64(req.OwnerID),
		AiQuality:  req.AIQuality,
		AiAbstract: req.AIAbstract,
		AiModel:    req.AIModel,
	}
	if req.Title != nil {
		pr.Title = req.Title
	}
	if req.Abstract != nil {
		pr.Abstract = req.Abstract
	}
	if req.Content != nil {
		pr.Content = req.Content
	}
	if req.Cover != nil {
		pr.Cover = req.Cover
	}
	if req.OpenComment != nil {
		pr.OpenComment = req.OpenComment
	}
	if req.CategoryID != nil {
		v := uint64(*req.CategoryID)
		pr.CategoryId = &v
	}
	if req.Status != nil {
		v := int32(*req.Status)
		pr.Status = &v
	}
	if req.TagList != nil {
		pr.TagList = &contentv1.TagListUpdate{Tags: *req.TagList}
	}

	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	_, err = c.UpdateArticle(ctx, pr)
	return err
}

// Category 分类。
type Category struct {
	ID        uint
	Title     string
	UserID    uint
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateCategory 创建或更新分类(id=0 创建)。
func CreateCategory(id uint, title string, userID uint) error {
	c, err := client()
	if err != nil {
		return err
	}
	pr := &contentv1.CreateCategoryRequest{Title: title, UserId: uint64(userID)}
	if id != 0 {
		v := uint64(id)
		pr.Id = &v
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	_, err = c.CreateCategory(ctx, pr)
	return err
}

// CategoryItem 分类列表条目。
type CategoryItem struct {
	Category     Category
	ArticleCount int
	Nickname     string
	Avatar       string
}

// ListCategories 分类列表。
func ListCategories(userID uint, withUser bool, page, limit int, key, order string) ([]CategoryItem, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.ListCategories(ctx, &contentv1.ListCategoriesRequest{
		UserId: uint64(userID), WithUser: withUser, Page: int32(page), Limit: int32(limit), Key: key, Order: order,
	})
	if err != nil {
		return nil, 0, false, err
	}
	items := make([]CategoryItem, 0, len(rep.GetList()))
	for _, it := range rep.GetList() {
		cat := it.GetCategory()
		items = append(items, CategoryItem{
			Category: Category{
				ID:        uint(cat.GetId()),
				Title:     cat.GetTitle(),
				UserID:    uint(cat.GetUserId()),
				CreatedAt: time.UnixMilli(cat.GetCreatedAt()),
				UpdatedAt: time.UnixMilli(cat.GetUpdatedAt()),
			},
			ArticleCount: int(it.GetArticleCount()),
			Nickname:     it.GetNickname(),
			Avatar:       it.GetAvatar(),
		})
	}
	return items, int(rep.GetCount()), rep.GetCapped(), nil
}

// RemoveCategories 删除分类;all=true 为管理员。
func RemoveCategories(ids []uint, userID uint, all bool) (int, error) {
	c, err := client()
	if err != nil {
		return 0, err
	}
	pr := &contentv1.RemoveCategoriesRequest{UserId: uint64(userID), All: all}
	for _, id := range ids {
		pr.Ids = append(pr.Ids, uint64(id))
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.RemoveCategories(ctx, pr)
	if err != nil {
		return 0, err
	}
	return int(rep.GetDeleted()), nil
}

// CategoryOption 分类选项。
type CategoryOption struct {
	Value uint
	Label string
}

// CategoryOptions 分类选项列表。
func CategoryOptions(userID uint) ([]CategoryOption, error) {
	c, err := client()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.CategoryOptions(ctx, &contentv1.CategoryOptionsRequest{UserId: uint64(userID)})
	if err != nil {
		return nil, err
	}
	opts := make([]CategoryOption, 0, len(rep.GetOptions()))
	for _, o := range rep.GetOptions() {
		opts = append(opts, CategoryOption{Value: uint(o.GetValue()), Label: o.GetLabel()})
	}
	return opts, nil
}
