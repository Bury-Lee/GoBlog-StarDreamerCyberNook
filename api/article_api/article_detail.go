package article_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	jwts "StarDreamerCyberNook/utils/jwts"
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// articleNotFoundCache 文章不存在的负缓存标记,避免不存在的ID反复查库
const articleNotFoundCache = "null"

type ArticleDetailResponse struct {
	models.ArticleModel
	CategoryTitle *string `json:"categoryTitle"`
	UserName      string  `json:"username"`
	NickName      string  `json:"nickname"`
	UserAvatar    string  `json:"userAvatar"`
}

func (ArticleApi) ArticleDetailView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	//redis缓存
	idStr := strconv.FormatUint(uint64(req.ID), 10)
	ctx := context.Background()
	res, exit := global.RedisHotPool.Get(ctx, "ArticleID"+idStr).Result()
	if exit == nil {
		// 负缓存:文章不存在,直接返回,避免不存在的ID反复查库
		if res == articleNotFoundCache {
			response.FailWithMsg("文章不存在", c)
			return
		}
		// 缓存命中
		var cached ArticleDetailResponse
		err := json.Unmarshal([]byte(res), &cached)
		if err != nil {
			response.FailWithMsg("缓存数据解析错误", c)
			return
		}
		//权限检查:仅管理员或作者可见未发布文章
		claims, _ := jwts.ParseTokenByGin(c)
		isAdminOrOwner := claims != nil && (claims.Role == enum.AdminRole || claims.UserID == cached.UserID)
		if !isAdminOrOwner && cached.Status != models.StatusPublished {
			response.FailWithMsg("文章不存在", c)
			return
		}

		// 计数只在响应阶段叠加,不写回详情缓存
		result := cached
		applyArticleCountDeltas([]*models.ArticleModel{&result.ArticleModel})

		response.OkWithData(result, c)
		return
	}

	// 未登录的用户，只能看到发布成功的文章
	// 登录用户，能看到自己的所有文章
	// 管理员，能看到全部的文章

	var article models.ArticleModel
	err := global.DB.Preload("UserModel").Preload("CategoryModel").Preload("ArticleAddition").Take(&article, req.ID).Error
	if err != nil {
		//写入短时负缓存,防止被不存在的ID刷库
		global.RedisHotPool.Set(ctx, "ArticleID"+idStr, articleNotFoundCache, time.Minute)
		response.FailWithMsg("文章不存在", c)
		return
	}

	//权限检查:仅管理员或作者可见未发布文章
	claims, _ := jwts.ParseTokenByGin(c)
	isAdminOrOwner := claims != nil && (claims.Role == enum.AdminRole || claims.UserID == article.UserID)
	if !isAdminOrOwner && article.Status != models.StatusPublished {
		response.FailWithMsg("文章不存在", c)
		return
	}

	var categoryTitle *string
	if article.CategoryModel != nil {
		categoryTitle = &article.CategoryModel.Title
	}
	//AI点评通过外键预加载,直接随文章一起返回
	cached := ArticleDetailResponse{
		ArticleModel:  article,
		CategoryTitle: categoryTitle,
		UserName:      article.UserModel.UserName,
		NickName:      article.UserModel.NickName,
		UserAvatar:    article.UserModel.Avatar,
	}

	// 计数只在响应阶段叠加,不写回详情缓存
	result := cached
	applyArticleCountDeltas([]*models.ArticleModel{&result.ArticleModel})

	response.OkWithData(result, c)

	//把数据加入缓存
	// logrus.Debug("缓存文章详情", idStr, idStr, cached) //debug
	jsonData, err := json.Marshal(cached)
	if err != nil {
		logrus.Error("缓存数据序列化错误:", err)
		return
	}
	global.RedisHotPool.Set(ctx, "ArticleID"+idStr, string(jsonData), 10*time.Minute)
}
