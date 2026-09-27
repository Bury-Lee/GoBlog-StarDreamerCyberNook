package article_api

import (
	"context"
	"fmt"
	"strconv"

	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/service/ai_service"
	"StarDreamerCyberNook/service/content_service"
	xss_filter "StarDreamerCyberNook/utils/XSSfilter"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ArticleUpdateRequest struct {
	ID          uint     `json:"id" binding:"required"`
	Title       string   `json:"title" binding:"required"`
	Abstract    string   `json:"abstract"`
	Content     string   `json:"content" binding:"required"`
	CategoryID  *uint    `json:"categoryID"`
	TagList     []string `json:"tagList"`
	Cover       string   `json:"cover"`
	OpenComment bool     `json:"openComment"`
}

// ArticleUpdateView 整体更新文章(经 content 服务落库)。
func (ArticleApi) ArticleUpdateView(c *gin.Context) {
	var req ArticleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	user, err := jwts.GetClaims(c).GetUser()
	if err != nil {
		response.FailWithMsg("用户不存在", c)
		return
	}

	detail, err := content_service.GetArticle(req.ID)
	if err != nil {
		response.FailWithMsg("文章不存在", c)
		return
	}
	article := detail.ArticleModel
	if article.UserID != user.ID {
		response.FailWithMsg("只能更新自己的文章", c)
		return
	}

	// 防xss注入
	xssFilter := xss_filter.NewXSSFilter()
	req.Title = xss_filter.SanitizeText(req.Title)
	if req.Content == "" {
		response.FailWithMsg("正文解析错误", c)
		return
	}
	req.Content = xssFilter.Sanitize(req.Content)
	if req.Abstract != "" {
		req.Abstract = xssFilter.Sanitize(req.Abstract)
	} else {
		req.Abstract = "该文章未设置简介"
	}

	// 状态:发布中且开启审核时,编辑后回到待审核
	status := article.Status
	if article.Status == models.StatusPublished && global.Config.Site.Article.EnableExamination {
		status = models.StatusPending
	}

	var aiQuality, aiAbstract string
	if global.Config.AI.Enable && global.Config.Site.Article.EnableExamination {
		res, e := ai_service.CreateSingleReply(
			"文章标题:"+req.Title+"\n文章摘要:"+req.Abstract+"\n文章内容:"+req.Content,
			global.SystemPromptArticleReview.String(),
		)
		if e != nil {
			logrus.Error("ai审核失败", e.Error())
			response.FailWithMsg("ai审核失败,已经自动改为为待审核状态", c)
			return
		}
		switch res {
		case "通过":
			status = models.StatusPublished
			if q, e := ai_service.GenerateArticleQuality(req.Title, req.Abstract, req.Content); e != nil {
				logrus.Errorf("ai自动创建评级失败: %s", e.Error())
			} else {
				aiQuality = q
			}
			if a, e := ai_service.GenerateArticleAbstract(req.Title, req.Abstract, req.Content); e != nil {
				logrus.Errorf("ai自动创建摘要失败: %s", e.Error())
			} else {
				aiAbstract = a
			}
		case "拒绝":
			status = models.StatusDraft
		default:
			logrus.Errorf("ai审核出错,回复内容:%s,文章详情:%s\n已自动换为待审核状态", res, fmt.Sprintf("%#v", req))
			status = models.StatusPending
		}
	}

	// 分类:空/0 表示清空
	var catPtr *uint
	if req.CategoryID == nil || *req.CategoryID == 0 {
		z := uint(0)
		catPtr = &z
	} else {
		catPtr = req.CategoryID
	}

	if err := content_service.UpdateArticle(content_service.UpdateReq{
		ID:          req.ID,
		OwnerID:     user.ID,
		Title:       &req.Title,
		Abstract:    &req.Abstract,
		Content:     &req.Content,
		Cover:       &req.Cover,
		OpenComment: &req.OpenComment,
		CategoryID:  catPtr,
		Status:      &status,
		TagList:     &req.TagList,
		AIQuality:   aiQuality,
		AIAbstract:  aiAbstract,
		AIModel:     global.Config.AI.Model,
	}); err != nil {
		response.FailWithMsg("更新失败", c)
		return
	}

	global.RedisHotPool.Del(context.Background(), "ArticleID"+strconv.FormatUint(uint64(req.ID), 10))
	response.OkWithMsg("文章更新成功,当前状态为:"+status.String(), c)
}

type ArticleUpdateRequest2 struct {
	ID          uint           `json:"id" binding:"required"`
	Title       *string        `json:"title"`
	Abstract    *string        `json:"abstract"`
	Content     *string        `json:"content"`
	CategoryID  *uint          `json:"categoryID"`
	TagList     *[]string      `json:"tagList"`
	Cover       *string        `json:"cover"`
	OpenComment *bool          `json:"openComment"`
	Status      *models.Status `json:"status"`
}

// ArticleUpdateView2 增量更新文章(经 content 服务落库)。
func (ArticleApi) ArticleUpdateView2(c *gin.Context) {
	var req ArticleUpdateRequest2
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	user, err := jwts.GetClaims(c).GetUser()
	if err != nil {
		response.FailWithMsg("用户不存在", c)
		return
	}

	detail, err := content_service.GetArticle(req.ID)
	if err != nil {
		response.FailWithMsg("文章不存在", c)
		return
	}
	article := detail.ArticleModel
	if article.UserID != user.ID {
		response.FailWithMsg("只能更新自己的文章", c)
		return
	}

	upd := content_service.UpdateReq{ID: req.ID, OwnerID: user.ID}
	finalStatus := article.Status
	changed := false

	if req.Title != nil {
		upd.Title = req.Title
		changed = true
	}
	if req.CategoryID != nil {
		upd.CategoryID = req.CategoryID // 0 表示清空
		changed = true
	}
	xssFilter := xss_filter.NewXSSFilter()
	if req.Content != nil {
		if *req.Content == "" {
			response.FailWithMsg("正文解析错误", c)
			return
		}
		v := xssFilter.Sanitize(*req.Content)
		upd.Content = &v
		changed = true
	}
	if req.Abstract != nil {
		v := "该文章未设置简介"
		if *req.Abstract != "" {
			v = xssFilter.Sanitize(*req.Abstract)
		}
		upd.Abstract = &v
		changed = true
	}
	if req.Cover != nil {
		upd.Cover = req.Cover
		changed = true
	}
	if req.OpenComment != nil {
		upd.OpenComment = req.OpenComment
		changed = true
	}
	if req.TagList != nil {
		upd.TagList = req.TagList
		changed = true
	}

	if !changed {
		response.OkWithMsg("未做任何修改", c)
		return
	}

	contentChanged := req.Title != nil || req.Abstract != nil || req.Content != nil

	var aiQuality, aiAbstract string
	if contentChanged {
		if article.Status == models.StatusPublished && global.Config.Site.Article.EnableExamination {
			finalStatus = models.StatusPending
		}
		if global.Config.AI.Enable && global.Config.Site.Article.EnableExamination {
			titleForAI := article.Title
			if req.Title != nil {
				titleForAI = *req.Title
			}
			abstractForAI := article.Abstract
			if upd.Abstract != nil {
				abstractForAI = *upd.Abstract
			}
			contentForAI := article.Content
			if upd.Content != nil {
				contentForAI = *upd.Content
			}

			res, e := ai_service.CreateSingleReply(
				"文章标题:"+titleForAI+"\n文章摘要:"+abstractForAI+"\n文章内容:"+contentForAI,
				global.SystemPromptArticleReview.String(),
			)
			if e != nil {
				logrus.Error("ai审核失败", e.Error())
				response.FailWithMsg("ai审核失败,已经自动改为为待审核状态", c)
				return
			}
			switch res {
			case "通过":
				finalStatus = models.StatusPublished
				if q, e := ai_service.GenerateArticleQuality(titleForAI, abstractForAI, contentForAI); e != nil {
					logrus.Errorf("ai自动创建评级失败: %s", e.Error())
				} else {
					aiQuality = q
				}
				if a, e := ai_service.GenerateArticleAbstract(titleForAI, abstractForAI, contentForAI); e != nil {
					logrus.Errorf("ai自动创建摘要失败: %s", e.Error())
				} else {
					aiAbstract = a
				}
			case "拒绝":
				finalStatus = models.StatusDraft
			default:
				logrus.Errorf("ai审核出错,回复内容:%s,文章详情:%s\n已自动换为待审核状态", res, fmt.Sprintf("%#v", req))
				finalStatus = models.StatusPending
			}
		}
	}

	// 显式状态变更:普通用户仅草稿/审核中,管理员任意合法状态
	if req.Status != nil {
		st := *req.Status
		if st < models.StatusDraft || st > models.StatusOffline {
			response.FailWithMsg("非法的文章状态", c)
			return
		}
		if user.Role != enum.AdminRole {
			if st != models.StatusDraft && st != models.StatusPending {
				response.FailWithMsg("非法的文章状态", c)
				return
			}
			if st == models.StatusPending && !global.Config.Site.Article.EnableExamination {
				st = models.StatusPublished
			}
		}
		finalStatus = st
	}
	upd.Status = &finalStatus
	upd.AIQuality = aiQuality
	upd.AIAbstract = aiAbstract
	upd.AIModel = global.Config.AI.Model

	if err := content_service.UpdateArticle(upd); err != nil {
		response.FailWithMsg("更新失败", c)
		return
	}

	global.RedisHotPool.Del(context.Background(), "ArticleID"+strconv.FormatUint(uint64(req.ID), 10))
	response.OkWithMsg("文章更新成功,当前状态为:"+finalStatus.String(), c)
}
