package article_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/service/content_service"
	"StarDreamerCyberNook/service/message_service"
	"StarDreamerCyberNook/service/redis_service/redis_count"
	jwts "StarDreamerCyberNook/utils/jwts"
	utils_other "StarDreamerCyberNook/utils/other"
	"encoding/json"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ArticleCollectRequest struct {
	ArticleID uint `json:"articleID" binding:"required"`
	CollectID uint `json:"collectID"`
}

func (ArticleApi) ArticleCollectView(c *gin.Context) {
	var req ArticleCollectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	claims := jwts.GetClaims(c)

	// 收藏/取消/移动 的落库下沉 content 服务
	action, err := content_service.ToggleArticleCollect(claims.UserID, req.ArticleID, req.CollectID)
	if err != nil {
		switch {
		case errors.Is(err, content_service.ErrNotFound):
			response.FailWithMsg("文章不存在", c)
		case errors.Is(err, content_service.ErrFolderNotFound):
			response.FailWithMsg("收藏夹不存在", c)
		default:
			response.FailWithMsg("操作失败", c)
		}
		return
	}

	switch action {
	case "removed":
		response.OkWithMsg("取消收藏成功", c)
	case "moved":
		response.OkWithMsg("已移动到新的收藏夹", c)
	default: // created
		response.OkWithMsg("收藏成功", c)
		// 发送收藏消息,失败只记录日志,不影响主流程
		if err := message_service.InsertCollectMessage(models.UserArticleCollectModel{
			UserID:    claims.UserID,
			ArticleID: req.ArticleID,
		}); err != nil {
			logrus.Error("发送收藏消息失败", err.Error())
		}
	}
}

type CollectCreateRequest struct { //创建收藏夹请求参数,请求创建时不用传id参数,除了创建也可以用于更新收藏夹
	Title    string `json:"title" binding:"required,max=32" s:"title"`
	Abstract string `json:"abstract" s:"abstract"`
	Cover    string `json:"cover" s:"cover"`
}

func (ArticleApi) CollectCreateView(c *gin.Context) {
	var req CollectCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	claims := jwts.GetClaims(c)
	if err := content_service.CreateCollectFolder(claims.UserID, req.Title, req.Abstract, req.Cover); err != nil {
		response.FailWithMsg("创建收藏夹失败", c)
		return
	}
	response.OkWithMsg("创建收藏夹成功", c)
}

type CollectUpdateRequest struct { //更新收藏夹请求参数,仅更新传入的字段
	ID       uint    `json:"id" binding:"required"`  //更新时需要传id参数
	Title    *string `json:"title" s:"title"`        // 收藏夹名称
	Abstract *string `json:"abstract" s:"abstract"`  // 收藏夹简介
	Cover    *string `json:"cover" s:"cover"`        // 收藏夹封面
	IsPublic *bool   `json:"isPublic" s:"is_public"` // 是否公开该收藏夹
}

func (ArticleApi) CollectUpdateView(c *gin.Context) {
	var req CollectUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	claims := jwts.GetClaims(c)
	updateMap := utils_other.StructToMap(&req, "s") //把请求参数转换成map,只更新有值的字段
	if len(updateMap) == 0 {
		response.FailWithMsg("没有需要更新的字段", c)
		return
	}
	patch, err := json.Marshal(updateMap)
	if err != nil {
		response.FailWithMsg("更新收藏夹失败", c)
		return
	}

	if err := content_service.UpdateCollectFolder(claims.UserID, req.ID, string(patch)); err != nil {
		if errors.Is(err, content_service.ErrNotFound) {
			response.FailWithMsg("收藏夹不存在", c)
			return
		}
		response.FailWithMsg("更新收藏夹失败", c)
		return
	}

	response.OkWithMsg("更新收藏夹成功", c)
}

