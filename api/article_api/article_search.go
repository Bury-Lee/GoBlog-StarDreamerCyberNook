package article_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	jwts "StarDreamerCyberNook/utils/jwts"
	"StarDreamerCyberNook/utils/sql"
	"context"
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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

// ArticleSearchView 搜索文章的API处理函数
// 主要逻辑：
// 1. 解析请求参数（分页、关键词、标签、排序类型）
// 2. 构建Elasticsearch查询DSL
// 3. 执行搜索并获取结果（含高亮）
// 4. 获取对应的完整文章数据（从数据库）
// 5. 合并数据并返回搜索结果
func (ArticleApi) ArticleSearchView(c *gin.Context) {
	// 1. 解析并验证请求参数
	var req ArticleSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数绑定失败", c)
		return
	}

	// 2. 根据请求的Type确定Elasticsearch排序字段
	var esSortMap = map[int8]string{
		0: "created_at",    // 最新发布：按创建时间排序
		1: "_score",        // 猜你喜欢：按相关性评分排序
		2: "comment_count", // 最多回复：按评论数排序
		3: "digg_count",    // 最多点赞：按点赞数排序
		4: "collect_count", // 最多收藏：按收藏数排序
	}
	var dbSortMap = map[int8]string{
		0: "created_at",    // 最新发布
		1: "created_at",    // 降级时没有ES评分，按发布时间兜底
		2: "comment_count", // 最多回复
		3: "digg_count",    // 最多点赞
		4: "collect_count", // 最多收藏
	}
	sortKey, ok := esSortMap[req.Type]
	if !ok { // 如果传入的Type不在map中，则返回错误
		response.FailWithMsg("搜索类型错误", c)
		return
	}
	dbSortKey := dbSortMap[req.Type]

	articleTopMap, topArticleIDList := getAdminTopArticleInfo()

	// ES不可用时，使用数据库全文搜索表降级
	if global.ES == nil {
		list, count, err := articleSearchFallback(c, req, dbSortKey, articleTopMap)
		if err != nil {
			logrus.Errorf("降级搜索失败 %s", err)
			response.FailWithMsg("查询失败", c)
			return
		}
		response.OkWithList(list, count, c)
		return
	}

	// 构建Elasticsearch Bool Query
	query := elastic.NewBoolQuery()
	// 如果有关键词，则在标题、摘要、内容中进行模糊匹配（Should关系）
	if req.Key != "" {
		query.Should(
			elastic.NewMatchQuery("title", req.Key),
			elastic.NewMatchQuery("abstract", req.Key),
			elastic.NewMatchQuery("content", req.Key),
		)
	}
	// 如果指定了标签，则必须匹配该标签（Must关系）
	if req.Tag != "" {
		query.Must(
			elastic.NewTermQuery("tag_list", req.Tag),
		)
	}

	// 只查询已发布的文章（Must关系）
	query.Must(elastic.NewTermQuery("status", int(models.StatusPublished)))

	// 处理管理员置顶逻辑
	var articleIDList []uint // 用于存储最终要返回的文章ID列表
	if len(topArticleIDList) > 0 {
		var topArticleIDListAny []interface{}
		for _, u := range topArticleIDList {
			topArticleIDListAny = append(topArticleIDListAny, u)
		}
		// 只给命中的置顶文章加权，不绕过搜索条件
		query.Should(elastic.NewTermsQuery("id", topArticleIDListAny...).Boost(10))
	}

	// 如果是"猜你喜欢"（Type=1），则加入用户兴趣标签查询
	if req.Type == 1 {
		// 尝试从JWT Token中解析用户信息
		likeTags, err := getCurrentUserLikeTags(c)
		if err != nil {
			response.FailWithMsg("用户信息不存在", c)
			return
		}
		// 如果用户配置中有感兴趣的文章标签
		if len(likeTags) > 0 {
			tagQuery := elastic.NewBoolQuery()
			var tagAnyList []interface{}
			for _, tag := range likeTags {
				tagAnyList = append(tagAnyList, tag)
			}
			// 兴趣标签至少命中一个
			tagQuery.Should(elastic.NewTermsQuery("tag_list", tagAnyList...)).
				MinimumNumberShouldMatch(1)
			query.Must(tagQuery)
		}
	}

	// 设置高亮显示，对标题和摘要字段进行高亮
	highlight := elastic.NewHighlight()
	highlight.Field("title")
	highlight.Field("abstract")

	// 执行Elasticsearch搜索
	result, err := global.ES.
		Search(models.ArticleModel{}.Index()). // 指定搜索的索引
		Query(query).                          // 设置查询DSL
		Highlight(highlight).                  // 设置高亮
		From(req.GetOffset()).                 // 设置分页偏移量
		Size(req.GetLimit()).                  // 设置分页大小
		Sort(sortKey, false).                  // 设置排序字段和方向(false表示降序)
		Do(context.Background())               // 执行搜索
	if err != nil {
		// 记录错误日志，包括错误信息和查询语句
		source, _ := query.Source()
		byteData, _ := json.Marshal(source)
		logrus.Errorf("查询失败 %s \n %s", err, string(byteData))
		response.FailWithMsg("查询失败", c)
		return
	}

	// 解析Elasticsearch返回的命中结果
	count := result.Hits.TotalHits.Value              // 获取总命中数
	var searchArticleMap = map[uint]ArticleBaseInfo{} // 用于存储ES返回的精简文章信息（含高亮）
	var articleIDSet = map[uint]struct{}{}            // 用于去重文章ID

	for _, hit := range result.Hits.Hits {
		var art ArticleBaseInfo
		// 将ES返回的JSON Source反序列化为ArticleBaseInfo结构体
		err = json.Unmarshal(hit.Source, &art)
		if err != nil {
			logrus.Warnf("解析失败 %s  %s", err, string(hit.Source))
			continue
		}
		// 如果有高亮结果，则替换原始内容
		if len(hit.Highlight["title"]) > 0 {
			art.Title = hit.Highlight["title"][0]
		}
		if len(hit.Highlight["abstract"]) > 0 {
			art.Abstract = hit.Highlight["abstract"][0]
		}

		searchArticleMap[art.ID] = art // 存入映射表
		if _, ok := articleIDSet[art.ID]; ok {
			continue
		}
		articleIDSet[art.ID] = struct{}{}
		articleIDList = append(articleIDList, art.ID)
	}

	// 没有命中时直接返回，避免无意义的数据库查询
	if len(articleIDList) == 0 {
		response.OkWithList([]ArticleSearchListResponse{}, int(count), c)
		return
	}

	list, err := getArticleSearchList(articleIDList, articleTopMap, searchArticleMap)
	if err != nil {
		logrus.Errorf("查询文章详情失败 %s", err)
		response.FailWithMsg("查询失败", c)
		return
	}

	//TODO:以后加入带点赞数和评论数的响应字段
	// 返回成功响应，包含文章列表和总数
	response.OkWithList(list, int(count), c)
}

