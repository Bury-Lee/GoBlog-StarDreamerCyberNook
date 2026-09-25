package article_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	jwts "StarDreamerCyberNook/utils/jwts"
	"StarDreamerCyberNook/utils/sql"
	"fmt"

	"github.com/gin-gonic/gin"
)

//TODO:写好注释到时候好好看看

type ArticleListRequest struct {
	common.PageInfo
	Type       string         `form:"type" binding:"required"`
	UserID     uint           `form:"userID"`
	CategoryID *uint          `form:"categoryID"`
	Status     *models.Status `form:"status"` // 用指针区分"不筛选"和"筛选草稿(status=0)"
}

type ArticleListResponse struct {
	models.ArticleModel
	UserTop       bool    `json:"userTop"`
	AdminTop      bool    `json:"adminTop"`
	CategoryTitle *string `json:"categoryTitle"`
	UserNickName  string  `json:"userNickName"`
	Avatar        string  `json:"avatar"`
}

func (ArticleApi) ArticleListView(c *gin.Context) {
	var req ArticleListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	var TopArticleIDList []uint

	var orderColumnMap = map[string]bool{
		"look_count desc":    true,
		"digg_count desc":    true,
		"comment_count desc": true,
		"collect_count desc": true,
		"look_count asc":     true,
		"digg_count asc":     true,
		"comment_count asc":  true,
		"collect_count asc":  true,
	}

	if req.Order != "" {
		_, ok := orderColumnMap[req.Order]
		if !ok {
			response.FailWithMsg("不支持的排序方式", c)
			return
		}
	}

	switch req.Type {
	case "other":
		// 查别人,用户id就是必填的
		// if req.UserID == 0 {
		// 	response.FailWithMsg("用户id是必填项", c)
		// 	return
		// }
		//啊算了,去除这个限制来支持查询最新文章

		// if req.Page > 2 || req.Limit > 10 {
		// 	response.FailWithMsg("查询更多，请登录", c)
		// 	return
		// }
		published := models.StatusPublished
		req.Status = &published // 查别人只能查已发布的
	case "self":
		// 查自己的
		claims, err := jwts.ParseTokenByGin(c)
		if err != nil || claims.UserID == 0 {
			response.FailWithMsg("请登录", c)
			return
		}
		req.UserID = claims.UserID
	case "admin":
		// 管理员
		claims, err := jwts.ParseTokenByGin(c)
		if err != nil || claims.Role != enum.AdminRole {
			response.FailWithMsg("角色错误", c)
			return
		}
	default:
		response.FailWithMsg("请求错误", c)
		return
	}
	var userTopMap = make(map[uint]bool)
	var adminTopMap = make(map[uint]bool)
	if req.UserID != 0 { // 查询用户置顶文章
		var userTopArticleList []models.UserTopArticleModel
		global.DB.Preload("UserModel").Order("created_at desc").Find(&userTopArticleList, "user_id = ?", req.UserID)
		for _, item := range userTopArticleList {
			TopArticleIDList = append(TopArticleIDList, item.ArticleID)
			if item.UserModel.Role == enum.AdminRole {
				adminTopMap[item.ArticleID] = true
			}
			userTopMap[item.ArticleID] = true
		}
	}

	var options = common.Options{
		Likes:         []string{"title"},
		PageInfo:      req.PageInfo,
		Preloads:      []string{"UserModel", "CategoryModel"}, //预加载用户和分类
		DefaultOrder:  "created_at desc",
		AllowedOrders: []string{"created_at", "look_count", "digg_count", "comment_count", "collect_count"},
		CountCap:      common.DefaultCountCap, //总数封顶
	}
	if len(TopArticleIDList) > 0 {
		options.DefaultOrder = fmt.Sprintf("%s, created_at desc", sql.ConvertSliceOrderSql(TopArticleIDList))
	}
	//状态必须用 Where 显式查询:结构体查询会忽略零值,草稿(status=0)会被当作"不筛选"
	if req.Status != nil {
		options.Where = global.DB.Where("status = ?", *req.Status)
	}

	_list, count, capped, _ := common.ListQuery(models.ArticleModel{
		UserID:     req.UserID,
		CategoryID: req.CategoryID,
	}, options)

	var list = make([]ArticleListResponse, 0)

	for _, model := range _list {
		data := ArticleListResponse{
			ArticleModel: model,
			UserTop:      userTopMap[model.ID],
			AdminTop:     adminTopMap[model.ID],
			UserNickName: model.UserModel.NickName,
			Avatar:       model.UserModel.Avatar,
		}
		if model.CategoryID != nil && model.CategoryModel != nil {
			data.CategoryTitle = &model.CategoryModel.Title
		}
		list = append(list, data)
	}

	//叠加Redis中未同步的计数增量,避免点赞/收藏/评论后要等定时任务回写才看到变化
	ptrs := make([]*models.ArticleModel, 0, len(list))
	for i := range list {
		ptrs = append(ptrs, &list[i].ArticleModel)
	}
	applyArticleCountDeltas(ptrs)

	response.OkWithListCapped(list, count, capped, c)
}