func (ArticleApi) CollectRemoveView(c *gin.Context) {
	var req models.RemoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	if len(req.IDList) == 0 {
		response.OkWithMsg("未找到可删除的收藏夹或无权限", c)
		return
	}

	claims := jwts.GetClaims(c)
	isAdmin := claims.Role == enum.AdminRole
	deltas, deleted, err := content_service.RemoveCollectFolders(claims.UserID, isAdmin, req.IDList)
	if err != nil {
		response.FailWithMsg("删除收藏夹失败: "+err.Error(), c)
		return
	}
	if deleted == 0 {
		response.OkWithMsg("未找到可删除的收藏夹或无权限", c)
		return
	}
	// 数据库提交成功后调整缓存计数
	for articleID, delta := range deltas {
		redis_count.SetCacheCollectBy(articleID, delta)
	}
	response.OkWithMsg("删除收藏夹成功", c)
}

type CollectListViewRequest struct {
	common.PageInfo
	ID uint `form:"id"`
}

// 先写着吧,一般来说是只有好友才能互看收藏夹的,或者以后给管理员看
func (ArticleApi) CollectListView(c *gin.Context) { //先看看用户有没有公开收藏夹,再查询用户收藏夹列表
	var req CollectListViewRequest

	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	if req.ID == 0 {
		response.FailWithMsg("参数错误", c)
		return
	}

	//收藏夹公开由文件夹级 is_public 控制(用户级 openCollect 已废弃),非本人只能看到公开的收藏夹
	claims, _ := jwts.ParseTokenByGin(c)
	includePrivate := claims != nil && claims.UserID == req.ID

	list, count, capped, err := content_service.ListCollectFolders(req.ID, includePrivate, req.Page, req.Limit, req.Key, req.Order, req.EndId)
	if err != nil {
		response.FailWithMsg("查询收藏夹列表失败", c)
		return
	}
	response.OkWithListCapped(list, count, capped, c)
}

type CollectArticleListViewRequest struct {
	common.PageInfo
	Likes []string `form:"likes"`
	ID    uint     `form:"id"`
}

func (ArticleApi) CollectArticleListView(c *gin.Context) {
	var req CollectArticleListViewRequest //查询用户的收藏夹文章列表
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	if req.ID == 0 {
		response.FailWithMsg("参数错误", c)
		return
	}

	//检查收藏夹所属/公开性(下沉 content 服务取详情,隐私判定在网关)
	folder, _, err := content_service.GetCollectFolder(req.ID)
	if err != nil {
		response.Fail(response.NotFound, "收藏夹不存在", response.EmptyData, c)
		return
	}
	claims, _ := jwts.ParseTokenByGin(c)
	if claims == nil || claims.UserID != folder.UserID {
		if !folder.IsPublic {
			response.Fail(response.NotFound, "收藏夹不存在", response.EmptyData, c)
			return
		}
	}

	data, count, capped, err := content_service.ListCollectArticles(req.ID, req.Page, req.Limit, req.Key, req.Order, req.EndId)
	if err != nil {
		response.FailWithMsg("查询收藏夹文章列表失败", c)
		return
	}

	//叠加Redis中未同步的计数增量,避免收藏/评论后要等定时任务回写才看到变化
	ptrs := make([]*models.ArticleModel, 0, len(data))
	for i := range data {
		ptrs = append(ptrs, &data[i])
	}
	applyArticleCountDeltas(ptrs)

	response.OkWithListCapped(data, count, capped, c)
}

// CollectDetailResponse 收藏夹详情响应
type CollectDetailResponse struct {
	models.CollectModel
	ArticleCount int64 `json:"articleCount"` // 收藏夹内文章数量
}

// CollectDetailView 获取单个收藏夹详情
func (ArticleApi) CollectDetailView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	folder, articleCount, err := content_service.GetCollectFolder(req.ID)
	if err != nil {
		response.Fail(response.NotFound, "收藏夹不存在", response.EmptyData, c)
		return
	}

	//非本人(含未登录)访问他人收藏夹时,仅公开收藏夹(is_public)可访问,否则统一按不存在处理,避免泄露隐私
	claims, _ := jwts.ParseTokenByGin(c)
	if claims == nil || claims.UserID != folder.UserID {
		if !folder.IsPublic {
			response.Fail(response.NotFound, "收藏夹不存在", response.EmptyData, c)
			return
		}
	}

	response.OkWithData(CollectDetailResponse{CollectModel: folder, ArticleCount: articleCount}, c)
}
