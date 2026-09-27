// api/moment_api/comment.go
// 动态评论:发表、列表、子评论、删除、点赞(经 content 服务)
package moment_api

import (
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

type MomentCommentCreateRequest struct {
	Content  string `json:"content" binding:"required"`
	MomentID uint   `json:"momentID" binding:"required"`
	ParentID uint   `json:"parentID"`
}

// MomentCommentCreateView 发表评论/回复
func (MomentApi) MomentCommentCreateView(c *gin.Context) {
	var req MomentCommentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	viewerID, isAdmin := viewerOf(c)
	if _, err := content_service.GetMoment(req.MomentID, viewerID, isAdmin); err != nil {
		response.FailWithMsg("动态不存在", c)
		return
	}

	content := xss_filter.NewXSSFilter().Sanitize(req.Content)
	if content == "" {
		response.FailWithMsg("评论内容不能为空", c)
		return
	}

	if global.Config.AI.Enable && global.Config.Site.Moment.EnableExamination {
		reply, err := ai_service.CreateSingleReply("评论内容:"+content, global.SystemPromptComment.String())
		if err != nil {
			response.FailWithMsg("ai审核失败,已经自动创建为待审核状态: "+err.Error(), c)
			return
		}
		switch reply {
		case "通过":
		case "拒绝":
			response.FailWithMsg("评论可能含有违规信息,已拒绝", c)
			return
		default:
			logrus.Errorf("动态评论AI审核出错,回复内容:%s,评论内容:%s", reply, content)
			return
		}
	}

	claims := jwts.GetClaims(c)
	model, err := content_service.CreateMomentComment(claims.UserID, req.MomentID, req.ParentID, content, viewerID, isAdmin)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			response.FailWithMsg(st.Message(), c)
		} else {
			response.FailWithMsg("评论失败", c)
		}
		return
	}
	response.OkWithData(model, c)
}

type MomentCommentListRequest struct {
	common.PageInfo
	MomentID uint `form:"momentID" binding:"required"`
}

// MomentCommentListView 获取动态的一级评论(分页)
func (MomentApi) MomentCommentListView(c *gin.Context) {
	var req MomentCommentListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	viewerID, isAdmin := viewerOf(c)
	list, count, capped, err := content_service.ListMomentComments(req.MomentID, viewerID, isAdmin, req.Page, req.Limit, req.Key, req.Order, req.EndId)
	if err != nil {
		response.FailWithMsg("查询评论失败", c)
		return
	}
	response.OkWithListCapped(list, count, capped, c)
}

type MomentCommentChildListRequest struct {
	common.PageInfo
	Root uint `form:"root" binding:"required"`
}

// MomentCommentChildListView 获取某条评论下的全部子评论(分页)
func (MomentApi) MomentCommentChildListView(c *gin.Context) {
	var req MomentCommentChildListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	list, count, capped, err := content_service.ListMomentChildComments(req.Root, req.Page, req.Limit, req.Key, req.Order, req.EndId)
	if err != nil {
		response.FailWithMsg("查询出错", c)
		return
	}
	response.OkWithListCapped(list, count, capped, c)
}

// MomentCommentDeleteView 删除动态评论(本人/管理员)
func (MomentApi) MomentCommentDeleteView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claims := jwts.GetClaims(c)
	isAdmin := claims.Role == enum.AdminRole
	if err := content_service.DeleteMomentComment(req.ID, claims.UserID, isAdmin); err != nil {
		if st, ok := status.FromError(err); ok {
			response.FailWithMsg(st.Message(), c)
		} else {
			response.FailWithMsg("删除评论失败", c)
		}
		return
	}
	response.OkWithMsg("删除评论成功", c)
}

// MomentCommentDiggView 评论点赞/取消
func (MomentApi) MomentCommentDiggView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claims := jwts.GetClaims(c)
	digged, diggCount, err := content_service.ToggleMomentCommentDigg(req.ID, claims.UserID)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			response.FailWithMsg(st.Message(), c)
		} else {
			response.FailWithMsg("操作失败", c)
		}
		return
	}
	response.OkWithData(gin.H{"digged": digged, "diggCount": diggCount}, c)
}
