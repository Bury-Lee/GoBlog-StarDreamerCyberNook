// api/moment_api/enter.go
// 动态/日记接口:发布、列表(隐私过滤)、详情、更新、删除(经 content 服务)
package moment_api

import (
	"strings"

	"StarDreamerCyberNook/common"
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
	"google.golang.org/grpc/status"
)

// MomentApi 动态接口
type MomentApi struct{}

type MomentCreateRequest struct {
	Content    string                  `json:"content"`
	Images     []string                `json:"images"`
	Type       models.MomentType       `json:"type"`
	Visibility models.MomentVisibility `json:"visibility"`
	Status     models.Status           `json:"status"`
}

// MomentCreateView 发布动态/日记
func (MomentApi) MomentCreateView(c *gin.Context) {
	var req MomentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	content := strings.TrimSpace(xss_filter.SanitizeText(req.Content))
	if content == "" && len(req.Images) == 0 {
		response.FailWithMsg("动态内容不能为空", c)
		return
	}
	if len([]rune(content)) > 16000 {
		response.FailWithMsg("内容过长", c)
		return
	}
	if len(req.Images) > 9 {
		response.FailWithMsg("最多上传9张图片", c)
		return
	}
	if req.Type != models.MomentTypeMoment && req.Type != models.MomentTypeDiary {
		req.Type = models.MomentTypeMoment
	}
	if req.Visibility < models.MomentVisibilityPublic || req.Visibility > models.MomentVisibilityPrivate {
		req.Visibility = models.MomentVisibilityPublic
	}
	status := models.StatusPublished
	if req.Status == models.StatusDraft {
		status = models.StatusDraft
	}

	if global.Config.AI.Enable && global.Config.Site.Moment.EnableExamination && status == models.StatusPublished {
		reply, err := ai_service.CreateSingleReply("动态内容:"+content, global.SystemPromptArticleReview.String())
		if err != nil {
			logrus.Error("动态AI审核失败:" + err.Error())
			response.FailWithMsg("ai审核失败,已经自动创建为待审核状态", c)
			return
		}
		switch reply {
		case "通过":
			status = models.StatusPublished
		case "拒绝":
			status = models.StatusDraft
		default:
			logrus.Errorf("动态AI审核出错,回复内容:%s,动态内容:%s", reply, content)
			status = models.StatusPending
		}
	}

	claims := jwts.GetClaims(c)
	model, err := content_service.CreateMoment(claims.UserID, req.Type, req.Visibility, content, req.Images, status)
	if err != nil {
		response.FailWithMsg("发布失败", c)
		return
	}
	response.OkWithData(model, c)
}

type MomentListRequest struct {
	common.PageInfo
	UserID uint               `form:"userID"`
	Type   *models.MomentType `form:"type"`
}

// MomentListView 动态列表:按隐私规则过滤(在 content 服务)
func (MomentApi) MomentListView(c *gin.Context) {
	var req MomentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	viewerID, isAdmin := viewerOf(c)

	list, count, capped, err := content_service.ListMoments(content_service.MomentListQuery{
		ViewerID: viewerID, IsAdmin: isAdmin, UserID: req.UserID, Type: req.Type,
		Page: req.Page, Limit: req.Limit, Key: req.Key, Order: req.Order, EndID: req.EndId,
	})
	if err != nil {
		response.FailWithMsg("查询失败", c)
		return
	}
	response.OkWithListCapped(list, count, capped, c)
}

// MomentDetailView 动态详情
func (MomentApi) MomentDetailView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	viewerID, isAdmin := viewerOf(c)
	moment, err := content_service.GetMoment(req.ID, viewerID, isAdmin)
	if err != nil {
		response.FailWithMsg("动态不存在", c)
		return
	}
	response.OkWithData(moment, c)
}

type MomentUpdateRequest struct {
	ID         uint                    `json:"id" binding:"required"`
	Content    string                  `json:"content"`
	Images     []string                `json:"images"`
	Type       models.MomentType       `json:"type"`
	Visibility models.MomentVisibility `json:"visibility"`
	Status     models.Status           `json:"status"`
}

// MomentUpdateView 更新自己的动态
func (MomentApi) MomentUpdateView(c *gin.Context) {
	var req MomentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claims := jwts.GetClaims(c)
	moment, err := content_service.GetMoment(req.ID, claims.UserID, false)
	if err != nil {
		response.FailWithMsg("动态不存在", c)
		return
	}
	if claims.UserID != moment.UserID {
		response.FailWithMsg("没有权限修改该动态", c)
		return
	}

	content := strings.TrimSpace(xss_filter.SanitizeText(req.Content))
	if content == "" && len(req.Images) == 0 {
		response.FailWithMsg("动态内容不能为空", c)
		return
	}
	if len([]rune(content)) > 2000 {
		response.FailWithMsg("内容过长", c)
		return
	}
	if len(req.Images) > 9 {
		response.FailWithMsg("最多上传9张图片", c)
		return
	}
	if req.Type != models.MomentTypeMoment && req.Type != models.MomentTypeDiary {
		req.Type = moment.Type
	}
	if req.Visibility < models.MomentVisibilityPublic || req.Visibility > models.MomentVisibilityPrivate {
		req.Visibility = moment.Visibility
	}
	newStatus := moment.Status
	if req.Status == models.StatusDraft || req.Status == models.StatusPublished {
		newStatus = req.Status
	}

	if err := content_service.UpdateMoment(req.ID, claims.UserID, content, req.Images, req.Type, req.Visibility, newStatus); err != nil {
		response.FailWithMsg("更新失败", c)
		return
	}
	response.OkWithMsg("更新成功", c)
}

// MomentRemoveView 删除动态(本人/管理员),级联删除评论与点赞
func (MomentApi) MomentRemoveView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claims := jwts.GetClaims(c)
	isAdmin := claims.Role == enum.AdminRole
	if err := content_service.RemoveMoment(req.ID, claims.UserID, isAdmin); err != nil {
		if st, ok := status.FromError(err); ok {
			response.FailWithMsg(st.Message(), c)
		} else {
			response.FailWithMsg("删除失败", c)
		}
		return
	}
	response.OkWithMsg("删除成功", c)
}

// MomentInteractionView 当前用户对该动态的点赞状态
func (MomentApi) MomentInteractionView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	viewerID, _ := viewerOf(c)
	digged, err := content_service.MomentInteraction(req.ID, viewerID)
	if err != nil {
		response.FailWithMsg("查询失败", c)
		return
	}
	response.OkWithData(gin.H{"digged": digged}, c)
}
