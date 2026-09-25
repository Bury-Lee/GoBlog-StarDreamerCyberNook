// api/feedback_api/enter.go
// 用户反馈墙:提交(登录可选/可匿名)、全站公开列表、管理员处理(增量更新)
package feedback_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	xss_filter "StarDreamerCyberNook/utils/XSSfilter"
	jwts "StarDreamerCyberNook/utils/jwts"
	"strings"

	"github.com/gin-gonic/gin"
)

// FeedbackApi 反馈接口
type FeedbackApi struct{}

type FeedbackCreateRequest struct {
	Content     string              `json:"content" binding:"required"` // 反馈内容
	Contact     string              `json:"contact"`                    // 联系方式,可选
	Type        models.FeedbackType `json:"type"`                       // 反馈类型,默认0
	IsAnonymous bool                `json:"isAnonymous"`                // 是否匿名提交
}

// FeedbackCreateView 提交反馈,登录可选(未登录记为 UserID=0);匿名提交仅隐藏展示,不改动落库内容
func (FeedbackApi) FeedbackCreateView(c *gin.Context) {
	var req FeedbackCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	content := strings.TrimSpace(xss_filter.SanitizeText(req.Content))
	if content == "" {
		response.FailWithMsg("反馈内容不能为空", c)
		return
	}
	if len([]rune(content)) > 2000 {
		response.FailWithMsg("反馈内容过长", c)
		return
	}
	if req.Type < models.FeedbackTypeOther || req.Type > models.FeedbackTypeReport {
		req.Type = models.FeedbackTypeOther
	}

	//登录可选:已登录才记录提交者ID,未登录记为0
	var userID uint
	if claims, err := jwts.ParseTokenByGin(c); err == nil && claims != nil {
		userID = claims.UserID
	}

	if err := global.DB.Create(&models.FeedbackModel{
		UserID:      userID,
		IsAnonymous: req.IsAnonymous,
		Content:     content,
		Contact:     strings.TrimSpace(req.Contact),
		Type:        req.Type,
	}).Error; err != nil {
		response.FailWithMsg("提交失败", c)
		return
	}
	response.OkWithMsg("感谢反馈,我们会尽快处理", c)
}

type FeedbackListRequest struct {
	common.PageInfo
	Status *models.FeedbackStatus `form:"status"` // 处理状态筛选,可选
	Type   *models.FeedbackType   `form:"type"`   // 类型筛选,可选
}

// FeedbackWallView 反馈墙列表(全站公开):直接走通用分页函数,返回前过滤隐私字段
func (FeedbackApi) FeedbackWallView(c *gin.Context) {
	var req FeedbackListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	options := common.Options{
		PageInfo:      req.PageInfo,
		DefaultOrder:  "created_at desc",
		AllowedOrders: []string{"id", "created_at", "status"},
		CountCap:      1000, //总数封顶,避免大表全表扫描
	}
	if req.Status != nil {
		options.Where = global.DB.Where("status = ?", *req.Status)
	}
	if req.Type != nil {
		if options.Where != nil {
			options.Where = options.Where.Where("type = ?", *req.Type)
		} else {
			options.Where = global.DB.Where("type = ?", *req.Type)
		}
	}

	list, count, capped, err := common.ListQuery(models.FeedbackModel{}, options)
	if err != nil {
		response.FailWithMsg("查询失败", c)
		return
	}

	//过滤隐私字段:联系方式/处理人不对外返回(空值被 omitempty 省略);匿名反馈不返回提交者ID
	for i := range list {
		list[i].Contact = ""
		list[i].HandlerID = 0
		if list[i].IsAnonymous {
			list[i].UserID = 0
		}
	}
	response.OkWithListCapped(list, count, capped, c)
}

type FeedbackHandleRequest struct {
	Status models.FeedbackStatus `json:"status"` // 目标状态:0待处理 1已采纳未处理 2正在处理 3已处理
	Reply  string                `json:"reply"`  // 管理员回复
}

// FeedbackHandleView 管理员处理反馈:更新状态与回复,并返回更新后的条目供前端增量替换
func (FeedbackApi) FeedbackHandleView(c *gin.Context) {
	var uri models.IDRequest
	if err := c.ShouldBindUri(&uri); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	var req FeedbackHandleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	//状态必须是合法枚举
	if req.Status < models.FeedbackStatusPending || req.Status > models.FeedbackStatusResolved {
		response.FailWithMsg("非法的状态", c)
		return
	}

	var feedback models.FeedbackModel
	if err := global.DB.Take(&feedback, uri.ID).Error; err != nil {
		response.FailWithMsg("反馈不存在", c)
		return
	}

	claims := jwts.GetClaims(c)
	reply := strings.TrimSpace(xss_filter.SanitizeText(req.Reply))
	if err := global.DB.Model(&feedback).Updates(map[string]any{
		"status":     req.Status,
		"reply":      reply,
		"handler_id": claims.UserID,
	}).Error; err != nil {
		response.FailWithMsg("处理失败", c)
		return
	}

	//回写内存值,并做与列表一致的隐私过滤后返回,前端据此在列表中增量替换
	feedback.Status = req.Status
	feedback.Reply = reply
	feedback.Contact = ""
	feedback.HandlerID = 0
	if feedback.IsAnonymous {
		feedback.UserID = 0
	}
	response.OkWithData(feedback, c)
}
