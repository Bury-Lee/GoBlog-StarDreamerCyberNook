package article_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/content_service"
	"StarDreamerCyberNook/service/es_service"
	"StarDreamerCyberNook/service/user_service"
	"StarDreamerCyberNook/utils"
	jwts "StarDreamerCyberNook/utils/jwts"
	"context"
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ArticleSearchRequest 搜索请求结构体，包含分页信息、标签和排序类型
type ArticleSearchRequest struct {
	common.PageInfo
	Tag  string `form:"tag"`  // 按标签筛选
	Type int8   `form:"type"` // 排序类型: 0 最新发布 1 猜你喜欢    2最多回复 3最多点赞 4最多收藏
}

// ArticleBaseInfo 搜索结果的基础信息结构体
type ArticleBaseInfo struct {
	ID       uint   `json:"id"`
	Title    string `json:"title"`
	Abstract string `json:"abstract"`
}

// ArticleSearchListResponse 搜索结果详情结构体，继承了文章模型，并增加了关联信息
type ArticleSearchListResponse struct {
	models.ArticleModel
	AdminTop      bool    `json:"adminTop"`      // 是否是管理员置顶
	CategoryTitle *string `json:"categoryTitle"` // 所属分类标题
	UserNickname  string  `json:"userNickname"`  // 发布用户昵称
	UserAvatar    string  `json:"userAvatar"`    // 发布用户头像
}

// ArticleSearchView 文章检索:检索本身(ES 或 DB 降级)由 search 服务负责,网关只编排缓存与详情补齐。
func (ArticleApi) ArticleSearchView(c *gin.Context) {
	// 1. 解析并验证请求参数
	var req ArticleSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数绑定失败", c)
		return
	}

	// 2. 根据请求的Type确定搜索排序(合法性由 search 服务最终校验)
	var esSortMap = map[int8]string{
		0: "created_at",    // 最新发布：按创建时间排序
		1: "_score",        // 猜你喜欢：按相关性评分排序
		2: "comment_count", // 最多回复：按评论数排序
		3: "digg_count",    // 最多点赞：按点赞数排序
		4: "collect_count", // 最多收藏：按收藏数排序
	}
	if _, ok := esSortMap[req.Type]; !ok {
		response.FailWithMsg("搜索类型错误", c)
		return
	}

	// 获取管理员置顶文章信息(下沉 content 服务)
	articleTopMap := map[uint]bool{}
	var topArticleIDList []uint
	{
		adminTopIDs, err := content_service.AdminTopArticleIDs()
		if err != nil {
			response.FailWithMsg("查询失败", c)
			return
		}
		topArticleIDList = adminTopIDs
		for _, articleID := range topArticleIDList {
			articleTopMap[articleID] = true
		}
	}

	// "猜你喜欢"(type=1)携带兴趣标签,供 search 服务加权
	likeTags := []string{}
	if req.Type == 1 {
		if claims, err := jwts.ParseTokenByGin(c); err == nil && claims != nil {
			detail, derr := user_service.GetUserDetail(claims.UserID)
			if derr != nil {
				response.FailWithMsg("用户信息不存在", c)
				return
			}
			likeTags = detail.LikeTags
		}
	}

	// 经 search 微服务执行检索(ES 或 DB 降级由服务内部决定)
	count, hits, serr := es_service.SearchArticles(
		models.ArticleModel{}.Index(), req.Key, req.Tag, req.Type,
		req.GetOffset(), req.GetLimit(), int(models.StatusPublished), topArticleIDList, likeTags,
	)
	if serr != nil {
		logrus.Errorf("search 服务检索失败 %s", serr)
		response.FailWithMsg("查询失败", c)
		return
	}

	searchArticleMap := map[uint]ArticleBaseInfo{}
	var articleIDList []uint
	for _, h := range hits {
		art := ArticleBaseInfo{ID: uint(h.ID), Title: h.Title, Abstract: h.Abstract}
		searchArticleMap[art.ID] = art
		articleIDList = append(articleIDList, art.ID)
	}

	// 没有命中时直接返回，避免无意义的数据库查询
	if len(articleIDList) == 0 {
		response.OkWithList([]ArticleSearchListResponse{}, int(count), c)
		return
	}

	list, cerr := buildSearchResults(articleIDList, searchArticleMap, articleTopMap, req.Key)
	if cerr != nil {
		logrus.Errorf("查询文章详情失败 %s", cerr)
		response.FailWithMsg("查询失败", c)
		return
	}
	applySearchCountDeltas(list)

	// 返回成功响应，包含文章列表和总数
	response.OkWithList(list, int(count), c)
}

// buildSearchResults 用 Redis 详情缓存 + content 服务补齐搜索结果(缓存未命中批量查 content)。
// key 用于关键词高亮。
func buildSearchResults(articleIDList []uint, searchArticleMap map[uint]ArticleBaseInfo,
	articleTopMap map[uint]bool, key string) ([]ArticleSearchListResponse, error) {

	keyList := []string{}
	for _, id := range articleIDList {
		keyList = append(keyList, "ArticleID"+strconv.FormatUint(uint64(id), 10))
	}

	ctx := context.Background()
	res, cacheErr := global.RedisHotPool.MGet(ctx, keyList...).Result()

	var cacheMissIDList []uint
	cacheHitMap := make(map[uint]ArticleSearchListResponse)

	if cacheErr == nil {
		for index, cacheData := range res {
			cacheStr, ok := cacheData.(string)
			if !ok || cacheStr == articleNotFoundCache {
				cacheMissIDList = append(cacheMissIDList, articleIDList[index])
				continue
			}
			var cached ArticleDetailResponse
			if err := json.Unmarshal([]byte(cacheStr), &cached); err != nil {
				logrus.Warnf("缓存数据解析错误: %s", err)
				cacheMissIDList = append(cacheMissIDList, articleIDList[index])
				continue
			}
			item := ArticleSearchListResponse{
				ArticleModel:  cached.ArticleModel,
				AdminTop:      articleTopMap[cached.ID],
				CategoryTitle: cached.CategoryTitle,
				UserNickname:  cached.NickName,
				UserAvatar:    cached.UserAvatar,
			}
			if article, ok := searchArticleMap[cached.ID]; ok {
				item.Title = article.Title
				item.Abstract = article.Abstract
			}
			cacheHitMap[cached.ID] = item
		}
	} else {
		cacheMissIDList = articleIDList
	}

	if len(cacheMissIDList) > 0 {
		details, derr := content_service.GetArticlesByIDs(cacheMissIDList, false)
		if derr != nil {
			return nil, derr
		}
		for _, d := range details {
			item := ArticleSearchListResponse{
				ArticleModel:  d.ArticleModel,
				AdminTop:      articleTopMap[d.ArticleModel.ID],
				CategoryTitle: d.CategoryTitle,
				UserNickname:  d.NickName,
				UserAvatar:    d.Avatar,
			}
			if article, ok := searchArticleMap[d.ArticleModel.ID]; ok {
				item.Title = article.Title
				item.Abstract = article.Abstract
			}
			cacheHitMap[d.ArticleModel.ID] = item
		}
	}

	list := make([]ArticleSearchListResponse, 0, len(articleIDList))
	for _, id := range articleIDList {
		if item, ok := cacheHitMap[id]; ok {
			item.Abstract = utils.HighlightKeyword(item.Abstract, key)
			item.Title = utils.HighlightKeyword(item.Title, key)
			item.Content = utils.HighlightKeyword(item.Content, key)
			list = append(list, item)
		}
	}
	return list, nil
}
