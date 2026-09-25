// api/moment_api/comment.go
// 动态评论:发表、列表、子评论、删除、点赞
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
	utils_other "StarDreamerCyberNook/utils/other"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type MomentCommentCreateRequest struct {
	Content  string `json:"content" binding:"required"`  // 评论内容
	MomentID uint   `json:"momentID" binding:"required"` // 动态ID
	ParentID uint   `json:"parentID"`                    // 父评论ID
}

// MomentCommentCreateView 发表评论/回复
func (MomentApi) MomentCommentCreateView(c *gin.Context) {
	var req MomentCommentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	var moment models.MomentModel
	if err := global.DB.Take(&moment, req.MomentID).Error; err != nil {
		response.FailWithMsg("动态不存在", c)
		return
	}
	viewerID, isAdmin := viewerOf(c)
	if !canViewMoment(moment, viewerID, isAdmin) {
		response.FailWithMsg("动态不存在", c)
		return
	}
	content := xss_filter.NewXSSFilter().Sanitize(req.Content)
	if content == "" {
		response.FailWithMsg("评论内容不能为空", c)
		return
	}

	//AI审核:受 ai.enable + site.moment.enableExamination 控制,逻辑同文章/评论审核
	if global.Config.AI.Enable && global.Config.Site.Moment.EnableExamination {
		reply, err := ai_service.CreateSingleReply(
			"评论内容:"+content,
			global.SystemPromptComment.String(),
		)
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
			logrus.Errorf("动态评论AI审核出错,回复内容:%s,评论内容:%s", reply, content)
			return
		}
	}

	claims := jwts.GetClaims(c)
	model := models.MomentCommentModel{
		MomentID: req.MomentID,
		UserID:   claims.UserID,
		Content:  content,
	}
	if req.ParentID != 0 {
		var parent models.MomentCommentModel
		if err := global.DB.Take(&parent, "id = ? and moment_id = ?", req.ParentID, req.MomentID).Error; err != nil {
			response.FailWithMsg("评论不存在", c)
			return
		}
		model.ParentPath = utils_other.EncodePath(parent.ParentPath, parent.ID)
		if parent.RootParentID == nil {
			model.RootParentID = &parent.ID
		} else {
			model.RootParentID = parent.RootParentID
		}
	}

	//评论与动态评论数放在同一事务
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		if e := tx.Create(&model).Error; e != nil {
			return e
		}
		return tx.Model(&models.MomentModel{}).Where("id = ?", req.MomentID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
	}); err != nil {
		response.FailWithMsg("评论失败", c)
		return
	}

	response.OkWithData(model, c)
}

// MomentCommentListRequest 一级评论列表请求
type MomentCommentListRequest struct {
	common.PageInfo
	MomentID uint `form:"momentID" binding:"required"` // 动态ID
}

// MomentCommentListView 获取动态的一级评论(分页)
func (MomentApi) MomentCommentListView(c *gin.Context) {
	var req MomentCommentListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	var moment models.MomentModel
	if err := global.DB.Take(&moment, req.MomentID).Error; err != nil {
		response.FailWithMsg("动态不存在", c)
		return
	}
	viewerID, isAdmin := viewerOf(c)
	if !canViewMoment(moment, viewerID, isAdmin) {
		response.FailWithMsg("动态不存在", c)
		return
	}

	options := common.Options{
		PageInfo:      req.PageInfo,
		Preloads:      []string{"UserModel"},
		Where:         global.DB.Where("moment_id = ? and root_parent_id is null", req.MomentID),
		DefaultOrder:  "created_at desc",
		AllowedOrders: []string{"id", "created_at", "digg_count"},
		CountCap:      common.DefaultCountCap,
	}
	list, count, capped, err := common.ListQuery(models.MomentCommentModel{}, options)
	if err != nil {
		response.FailWithMsg("查询评论失败", c)
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
	}
	response.OkWithListCapped(list, count, capped, c)
}

// MomentCommentChildListRequest 子评论列表请求
type MomentCommentChildListRequest struct {
	common.PageInfo
	Root uint `form:"root" binding:"required"` // 根评论ID
}