// getAdminTopArticleInfo 获取管理员置顶文章信息
// 返回:articleTopMap - 置顶文章映射
// 返回:topArticleIDList - 置顶文章ID列表
// 说明:先查管理员ID，再查管理员置顶文章
func getAdminTopArticleInfo() (articleTopMap map[uint]bool, topArticleIDList []uint) {
	var userIDList []uint
	articleTopMap = map[uint]bool{}

	global.DB.Model(models.UserModel{}).Where("role = ?", enum.AdminRole).Select("id").Scan(&userIDList)
	if len(userIDList) == 0 {
		return articleTopMap, topArticleIDList
	}

	global.DB.Model(models.UserTopArticleModel{}).Where("user_id in ?", userIDList).Select("article_id").Scan(&topArticleIDList)
	for _, articleID := range topArticleIDList {
		articleTopMap[articleID] = true
	}
	return articleTopMap, topArticleIDList
}

// getCurrentUserLikeTags 获取当前用户兴趣标签
// 参数:c - gin上下文
// 返回:likeTags - 兴趣标签列表
// 返回:err - 错误信息
// 说明:未登录返回空切片，已登录时读取用户like_tags
func getCurrentUserLikeTags(c *gin.Context) (likeTags []string, err error) {
	claims, err := jwts.ParseTokenByGin(c)
	if err != nil || claims == nil {
		return nil, nil
	}

	var user models.UserModel
	err = global.DB.Select("id", "like_tags").Take(&user, claims.UserID).Error
	if err != nil {
		return nil, err
	}
	return user.LikeTags, nil
}

// buildJSONTagLikeQuery 构建JSON标签模糊查询
// 参数:column - 标签字段名
// 参数:tags - 标签列表
// 返回:query - GORM查询条件
// 说明:JSON序列化后按带引号的标签匹配，避免子串误匹配
func buildJSONTagLikeQuery(column string, tags []string) *gorm.DB {
	query := global.DB.Session(&gorm.Session{NewDB: true})
	for index, tag := range tags {
		likeValue := "%\"" + tag + "\"%"
		if index == 0 {
			query = query.Where(column+" LIKE ?", likeValue)
			continue
		}
		query = query.Or(column+" LIKE ?", likeValue)
	}
	return query
}

