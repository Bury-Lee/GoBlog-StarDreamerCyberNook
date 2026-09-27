package internal

import (
	"context"
	"encoding/json"
	"fmt"

	searchv1 "StarDreamerCyberNook/gen/search/v1"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/pkg/dbx"

	"github.com/olivere/elastic/v7"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// Config 是 search 服务的配置:ES 连接 + 数据库(降级用)。
type Config struct {
	URL      string
	UserName string
	Password string
	Driver   string // mysql | postgres | sqlite(降级搜索用)
	DSN      string
	Standalone bool // true: 自建独立库连接池;false: 复用宿主 global.DB
}

// Server 实现 SearchService,持有 Elasticsearch 客户端与降级数据库。
type Server struct {
	searchv1.UnimplementedSearchServiceServer
	cli *elastic.Client
	db  *gorm.DB
}

// NewServer 按配置连接 ES(可选)与数据库(可选)。
// URL 为空则不启用 ES;数据库句柄优先复用宿主 global.DB,开启 dbStandalone 时自建。两者至少一个。
func NewServer(cfg Config) (*Server, error) {
	s := &Server{}
	if cfg.URL != "" {
		cli, err := elastic.NewClient(
			elastic.SetURL(cfg.URL),
			elastic.SetSniff(false),
			elastic.SetHealthcheck(false), // ES 不可达时仍允许启动,查询阶段失败后回退 DB
			elastic.SetBasicAuth(cfg.UserName, cfg.Password),
		)
		if err != nil {
			return nil, err
		}
		s.cli = cli
	}
	if cfg.Standalone {
		db, err := dbx.Open(cfg.Driver, cfg.DSN)
		if err != nil {
			return nil, fmt.Errorf("search: open db: %w", err)
		}
		s.db = db
	} else {
		s.db = global.DB
	}
	if s.cli == nil && s.db == nil {
		return nil, fmt.Errorf("search: neither es url nor db dsn configured")
	}
	return s, nil
}

func (s *Server) exists(ctx context.Context, index string) (bool, error) {
	return s.cli.IndexExists(index).Do(ctx)
}

// EnsureIndex 确保索引存在。
func (s *Server) EnsureIndex(ctx context.Context, req *searchv1.EnsureIndexRequest) (*searchv1.EnsureIndexReply, error) {
	if s.cli == nil {
		return nil, status.Error(codes.Unavailable, "search: es not enabled")
	}
	index := req.GetIndex()
	ok, err := s.exists(ctx, index)
	if err != nil {
		return nil, err
	}
	if req.GetRecreate() && ok {
		if _, err := s.cli.DeleteIndex(index).Do(ctx); err != nil {
			return nil, err
		}
		ok = false
	}
	if !ok {
		if _, err := s.cli.CreateIndex(index).BodyString(req.GetMapping()).Do(ctx); err != nil {
			return nil, err
		}
	}
	return &searchv1.EnsureIndexReply{}, nil
}

// ExistsIndex 判断索引是否存在。
func (s *Server) ExistsIndex(ctx context.Context, req *searchv1.ExistsIndexRequest) (*searchv1.ExistsIndexReply, error) {
	if s.cli == nil {
		return nil, status.Error(codes.Unavailable, "search: es not enabled")
	}
	ok, err := s.exists(ctx, req.GetIndex())
	if err != nil {
		return nil, err
	}
	return &searchv1.ExistsIndexReply{Exists: ok}, nil
}

// DeleteIndex 删除索引。
func (s *Server) DeleteIndex(ctx context.Context, req *searchv1.DeleteIndexRequest) (*searchv1.DeleteIndexReply, error) {
	if s.cli == nil {
		return nil, status.Error(codes.Unavailable, "search: es not enabled")
	}
	if _, err := s.cli.DeleteIndex(req.GetIndex()).Do(ctx); err != nil {
		return nil, err
	}
	return &searchv1.DeleteIndexReply{}, nil
}

// SearchArticles 文章检索:优先 ES,失败时自动回退数据库全文检索表。
func (s *Server) SearchArticles(ctx context.Context, req *searchv1.SearchArticlesRequest) (*searchv1.SearchArticlesReply, error) {
	if s.cli != nil {
		rep, err := s.searchES(ctx, req)
		if err == nil {
			return rep, nil
		}
		if s.db == nil {
			return nil, err
		}
		// ES 不可用 → 回退数据库
	}
	if s.db != nil {
		return s.searchDB(req)
	}
	return nil, status.Error(codes.Unavailable, "search: no backend available")
}

// searchES 使用 Elasticsearch 检索(关键词/标签/排序/置顶加权/兴趣标签 + 高亮)。
func (s *Server) searchES(ctx context.Context, req *searchv1.SearchArticlesRequest) (*searchv1.SearchArticlesReply, error) {
	sortMap := map[int32]string{
		0: "created_at",
		1: "_score",
		2: "comment_count",
		3: "digg_count",
		4: "collect_count",
	}
	sortKey, ok := sortMap[req.GetType()]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "搜索类型错误")
	}

	query := elastic.NewBoolQuery()
	if req.GetKey() != "" {
		query.Should(
			elastic.NewMatchQuery("title", req.GetKey()),
			elastic.NewMatchQuery("abstract", req.GetKey()),
			elastic.NewMatchQuery("content", req.GetKey()),
		)
	}
	if req.GetTag() != "" {
		query.Must(elastic.NewTermQuery("tag_list", req.GetTag()))
	}
	query.Must(elastic.NewTermQuery("status", int(req.GetStatus())))
	if len(req.GetTopIds()) > 0 {
		any := make([]interface{}, 0, len(req.GetTopIds()))
		for _, id := range req.GetTopIds() {
			any = append(any, id)
		}
		query.Should(elastic.NewTermsQuery("id", any...).Boost(10))
	}
	if req.GetType() == 1 && len(req.GetLikeTags()) > 0 {
		tagAny := make([]interface{}, 0, len(req.GetLikeTags()))
		for _, t := range req.GetLikeTags() {
			tagAny = append(tagAny, t)
		}
		tagQuery := elastic.NewBoolQuery()
		tagQuery.Should(elastic.NewTermsQuery("tag_list", tagAny...)).MinimumNumberShouldMatch(1)
		query.Must(tagQuery)
	}

	highlight := elastic.NewHighlight()
	highlight.Field("title")
	highlight.Field("abstract")
	highlight.Field("content")

	result, err := s.cli.Search(req.GetIndex()).
		Query(query).
		Highlight(highlight).
		From(int(req.GetOffset())).
		Size(int(req.GetSize())).
		Sort(sortKey, false).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	if result == nil || result.Hits == nil {
		return nil, status.Error(codes.Internal, "search: empty es result")
	}

	rep := &searchv1.SearchArticlesReply{Total: result.Hits.TotalHits.Value}
	seen := make(map[uint64]struct{})
	for _, hit := range result.Hits.Hits {
		var art struct {
			ID       uint64 `json:"id"`
			Title    string `json:"title"`
			Abstract string `json:"abstract"`
		}
		if err := json.Unmarshal(hit.Source, &art); err != nil {
			continue
		}
		if len(hit.Highlight["title"]) > 0 {
			art.Title = hit.Highlight["title"][0]
		}
		if len(hit.Highlight["abstract"]) > 0 {
			art.Abstract = hit.Highlight["abstract"][0]
		}
		if _, ok := seen[art.ID]; ok {
			continue
		}
		seen[art.ID] = struct{}{}
		rep.Hits = append(rep.Hits, &searchv1.SearchHit{Id: art.ID, Title: art.Title, Abstract: art.Abstract})
	}
	return rep, nil
}

