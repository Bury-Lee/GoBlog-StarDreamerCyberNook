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

type ArticleCreateRequest struct {
	Title       string        `json:"title" binding:"required"`   // 文章标题，最大32字符
	Abstract    string        `json:"abstract"`                   // 文章摘要，最大256字符
	Content     string        `json:"content" binding:"required"` // 文章内容
	CategoryID  *uint         `json:"categoryID"`                 // 文章分类ID，关联分类表
	TagList     []string      `json:"tagList"`                    // 标签列表
	Cover       string        `json:"cover"`                      // 文章封面图片URL
	OpenComment bool          `json:"openComment"`                // 是否开启评论
	Stats       models.Status `json:"status"`                     // 状态(普通用户仅草稿/审核中)
}

func (ArticleApi) ArticleCreateView(c *gin.Context) {
	var req ArticleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("文章参数绑定失败: "+err.Error(), c)
		return
	}

	User, err := jwts.GetClaims(c).GetUser()
	if err != nil {
		response.FailWithMsg("获取用户信息失败", c)
		return
	}
	if global.Config.Site.SiteInfo.Mode == 2 && User.Role != enum.AdminRole {
		response.FailWithMsg("未开放文章创建", c)
		return
	}
	if (req.Stats != 1 && req.Stats != 0) && User.Role != enum.AdminRole {
		response.FailWithMsg("非法参数", c)
		return
	}
	if len(req.Abstract) > 200 { //避免在搜索时刷屏
		response.FailWithMsg("简介过长", c)
		return
	}

	if !global.Config.Site.Article.EnableExamination && req.Stats == models.StatusPending {
		req.Stats = models.StatusPublished
		//未启用审核且设置为审核中状态时跳过审核
	}

	//防xss注入
	xssFilter := xss_filter.NewXSSFilter()
	req.Title = xss_filter.SanitizeText(req.Title)
	req.Content = xssFilter.Sanitize(req.Content)
	if req.Content == "" {
		response.FailWithMsg("正文解析错误", c)
		return
	}
	//不传简介时就设为无,传了就做清洗
	if req.Abstract != "" {
		x := xss_filter.NewXSSFilter()
		req.Abstract = x.Sanitize(req.Abstract)
	} else {
		req.Abstract = "该文章未设置简介"
	}

	if global.Config.AI.Enable && global.Config.Site.Article.EnableExamination { //启用ai审核
		reply, err := ai_service.CreateSingleReply(
			"文章标题:"+req.Title+"\n文章摘要:"+req.Abstract+"\n文章内容:"+req.Content,
			global.SystemPromptArticleReview.String(),
		)
		if err != nil {
			logrus.Error("ai审核失败:" + err.Error())
			response.FailWithMsg("ai审核失败,已经自动创建为待审核状态", c)
			return
		}
		switch reply {
		case "通过":
			req.Stats = models.StatusPublished
		case "拒绝":
			req.Stats = models.StatusDraft
		default:
			logrus.Errorf("ai审核出错,回复内容:%s,文章详情:%s\n已自动换为待审核状态", reply, fmt.Sprintf("%#v", req))
			req.Stats = models.StatusPending
		}
	}

	//生成AI点评(摘要+评级)
	var aiQuality, aiAbstract string
	if global.Config.AI.Enable {
		if reply, e := ai_service.GenerateArticleAbstract(req.Title, req.Abstract, req.Content); e != nil {
			logrus.Errorf("ai自动创建摘要失败: %s", e.Error())
		} else {
			aiAbstract = reply
		}
		if reply, e := ai_service.GenerateArticleQuality(req.Title, req.Abstract, req.Content); e != nil {
			logrus.Errorf("ai自动创建评级失败: %s", e.Error())
		} else {
			aiQuality = reply
		}
	}

	// 经 content 服务持久化(分类归属校验、AI点评写入均在服务端)
	articleID, cerr := content_service.CreateArticle(content_service.CreateReq{
		UserID:      User.ID,
		Title:       req.Title,
		Abstract:    req.Abstract,
		Content:     req.Content,
		CategoryID:  req.CategoryID,
		TagList:     req.TagList,
		Cover:       req.Cover,
		OpenComment: req.OpenComment,
		Status:      req.Stats,
		AIQuality:   aiQuality,
		AIAbstract:  aiAbstract,
		AIModel:     global.Config.AI.Model,
	})
	if cerr != nil {
		response.FailWithMsg("文章创建失败", c)
		return
	}

	//清理可能存在的负缓存(SQLite等复用ID)
	global.RedisHotPool.Del(context.Background(), "ArticleID"+strconv.FormatUint(uint64(articleID), 10))

	response.OkWithMsg(fmt.Sprintf("文章创建成功,当前状态:%s", req.Stats.String()), c)
}
