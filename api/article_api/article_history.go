package article_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/content_service"
	"StarDreamerCyberNook/service/redis_service/redis_count"
	jwts "StarDreamerCyberNook/utils/jwts"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type ArticleLookRequest struct {
	ArticleID  uint `json:"articleID" binding:"required"`
	TimeSecond int  `json:"timeSecond"` // 读文章一共用了多久,预备字段,也许以后可以用于做喜好分析
}

// 创建历史记录
func (ArticleApi) ArticleLookView(c *gin.Context) {
	//这里浏览量的增加任务已经交给Redis了,由Redis在后台记录增量并定期刷新,由于浏览量的增加不需要特别强的一致性,所以在这里直接返回成功,真正的增加浏览量的任务交给Redis去做,这样可以大大提高接口的响应速度,并且可以防止刷浏览量的攻击,因为Redis天然支持去重,所以同一个用户在同一天内多次请求这个接口,只有第一次会增加浏览量,后续的请求都会被Redis拦截掉,不会增加浏览量
	var req ArticleLookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(err.Error(), c)
		return
	}
	claims, err := jwts.ParseTokenByGin(c)
	if err != nil {
		response.OkWithMsg("未登录", c) //未登录不给加浏览量
		return
	}

	// 引入缓存: 当天这个用户请求这个文章之后，将用户id和文章id作为key存入缓存，在这里进行判断，如果存在就直接返回
	if redis_count.GetUserArticleHistoryCache(req.ArticleID, claims.UserID) {
		response.OkWithMsg("成功", c)
		return
	}
	// 落库下沉 content 服务(文章存在性校验 + 当日去重)
	created, err := content_service.RecordArticleLook(claims.UserID, req.ArticleID)
	if err != nil {
		if errors.Is(err, content_service.ErrNotFound) {
			response.FailWithMsg("文章不存在", c)
			return
		}
		response.FailWithMsg("失败", c)
		return
	}

	//仅在真正新建历史记录时增加浏览量,并补写去重缓存
	if created {
		redis_count.SetCacheLook(req.ArticleID, true)
	}
	redis_count.SetUserArticleHistoryCache(req.ArticleID, claims.UserID)
	response.OkWithMsg("成功", c)
}

type ArticleLookListRequest struct {
	common.PageInfo
	UserID uint `form:"userID"` //为0时是查询自己的浏览记录,否则是指定用户的浏览记录
}

type ArticleLookListResponse struct {
	ID        uint      `json:"id"`       // 浏览记录的id
	LookDate  time.Time `json:"lookDate"` // 浏览的时间
	Title     string    `json:"title"`
	Cover     string    `json:"cover"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	UserID    uint      `json:"userID"`
	ArticleID uint      `json:"articleID"`
}

func (ArticleApi) ArticleLookListView(c *gin.Context) { //除了可以记录浏览量之外,还可以根据用户id查询用户的浏览记录
	var req ArticleLookListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg(err.Error(), c)
		return
	}

	claims, _ := jwts.ParseTokenByGin(c)
	viewerLogged := claims != nil
	var viewerID uint
	if claims != nil {
		viewerID = claims.UserID
	}
	if req.UserID == 0 {
		if !viewerLogged {
			response.FailWithMsg("未登录", c)
			return
		}
		req.UserID = viewerID
	}

	items, count, capped, err := content_service.ListArticleLook(
		req.UserID, viewerID, viewerLogged, req.Page, req.Limit, req.Order, req.EndId)
	if err != nil {
		switch {
		case errors.Is(err, content_service.ErrNotFound):
			response.FailWithMsg("用户不存在", c)
		case errors.Is(err, content_service.ErrPermissionDenied):
			response.FailWithMsg("用户未公开浏览记录", c)
		default:
			response.FailWithMsg("查询失败", c)
		}
		return
	}

	var list = make([]ArticleLookListResponse, 0, len(items))
	for _, it := range items {
		list = append(list, ArticleLookListResponse{
			ID:        it.ID,
			LookDate:  it.LookDate,
			Title:     it.Title,
			Cover:     it.Cover,
			Nickname:  it.Nickname,
			Avatar:    it.Avatar,
			UserID:    it.UserID,
			ArticleID: it.ArticleID,
		})
	}

	response.OkWithListCapped(list, count, capped, c)
}

func (ArticleApi) ArticleLookRemoveView(c *gin.Context) { //TODO:写一个定时任务?如果数据库里超过1个月的浏览记录,就删除
	var req models.RemoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(err.Error(), c)
		return
	}

	claims := jwts.GetClaims(c)
	deleted, err := content_service.RemoveArticleLook(claims.UserID, req.IDList)
	if err != nil {
		response.FailWithMsg("历史记录删除失败", c)
		return
	}

	response.OkWithMsg(fmt.Sprintf("删除历史记录成功 共删除%d条", deleted), c)
}