// MomentCommentChildListView 获取某条评论下的全部子评论(分页)
func (MomentApi) MomentCommentChildListView(c *gin.Context) {
	var req MomentCommentChildListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	var root models.MomentCommentModel
	if err := global.DB.Take(&root, req.Root).Error; err != nil {
		response.FailWithMsg("评论不存在", c)
		return
	}
	path := utils_other.EncodePath(root.ParentPath, root.ID)
	options := common.Options{
		PageInfo: req.PageInfo,
		Preloads: []string{"UserModel"},
		Where: global.DB.Where(
			"moment_id = ? and (parent_path = ? or parent_path like ?)",
			root.MomentID, path, path+"/%",
		),
		DefaultOrder:  "created_at desc",
		AllowedOrders: []string{"id", "created_at", "digg_count"},
		CountCap:      common.DefaultCountCap,
	}
	list, count, capped, err := common.ListQuery(models.MomentCommentModel{}, options)
	if err != nil {
		response.FailWithMsg("查询出错", c)
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
	var comment models.MomentCommentModel
	if err := global.DB.Take(&comment, req.ID).Error; err != nil {
		response.FailWithMsg("评论不存在", c)
		return
	}
	claims := jwts.GetClaims(c)
	if claims.UserID != comment.UserID && claims.Role != enum.AdminRole {
		response.FailWithMsg("没有权限删除该评论", c)
		return
	}
	//TODO:删除后应该发消息通知

	if comment.RootParentID == nil {
		//一级评论:连同子评论及其点赞一起删除,全部放在同一事务
		var childCount int64
		if err := global.DB.Model(&models.MomentCommentModel{}).Where("root_parent_id = ?", comment.ID).Count(&childCount).Error; err != nil {
			response.FailWithMsg("删除评论失败", c)
			return
		}
		if err := global.DB.Transaction(func(tx *gorm.DB) error {
			//先删该评论树(自身+子评论)的点赞,需在删评论前拿得到评论ID
			childIDs := tx.Model(&models.MomentCommentModel{}).Select("id").Where("root_parent_id = ?", comment.ID)
			if e := tx.Where("moment_comment_id = ? or moment_comment_id in (?)", comment.ID, childIDs).
				Delete(&models.MomentCommentDiggModel{}).Error; e != nil {
				return e
			}
			if e := tx.Delete(&models.MomentCommentModel{}, "root_parent_id = ?", comment.ID).Error; e != nil {
				return e
			}
			if e := tx.Delete(&comment).Error; e != nil {
				return e
			}
			return tx.Model(&models.MomentModel{}).Where("id = ? and comment_count >= ?", comment.MomentID, childCount+1).
				UpdateColumn("comment_count", gorm.Expr("comment_count - ?", childCount+1)).Error
		}); err != nil {
			response.FailWithMsg("删除评论失败", c)
			return
		}
	} else {
		//子评论:删自身及其点赞,按影响行数决定是否 -1,避免漂移
		if err := global.DB.Transaction(func(tx *gorm.DB) error {
			res := tx.Delete(&comment)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return nil
			}
			if e := tx.Where("moment_comment_id = ?", comment.ID).Delete(&models.MomentCommentDiggModel{}).Error; e != nil {
				return e
			}
			return tx.Model(&models.MomentModel{}).Where("id = ? and comment_count > 0", comment.MomentID).
				UpdateColumn("comment_count", gorm.Expr("comment_count - 1")).Error
		}); err != nil {
			response.FailWithMsg("删除评论失败", c)
			return
		}
	}
	response.OkWithMsg("删除评论成功", c)
}

// 此处我们默认评论的点赞和取消是低频的事件,因此我们不使用Redis而直接改数据库
// MomentCommentDiggView 评论点赞/取消
func (MomentApi) MomentCommentDiggView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	var comment models.MomentCommentModel
	if err := global.DB.Take(&comment, req.ID).Error; err != nil {
		response.FailWithMsg("评论不存在", c)
		return
	}
	claims := jwts.GetClaims(c)

	//点赞关系与计数同一事务,取消时按影响行数决定是否 -1,避免漂移
	digged := false
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		takeErr := tx.Take(&models.MomentCommentDiggModel{}, "user_id = ? and moment_comment_id = ?", claims.UserID, req.ID).Error
		switch {
		case errors.Is(takeErr, gorm.ErrRecordNotFound):
			if e := tx.Create(&models.MomentCommentDiggModel{UserID: claims.UserID, MomentCommentID: req.ID}).Error; e != nil {
				return e
			}
			if e := tx.Model(&models.MomentCommentModel{}).Where("id = ?", req.ID).
				UpdateColumn("digg_count", gorm.Expr("digg_count + 1")).Error; e != nil {
				return e
			}
			digged = true
		case takeErr != nil:
			return takeErr
		default:
			res := tx.Where("user_id = ? and moment_comment_id = ?", claims.UserID, req.ID).Delete(&models.MomentCommentDiggModel{})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected > 0 {
				if e := tx.Model(&models.MomentCommentModel{}).Where("id = ? and digg_count > 0", req.ID).
					UpdateColumn("digg_count", gorm.Expr("digg_count - 1")).Error; e != nil {
					return e
				}
			}
		}
		return nil
	})
	if err != nil {
		response.FailWithMsg("操作失败", c)
		return
	}

	var diggCount int
	global.DB.Model(&models.MomentCommentModel{}).Where("id = ?", req.ID).Select("digg_count").Scan(&diggCount)
	response.OkWithData(gin.H{"digged": digged, "diggCount": diggCount}, c)
}
