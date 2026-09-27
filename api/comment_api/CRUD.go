package comment_api

import (
	"fmt"

	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/service/ai_service"
	"StarDreamerCyberNook/service/content_service"
	"StarDreamerCyberNook/service/message_service"
	"StarDreamerCyberNook/service/redis_service/redis_count"
	xss_filter "StarDreamerCyberNook/utils/XSSfilter"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/status"
)

// 注:评论没有修改这一说,只有增删查
type CommentCreateRequest struct {
	Content   string `json:"content" binding:"required"`
	ArticleID uint   `json:"articleID" binding:"required"`
	ParentID  uint   `json:"parentID"` // 父评论ID
}

// CommentCreateView 创建评论:AI 审核与消息/计数在网关,校验与落库在 content 服务。
func (CommentApi) CommentCreateView(c *gin.Context) {
	var req CommentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claims := jwts.GetClaims(c)

	//评论内容防xss注入
	req.Content = xss_filter.NewXSSFilter().Sanitize(req.Content)
	if req.Content == "" {
		response.FailWithMsg("评论内容不能为空", c)
		return
	}

	if global.Config.AI.Enable { //启用ai审核
		reply, err := ai_service.CreateSingleReply("评论内容:"+req.Content, global.SystemPromptComment.String())
		if err != nil {
			response.FailWithMsg("ai审核失败,已经自动创建为待审核状态: "+err.Error(), c)
			return
		}
		switch reply {
		case "通过":
			//通过就正常执行流程
		case "拒绝":
			response.FailWithMsg("评论可能含有违规信息,已拒绝", c)
			return
		default:
			logrus.Errorf("ai审核出错,回复内容:%s,评论详情:%s\n,评论不发布", reply, fmt.Sprintf("%#v", req))
			return
		}
	}

	res, err := content_service.CreateComment(claims.UserID, req.ArticleID, req.ParentID, req.Content)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			response.FailWithMsg(st.Message(), c)
		} else {
			response.FailWithMsg("发布评论失败", c)
		}
		return
	}

	//评论已落库,再发消息(消息需要评论ID)
	model := models.CommentModel{Model: models.Model{ID: res.ID}, Content: req.Content, UserID: claims.UserID, ArticleID: req.ArticleID}
	for _, revUserID := range res.ReplyRevUserIDs {
		if err := message_service.InsertReplyMessage(model, revUserID); err != nil {
			logrus.Error("系统消息发送失败:", err.Error())
		}
	}
	if res.ArticleUserID != claims.UserID {
		if err := message_service.InsertCommentMessage(model, res.ArticleUserID); err != nil {
			logrus.Error("系统消息发送失败:", err.Error())
		}
	}

	redis_count.SetCacheComment(req.ArticleID, true) //增量更新缓存
	response.OkWithMsg("发布评论成功", c)
}

// 分页获取指定文章的一级评论
type CommentDetailRequest struct {
	common.PageInfo
	ArticleID uint `form:"articleID" binding:"required"` // 文章ID
}

// CommentListlView 某文章的一级评论(经 content 服务)。
func (CommentApi) CommentListlView(c *gin.Context) {
	var req CommentDetailRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	List, count, capped, err := content_service.ListComments(req.ArticleID, req.Page, req.Limit, req.Key, req.Order, req.EndId)
	if err != nil {
		response.FailWithMsg("查询评论失败", c)
		return
	}

	//叠加Redis里的点赞增量
	if len(List) > 0 {
		ids := make([]uint, 0, len(List))
		for _, v := range List {
			ids = append(ids, v.ID)
		}
		diggMap := redis_count.GetAllCacheCommentDigg(ids)
		for i := range List {
			List[i].DiggCount += diggMap[List[i].ID]
		}
	}
	response.OkWithListCapped(List, count, capped, c)
}

// 分页获取某条评论下的子评论详情(多条)
type CommentListRequest struct {
	common.PageInfo
	Root uint `form:"root" binding:"required"` //根评论ID
}

// CommentChildListView 某根评论下的子评论(经 content 服务)。
func (CommentApi) CommentChildListView(c *gin.Context) {
	var req CommentListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	List, count, capped, err := content_service.ListChildComments(req.Root, req.Page, req.Limit, req.Key, req.Order, req.EndId)
	if err != nil {
		response.FailWithMsg("查询评论失败", c)
		return
	}

	if len(List) > 0 {
		ids := make([]uint, 0, len(List))
		for _, v := range List {
			ids = append(ids, v.ID)
		}
		diggMap := redis_count.GetAllCacheCommentDigg(ids)
		for i := range List {
			List[i].DiggCount += diggMap[List[i].ID]
		}
	}
	response.OkWithListCapped(List, count, capped, c)
}

// CommentDeleteView 删除评论(经 content 服务),评论数增量在网关调整。
func (CommentApi) CommentDeleteView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claim := jwts.GetClaims(c)
	isAdmin := claim.Role == enum.AdminRole

	res, err := content_service.DeleteComment(req.ID, claim.UserID, isAdmin)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			response.FailWithMsg(st.Message(), c)
		} else {
			response.FailWithMsg("删除评论失败", c)
		}
		return
	}
	if res.Ok {
		redis_count.SetCacheCommentBy(res.ArticleID, -res.Delta)
	}
	response.OkWithMsg("删除评论成功", c)
}

// CommentDiggView 评论点赞(经 content 服务),Redis 增量与返回值在网关。
func (CommentApi) CommentDiggView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claim := jwts.GetClaims(c)

	res, err := content_service.ToggleCommentDigg(req.ID, claim.UserID)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			response.FailWithMsg(st.Message(), c)
		} else {
			response.FailWithMsg("点赞失败", c)
		}
		return
	}
	redis_count.SetCacheCommentDigg(req.ID, res.Digged)

	diggCount := res.BaseDiggCount
	if delta, ok := redis_count.GetAllCacheCommentDigg([]uint{req.ID})[req.ID]; ok {
		diggCount += delta
	}
	if diggCount < 0 {
		diggCount = 0
	}
	response.OkWithData(gin.H{"digged": res.Digged, "diggCount": diggCount}, c)
}
