// api/moment_api/enter.go
// 动态/日记接口:发布、列表(隐私过滤)、详情、更新、删除、点赞、转发
package moment_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/service/ai_service"
	xss_filter "StarDreamerCyberNook/utils/XSSfilter"
	jwts "StarDreamerCyberNook/utils/jwts"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// MomentApi 动态接口
type MomentApi struct{}

type MomentCreateRequest struct {
	Content    string                  `json:"content"`    // 正文
	Images     []string                `json:"images"`     // 图片URL列表
	Type       models.MomentType       `json:"type"`       // 0动态 1日记
	Visibility models.MomentVisibility `json:"visibility"` // 0公开 1仅好友 2私密
	Status     models.Status           `json:"status"`     // 0草稿 2已发布
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
	//普通用户只能存草稿或发布,不能设置审核中/下线
	status := models.StatusPublished
	if req.Status == models.StatusDraft {
		status = models.StatusDraft
	}

	//AI审核:与文章审核同款逻辑(受 ai.enable + site.moment.enableExamination 控制);草稿不审
	if global.Config.AI.Enable && global.Config.Site.Moment.EnableExamination && status == models.StatusPublished {
		reply, err := ai_service.CreateSingleReply(
			"动态内容:"+content,
			global.SystemPromptArticleReview.String(),
		)
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
	model := models.MomentModel{
		UserID:     claims.UserID,
		Type:       req.Type,
		Visibility: req.Visibility,
		Content:    content,
		Images:     req.Images,
		Status:     status,
	}
	if err := global.DB.Create(&model).Error; err != nil {
		response.FailWithMsg("发布失败", c)
		return
	}
	//回填作者信息,便于前端直接展示(仅公开字段)
	var user models.UserModel
	if global.DB.Take(&user, claims.UserID).Error == nil {
		model.UserModel = models.UserModel{
			Model:         user.Model,
			NickName:      user.NickName,
			Avatar:        user.Avatar,
			LastLoginTime: user.LastLoginTime,
			Age:           user.Age,
			LikeTags:      user.LikeTags,
		}
	}
	response.OkWithData(model, c)
}

type MomentListRequest struct {
	common.PageInfo
	UserID uint               `form:"userID"` // 指定作者,可选
	Type   *models.MomentType `form:"type"`   // 动态/日记筛选,可选
}

// MomentListView 动态列表:按隐私规则过滤
func (MomentApi) MomentListView(c *gin.Context) {
	var req MomentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	viewerID, isAdmin := viewerOf(c)

	where := global.DB
	if req.UserID != 0 {
		where = where.Where("user_id = ?", req.UserID)
		//本人或管理员可见全部,不加过滤
		if req.UserID != viewerID && !isAdmin {
			//好友判定:访问者对作者存在 friend=true 的关注记录(双向关注)
			var rel models.UserFollowModel
			if viewerID != 0 && global.DB.Take(&rel, "user_id = ? and focus_user_id = ? and friend = ?", viewerID, req.UserID, true).Error == nil {
				where = where.Where("status = ? and visibility in ?", models.StatusPublished,
					[]models.MomentVisibility{models.MomentVisibilityPublic, models.MomentVisibilityFriends})
			} else {
				where = where.Where("status = ? and visibility = ?", models.StatusPublished, models.MomentVisibilityPublic)
			}
		}
	} else {
		//未登录时:仅公开且已发布
		where = where.Where("status = ? and visibility = ?", models.StatusPublished, models.MomentVisibilityPublic)
	}
	if req.Type != nil {
		where = where.Where("type = ?", *req.Type)
	}

	options := common.Options{
		PageInfo:      req.PageInfo,
		Preloads:      []string{"UserModel", "RepostFrom.UserModel"},
		Where:         where,
		DefaultOrder:  "created_at desc",
		AllowedOrders: []string{"id", "created_at", "like_count", "comment_count"},
		CountCap:      common.DefaultCountCap,
	}
	list, count, capped, err := common.ListQuery(models.MomentModel{}, options)
	if err != nil {
		response.FailWithMsg("查询失败", c)
		return
	}
	for i := range list {
		list[i].UserModel = models.UserModel{
			Model:         list[i].UserModel.Model,
			NickName:      list[i].UserModel.NickName,
			Avatar:        list[i].UserModel.Avatar,
			LastLoginTime: list[i].UserModel.LastLoginTime,
			Age:           list[i].UserModel.Age,
			LikeTags:      list[i].UserModel.LikeTags,
		}
		if list[i].RepostFrom != nil {
			list[i].RepostFrom.UserModel = models.UserModel{
				Model:         list[i].RepostFrom.UserModel.Model,
				NickName:      list[i].RepostFrom.UserModel.NickName,
				Avatar:        list[i].RepostFrom.UserModel.Avatar,
				LastLoginTime: list[i].RepostFrom.UserModel.LastLoginTime,
				Age:           list[i].RepostFrom.UserModel.Age,
				LikeTags:      list[i].RepostFrom.UserModel.LikeTags,
			}
		}
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
	var moment models.MomentModel
	if err := global.DB.Preload("UserModel").Preload("RepostFrom.UserModel").Take(&moment, req.ID).Error; err != nil {
		response.FailWithMsg("动态不存在", c)
		return
	}
	viewerID, isAdmin := viewerOf(c)
	if !canViewMoment(moment, viewerID, isAdmin) {
		response.FailWithMsg("动态不存在", c)
		return
	}
	moment.UserModel = models.UserModel{
		Model:         moment.UserModel.Model,
		NickName:      moment.UserModel.NickName,
		Avatar:        moment.UserModel.Avatar,
		LastLoginTime: moment.UserModel.LastLoginTime,
		Age:           moment.UserModel.Age,
		LikeTags:      moment.UserModel.LikeTags,
	}
	if moment.RepostFrom != nil {
		moment.RepostFrom.UserModel = models.UserModel{
			Model:         moment.RepostFrom.UserModel.Model,
			NickName:      moment.RepostFrom.UserModel.NickName,
			Avatar:        moment.RepostFrom.UserModel.Avatar,
			LastLoginTime: moment.RepostFrom.UserModel.LastLoginTime,
			Age:           moment.RepostFrom.UserModel.Age,
			LikeTags:      moment.RepostFrom.UserModel.LikeTags,
		}
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
	var moment models.MomentModel
	if err := global.DB.Take(&moment, req.ID).Error; err != nil {
		response.FailWithMsg("动态不存在", c)
		return
	}
	claims := jwts.GetClaims(c)
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
	status := moment.Status
	if req.Status == models.StatusDraft || req.Status == models.StatusPublished {
		status = req.Status
	}

	if err := global.DB.Model(&moment).Updates(map[string]any{
		"content":    content,
		"images":     imagesJSON(req.Images),
		"type":       req.Type,
		"visibility": req.Visibility,
		"status":     status,
	}).Error; err != nil {
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
	var moment models.MomentModel
	if err := global.DB.Take(&moment, req.ID).Error; err != nil {
		response.FailWithMsg("动态不存在", c)
		return
	}
	claims := jwts.GetClaims(c)
	if claims.UserID != moment.UserID && claims.Role != enum.AdminRole {
		response.FailWithMsg("没有权限删除该动态", c)
		return
	}

	//事务:级联删除 + 转发来源计数-1
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		//先删该动态下所有评论的点赞(必须在删评论之前,子查询才拿得到评论ID)
		if err := tx.Where("moment_comment_id in (?)",
			tx.Model(&models.MomentCommentModel{}).Select("id").Where("moment_id = ?", moment.ID)).
			Delete(&models.MomentCommentDiggModel{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.MomentCommentModel{}, "moment_id = ?", moment.ID).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.MomentDiggModel{}, "moment_id = ?", moment.ID).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.MomentModel{}, "id = ?", moment.ID).Error; err != nil {
			return err
		}
		if moment.RepostFromID != nil {
			if err := tx.Model(&models.MomentModel{}).Where("id = ? and repost_count > 0", *moment.RepostFromID).
				UpdateColumn("repost_count", gorm.Expr("repost_count - 1")).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		response.FailWithMsg("删除失败", c)
		return
	}
	response.OkWithMsg("删除成功", c)
}

// MomentInteractionView 当前用户对该动态的点赞状态(未登录视为未点赞)
func (MomentApi) MomentInteractionView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	viewerID, _ := viewerOf(c)
	digged := false
	if viewerID != 0 {
		err := global.DB.Take(&models.MomentDiggModel{}, "user_id = ? and moment_id = ?", viewerID, req.ID).Error
		digged = err == nil
	}
	response.OkWithData(gin.H{"digged": digged}, c)
}