// articleSearchFallback 降级搜索文章
// 参数:c - gin上下文
// 参数:req - 搜索请求
// 参数:sortKey - 排序字段
// 参数:articleTopMap - 置顶文章映射
// 返回:list - 搜索结果列表
// 返回:count - 总数
// 返回:err - 错误信息
// 说明:用ArticleSearchModel定位文章ID，优先读Redis，未命中再查数据库
func articleSearchFallback(c *gin.Context, req ArticleSearchRequest, sortKey string, articleTopMap map[uint]bool) (list []ArticleSearchListResponse, count int, err error) {
	query := global.DB.Model(&models.ArticleModel{}).Where("status = ?", models.StatusPublished)

	if req.Key != "" {
		likeValue := "%" + req.Key + "%"
		subQuery := global.DB.Model(&models.ArticleSearchModel{}).
			Select("DISTINCT article_id").
			Where(global.DB.Where("title LIKE ?", likeValue).Or("abstract LIKE ?", likeValue))
		query = query.Where("id IN (?)", subQuery)
	}

	if req.Tag != "" {
		query = query.Where("tag_list LIKE ?", "%\""+req.Tag+"\"%")
	}

	if req.Type == 1 {
		likeTags, likeErr := getCurrentUserLikeTags(c)
		if likeErr != nil {
			return nil, 0, likeErr
		}
		if len(likeTags) > 0 {
			query = query.Where(buildJSONTagLikeQuery("tag_list", likeTags))
		}
	}

	var total int64
	if err = query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []ArticleSearchListResponse{}, 0, nil
	}

	var articleBaseList []ArticleBaseInfo
	err = query.Select("id", "title", "abstract").
		Order(sortKey + " DESC").
		Order("id DESC").
		Offset(req.GetOffset()).
		Limit(req.GetLimit()).
		Find(&articleBaseList).Error
	if err != nil {
		return nil, 0, err
	}

	if len(articleBaseList) == 0 {
		return []ArticleSearchListResponse{}, int(total), nil
	}

	var articleIDList []uint
	searchArticleMap := map[uint]ArticleBaseInfo{}
	for _, article := range articleBaseList {
		articleIDList = append(articleIDList, article.ID)
		searchArticleMap[article.ID] = article
	}

	list, err = getArticleSearchList(articleIDList, articleTopMap, searchArticleMap)
	if err != nil {
		return nil, 0, err
	}
	return list, int(total), nil
}

// getArticleSearchList 获取搜索结果详情
// 参数:articleIDList - 文章ID列表
// 参数:articleTopMap - 置顶文章映射
// 参数:searchArticleMap - 搜索命中的标题摘要映射
// 返回:list - 搜索结果列表
// 返回:err - 错误信息
// 说明:优先从Redis读取文章详情，未命中再查询数据库，并按原始顺序组装
func getArticleSearchList(articleIDList []uint, articleTopMap map[uint]bool, searchArticleMap map[uint]ArticleBaseInfo) (list []ArticleSearchListResponse, err error) {
	keyList := []string{}
	for _, id := range articleIDList {
		idStr := strconv.FormatUint(uint64(id), 10)
		keyList = append(keyList, "ArticleID"+idStr)
	}

	ctx := context.Background()
	res, cacheErr := global.RedisHotPool.MGet(ctx, keyList...).Result()

	var cacheMissIDList []uint
	cacheHitMap := make(map[uint]ArticleSearchListResponse)

	if cacheErr == nil {
		for index, cacheData := range res {
			if cacheData == nil {
				cacheMissIDList = append(cacheMissIDList, articleIDList[index])
				continue
			}

			var cached ArticleDetailResponse
			err = json.Unmarshal([]byte(cacheData.(string)), &cached)
			if err != nil {
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
		where := global.DB.Where("id in ?", cacheMissIDList)
		modelList, _, err := common.ListQuery(models.ArticleModel{}, common.Options{
			Where:        where,
			Preloads:     []string{"CategoryModel", "UserModel"},
			DefaultOrder: sql.ConvertSliceOrderSql(cacheMissIDList),
		})
		if err != nil {
			return nil, err
		}

		for _, model := range modelList {
			item := ArticleSearchListResponse{
				ArticleModel: model,
				AdminTop:     articleTopMap[model.ID],
				UserNickname: model.UserModel.NickName,
				UserAvatar:   model.UserModel.Avatar,
			}
			if model.CategoryModel != nil {
				item.CategoryTitle = &model.CategoryModel.Title
			}
			if article, ok := searchArticleMap[model.ID]; ok {
				item.Title = article.Title
				item.Abstract = article.Abstract
			}
			cacheHitMap[model.ID] = item
		}
	}

	list = make([]ArticleSearchListResponse, 0, len(articleIDList))
	for _, id := range articleIDList {
		if item, ok := cacheHitMap[id]; ok {
			list = append(list, item)
		}
	}
	return list, nil
}