// searchDB 数据库全文检索降级:搜索表关键词匹配 + 文章表状态/标签过滤与排序。
func (s *Server) searchDB(req *searchv1.SearchArticlesRequest) (*searchv1.SearchArticlesReply, error) {
	build := func() *gorm.DB {
		q := s.db.Table("article_search_models AS s").
			Joins("JOIN article_models AS a ON a.id = s.article_id").
			Where("a.status = ?", req.GetStatus())
		if req.GetKey() != "" {
			like := "%" + req.GetKey() + "%"
			q = q.Where("(s.title LIKE ? OR s.abstract LIKE ?)", like, like)
		}
		if req.GetTag() != "" {
			q = q.Where("a.tag_list LIKE ?", "%\""+req.GetTag()+"\"%")
		}
		return q
	}
	var total int64
	if err := build().Count(&total).Error; err != nil {
		return nil, err
	}
	rep := &searchv1.SearchArticlesReply{Total: total}
	if total == 0 {
		return rep, nil
	}
	orderColumns := map[int32]string{0: "created_at", 1: "created_at", 2: "comment_count", 3: "digg_count", 4: "collect_count"}
	orderColumn := orderColumns[req.GetType()]
	if orderColumn == "" {
		orderColumn = "created_at"
	}
	type hit struct {
		ArticleID uint
		Title     string
		Abstract  string
	}
	var hits []hit
	if err := build().Select("s.article_id, s.title, s.abstract").
		Order("a." + orderColumn + " DESC").
		Order("s.article_id DESC").
		Offset(int(req.GetOffset())).Limit(int(req.GetSize())).
		Scan(&hits).Error; err != nil {
		return nil, err
	}
	for _, h := range hits {
		rep.Hits = append(rep.Hits, &searchv1.SearchHit{Id: uint64(h.ArticleID), Title: h.Title, Abstract: h.Abstract})
	}
	return rep, nil
}
