package internal

import (
	"context"
	"fmt"

	"StarDreamerCyberNook/common"
	contentv1 "StarDreamerCyberNook/gen/content/v1"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/pkg/dbx"
	"StarDreamerCyberNook/utils/sql"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// Config 是 content 服务的配置。
type Config struct {
	Driver     string // mysql | postgres | sqlite
	DSN        string
	Standalone bool // true: 自建独立库连接池;false: 复用宿主 global.DB
}

// Server 实现 ContentService。
type Server struct {
	contentv1.UnimplementedContentServiceServer
	db     *gorm.DB
	driver string
}

// NewServer 获取数据库句柄:未启用独立库时复用宿主 global.DB,启用时自建连接池。
func NewServer(cfg Config) (*Server, error) {
	db := global.DB
	if cfg.Standalone {
		d, err := dbx.Open(cfg.Driver, cfg.DSN)
		if err != nil {
			return nil, fmt.Errorf("content: open db: %w", err)
		}
		db = d
	}
	if db == nil {
		return nil, fmt.Errorf("content: 无可用数据库(共享模式需宿主先初始化 global.DB;独立模式请开启 dbStandalone)")
	}
	return &Server{db: db, driver: cfg.Driver}, nil
}

func toProtoArticle(a models.ArticleModel) *contentv1.Article {
	pa := &contentv1.Article{
		Id:           uint64(a.ID),
		Title:        a.Title,
		Abstract:     a.Abstract,
		Content:      a.Content,
		TagList:      a.TagList,
		Cover:        a.Cover,
		UserId:       uint64(a.UserID),
		LookCount:    int32(a.LookCount),
		DiggCount:    int32(a.DiggCount),
		CommentCount: int32(a.CommentCount),
		CollectCount: int32(a.CollectCount),
		OpenComment:  a.OpenComment,
		Status:       int32(a.Status),
		CreatedAt:    a.CreatedAt.UnixMilli(),
		UpdatedAt:    a.UpdatedAt.UnixMilli(),
	}
	if a.CategoryID != nil {
		cid := uint64(*a.CategoryID)
		pa.CategoryId = &cid
	}
	return pa
}

// ListArticles 文章列表(只读)。
func (s *Server) ListArticles(_ context.Context, req *contentv1.ListArticlesRequest) (*contentv1.ListArticlesReply, error) {
	var categoryID *uint
	if req.CategoryId != nil {
		c := uint(*req.CategoryId)
		categoryID = &c
	}

	opts := common.Options{
		Likes:         []string{"title"},
		PageInfo:      common.PageInfo{Page: int(req.GetPage()), Limit: int(req.GetLimit()), Key: req.GetKey(), Order: req.GetOrder(), EndId: uint(req.GetEndId())},
		Preloads:      []string{"UserModel", "CategoryModel"},
		DefaultOrder:  "created_at desc",
		AllowedOrders: []string{"created_at", "look_count", "digg_count", "comment_count", "collect_count"},
		CountCap:      common.DefaultCountCap,
	}
	if len(req.GetTopIds()) > 0 {
		top := make([]uint, 0, len(req.GetTopIds()))
		for _, id := range req.GetTopIds() {
			top = append(top, uint(id))
		}
		opts.DefaultOrder = fmt.Sprintf("%s, created_at desc", sql.ConvertSliceOrderSqlBy(s.driver, top))
	}
	if req.Status != nil {
		opts.Where = s.db.Where("status = ?", *req.Status)
	}

	list, count, capped, err := common.ListQuery(s.db, models.ArticleModel{
		UserID:     uint(req.GetUserId()),
		CategoryID: categoryID,
	}, opts)
	if err != nil {
		return nil, err
	}

	reply := &contentv1.ListArticlesReply{Count: int64(count), Capped: capped}
	for _, m := range list {
		item := &contentv1.ArticleListItem{
			Article:      toProtoArticle(m),
			UserNickName: m.UserModel.NickName,
			Avatar:       m.UserModel.Avatar,
		}
		if m.CategoryID != nil && m.CategoryModel != nil {
			item.CategoryTitle = m.CategoryModel.Title
		}
		reply.List = append(reply.List, item)
	}
	return reply, nil
}

// GetArticle 文章详情(含作者、分类、AI点评附录)。
func (s *Server) GetArticle(_ context.Context, req *contentv1.GetArticleRequest) (*contentv1.GetArticleReply, error) {
	var article models.ArticleModel
	if err := s.db.Preload("UserModel").Preload("CategoryModel").Preload("ArticleAddition").
		Take(&article, uint(req.GetId())).Error; err != nil {
		return nil, err
	}

	rep := &contentv1.GetArticleReply{
		Article:    toProtoArticle(article),
		UserName:   article.UserModel.UserName,
		NickName:   article.UserModel.NickName,
		UserAvatar: article.UserModel.Avatar,
	}
	if article.CategoryModel != nil {
		rep.CategoryTitle = article.CategoryModel.Title
	}
	if article.ArticleAddition != nil {
		rep.Addition = &contentv1.ArticleAddition{
			AdminComment: article.ArticleAddition.AdminComment,
			AiQuality:    article.ArticleAddition.AIQuality,
			AiAbstract:   article.ArticleAddition.AIAbstract,
			AiModel:      article.ArticleAddition.AIModel,
		}
	}
	return rep, nil
}

// RemoveArticles 批量删除文章;owner_id 非空时校验归属。
func (s *Server) RemoveArticles(_ context.Context, req *contentv1.RemoveArticlesRequest) (*contentv1.RemoveArticlesReply, error) {
	ids := req.GetIds()
	if len(ids) == 0 {
		return &contentv1.RemoveArticlesReply{}, nil
	}
	uintIDs := make([]uint, len(ids))
	for i, id := range ids {
		uintIDs[i] = uint(id)
	}

	if req.OwnerId != nil {
		var cnt int64
		if err := s.db.Model(&models.ArticleModel{}).
			Where("user_id = ? AND id IN ?", uint(*req.OwnerId), uintIDs).
			Count(&cnt).Error; err != nil {
			return nil, err
		}
		if int(cnt) != len(uintIDs) {
			return nil, status.Error(codes.FailedPrecondition, "部分文章不存在或不属于当前用户")
		}
	}

	var list []models.ArticleModel
	if err := s.db.Find(&list, "id in ?", uintIDs).Error; err != nil {
		return nil, err
	}
	if len(list) > 0 {
		if err := s.db.Delete(&list).Error; err != nil {
			return nil, err
		}
	}

	rep := &contentv1.RemoveArticlesReply{Deleted: int32(len(list))}
	for _, a := range list {
		rep.Titles = append(rep.Titles, a.Title)
	}
	return rep, nil
}

// CreateArticle 创建文章(含分类归属校验、AI点评写入)。
func (s *Server) CreateArticle(_ context.Context, req *contentv1.CreateArticleRequest) (*contentv1.CreateArticleReply, error) {
	if req.CategoryId != nil && *req.CategoryId != 0 {
		var cat models.CategoryModel
		if err := s.db.Take(&cat, "id = ? and user_id = ?", uint(*req.CategoryId), uint(req.GetUserId())).Error; err != nil {
			return nil, status.Error(codes.NotFound, "分类不存在")
		}
	}

	article := models.ArticleModel{
		Title:       req.GetTitle(),
		UserID:      uint(req.GetUserId()),
		Abstract:    req.GetAbstract(),
		Content:     req.GetContent(),
		TagList:     req.GetTagList(),
		Cover:       req.GetCover(),
		OpenComment: req.GetOpenComment(),
		Status:      models.Status(req.GetStatus()),
	}
	if req.CategoryId != nil {
		cid := uint(*req.CategoryId)
		article.CategoryID = &cid
	}

	if err := s.db.Create(&article).Error; err != nil {
		return nil, err
	}

	if req.GetAiQuality() != "" || req.GetAiAbstract() != "" {
		if e := s.db.Where(models.ArticleAddition{ArticleID: article.ID}).
			Assign(map[string]any{
				"ai_quality":  req.GetAiQuality(),
				"ai_abstract": req.GetAiAbstract(),
				"ai_model":    req.GetAiModel(),
			}).
			FirstOrCreate(&models.ArticleAddition{}).Error; e != nil {
			// 不影响文章创建,交由定时任务补全
			_ = e
		}
	}

	return &contentv1.CreateArticleReply{Id: uint64(article.ID)}, nil
}

// UpdateArticle 增量更新文章(归属校验 + 分类校验 + 可选标签与AI点评)。
func (s *Server) UpdateArticle(_ context.Context, req *contentv1.UpdateArticleRequest) (*contentv1.UpdateArticleReply, error) {
	var article models.ArticleModel
	if err := s.db.Take(&article, uint(req.GetId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "文章不存在")
	}
	if article.UserID != uint(req.GetOwnerId()) {
		return nil, status.Error(codes.PermissionDenied, "只能更新自己的文章")
	}

	if req.CategoryId != nil && *req.CategoryId != 0 {
		var cat models.CategoryModel
		if err := s.db.Take(&cat, "id = ? and user_id = ?", uint(*req.CategoryId), uint(req.GetOwnerId())).Error; err != nil {
			return nil, status.Error(codes.NotFound, "文章分类不存在")
		}
	}

	mps := map[string]any{}
	if req.Title != nil {
		mps["title"] = req.GetTitle()
	}
	if req.Abstract != nil {
		mps["abstract"] = req.GetAbstract()
	}
	if req.Content != nil {
		mps["content"] = req.GetContent()
	}
	if req.Cover != nil {
		mps["cover"] = req.GetCover()
	}
	if req.OpenComment != nil {
		mps["open_comment"] = req.GetOpenComment()
	}
	if req.CategoryId != nil {
		if *req.CategoryId == 0 {
			mps["category_id"] = nil
		} else {
			mps["category_id"] = uint(*req.CategoryId)
		}
	}
	if req.Status != nil {
		mps["status"] = models.Status(req.GetStatus())
	}

	if len(mps) > 0 {
		if err := s.db.Model(&article).Updates(mps).Error; err != nil {
			return nil, err
		}
	}

	// tag_list 走结构体更新,保证 serializer:json 生效
	if req.TagList != nil {
		if err := s.db.Model(&article).Select("tag_list").
			Updates(models.ArticleModel{TagList: req.TagList.GetTags()}).Error; err != nil {
			return nil, err
		}
	}

	if req.GetAiQuality() != "" || req.GetAiAbstract() != "" {
		_ = s.db.Where(models.ArticleAddition{ArticleID: article.ID}).
			Assign(map[string]any{
				"ai_quality":  req.GetAiQuality(),
				"ai_abstract": req.GetAiAbstract(),
				"ai_model":    req.GetAiModel(),
			}).
			FirstOrCreate(&models.ArticleAddition{}).Error
	}

	return &contentv1.UpdateArticleReply{}, nil
}

func toProtoCategory(c models.CategoryModel) *contentv1.Category {
	return &contentv1.Category{
		Id:        uint64(c.ID),
		Title:     c.Title,
		UserId:    uint64(c.UserID),
		CreatedAt: c.CreatedAt.UnixMilli(),
		UpdatedAt: c.UpdatedAt.UnixMilli(),
	}
}

// CreateCategory 创建或更新分类(ID 为空/0 创建,否则更新本人分类)。
func (s *Server) CreateCategory(_ context.Context, req *contentv1.CreateCategoryRequest) (*contentv1.CreateCategoryReply, error) {
	title := req.GetTitle()
	uid := uint(req.GetUserId())

	if req.Id == nil || *req.Id == 0 {
		var exist models.CategoryModel
		if err := s.db.Take(&exist, "user_id = ? and title = ?", uid, title).Error; err == nil {
			return nil, status.Error(codes.AlreadyExists, "分类名称重复")
		}
		m := models.CategoryModel{Title: title, UserID: uid}
		if err := s.db.Create(&m).Error; err != nil {
			return nil, err
		}
		return &contentv1.CreateCategoryReply{Id: uint64(m.ID)}, nil
	}

	var m models.CategoryModel
	if err := s.db.Take(&m, "user_id = ? and id = ?", uid, uint(*req.Id)).Error; err != nil {
		return nil, status.Error(codes.NotFound, "分类不存在")
	}
	if err := s.db.Model(&m).Update("title", title).Error; err != nil {
		return nil, err
	}
	return &contentv1.CreateCategoryReply{Id: uint64(m.ID)}, nil
}

// ListCategories 分类列表(含各分类文章数)。
func (s *Server) ListCategories(_ context.Context, req *contentv1.ListCategoriesRequest) (*contentv1.ListCategoriesReply, error) {
	opts := common.Options{
		PageInfo:      common.PageInfo{Page: int(req.GetPage()), Limit: int(req.GetLimit()), Key: req.GetKey(), Order: req.GetOrder()},
		Likes:         []string{"title"},
		AllowedOrders: []string{"id", "created_at"},
		CountCap:      common.DefaultCountCap,
	}
	if req.GetWithUser() {
		opts.Preloads = []string{"UserModel"}
	}

	list, count, capped, err := common.ListQuery(s.db, models.CategoryModel{UserID: uint(req.GetUserId())}, opts)
	if err != nil {
		return nil, err
	}

	countMap := make(map[uint]int, len(list))
	if len(list) > 0 {
		ids := make([]uint, 0, len(list))
		for _, it := range list {
			ids = append(ids, it.ID)
		}
		type row struct {
			CategoryID uint
			Cnt        int
		}
		var rows []row
		_ = s.db.Model(&models.ArticleModel{}).
			Select("category_id, count(*) as cnt").
			Where("category_id in ?", ids).
			Group("category_id").
			Scan(&rows).Error
		for _, r := range rows {
			countMap[r.CategoryID] = r.Cnt
		}
	}

	rep := &contentv1.ListCategoriesReply{Count: int64(count), Capped: capped}
	for _, it := range list {
		rep.List = append(rep.List, &contentv1.CategoryListItem{
			Category:     toProtoCategory(it),
			ArticleCount: int32(countMap[it.ID]),
			Nickname:     it.UserModel.NickName,
			Avatar:       it.UserModel.Avatar,
		})
	}
	return rep, nil
}

// RemoveCategories 删除分类;all=true 时不限归属(管理员)。
func (s *Server) RemoveCategories(_ context.Context, req *contentv1.RemoveCategoriesRequest) (*contentv1.RemoveCategoriesReply, error) {
	ids := make([]uint, 0, len(req.GetIds()))
	for _, id := range req.GetIds() {
		ids = append(ids, uint(id))
	}
	q := s.db.Where("id in ?", ids)
	if !req.GetAll() {
		q = q.Where("user_id = ?", uint(req.GetUserId()))
	}
	var list []models.CategoryModel
	if err := q.Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) > 0 {
		if err := s.db.Delete(&list).Error; err != nil {
			return nil, err
		}
	}
	return &contentv1.RemoveCategoriesReply{Deleted: int32(len(list))}, nil
}

// CategoryOptions 分类选项(id/title)。
func (s *Server) CategoryOptions(_ context.Context, req *contentv1.CategoryOptionsRequest) (*contentv1.CategoryOptionsReply, error) {
	var rows []struct {
		Value uint
		Label string
	}
	if err := s.db.Model(models.CategoryModel{}).Where("user_id = ?", uint(req.GetUserId())).
		Select("id as value", "title as label").Scan(&rows).Error; err != nil {
		return nil, err
	}
	rep := &contentv1.CategoryOptionsReply{}
	for _, r := range rows {
		rep.Options = append(rep.Options, &contentv1.CategoryOption{Value: uint64(r.Value), Label: r.Label})
	}
	return rep, nil
}
