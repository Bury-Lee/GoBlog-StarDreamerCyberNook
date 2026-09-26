// api/moment_api/interaction.go
// 动态的点赞与转发
package moment_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	xss_filter "StarDreamerCyberNook/utils/XSSfilter"
	jwts "StarDreamerCyberNook/utils/jwts"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MomentDiggView 点赞/取消点赞动态,返回最新状态与点赞数
func (MomentApi) MomentDiggView(c *gin.Context) {
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

	//点赞关系与计数放在同一事务,取消时按影响行数决定是否 -1,避免并发下的计数漂移
	//计数暂用数据库原子自增:动态点赞预计低频;若日后变高频,再引入 Redis 增量 + cron 回写(也为微服务化预留)
	digged := false
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		takeErr := tx.Take(&models.MomentDiggModel{}, "user_id = ? and moment_id = ?", claims.UserID, req.ID).Error
		switch {
		case errors.Is(takeErr, gorm.ErrRecordNotFound):
			if e := tx.Create(&models.MomentDiggModel{UserID: claims.UserID, MomentID: req.ID}).Error; e != nil {
				return e
			}
			if e := tx.Model(&models.MomentModel{}).Where("id = ?", req.ID).
				UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error; e != nil {
				return e
			}
			digged = true
		case takeErr != nil:
			return takeErr
		default:
			res := tx.Where("user_id = ? and moment_id = ?", claims.UserID, req.ID).Delete(&models.MomentDiggModel{})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected > 0 {
				if e := tx.Model(&models.MomentModel{}).Where("id = ? and like_count > 0", req.ID).
					UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error; e != nil {
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

	var likeCount int
	global.DB.Model(&models.MomentModel{}).Where("id = ?", req.ID).Select("like_count").Scan(&likeCount)
	response.OkWithData(gin.H{"digged": digged, "likeCount": likeCount}, c)
}

type MomentRepostRequest struct {
	Content    string                  `json:"content"`    // 转发附带评论,可选
	Visibility models.MomentVisibility `json:"visibility"` // 可见性,默认公开
}

// MomentRepostView 转发动态:生成一条引用原动态的新动态
func (MomentApi) MomentRepostView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	var body MomentRepostRequest
	_ = c.ShouldBindJSON(&body) //转发可带可不带内容,绑定失败不阻断

	var source models.MomentModel
	if err := global.DB.Preload("UserModel").Take(&source, req.ID).Error; err != nil {
		response.FailWithMsg("动态不存在", c)
		return
	}
	viewerID, isAdmin := viewerOf(c)
	if !canViewMoment(source, viewerID, isAdmin) {
		response.FailWithMsg("动态不存在", c)
		return
	}
	//不能转发自己的私密/草稿以外?统一按可见性校验即可
	if body.Visibility < models.MomentVisibilityPublic || body.Visibility > models.MomentVisibilityPrivate {
		body.Visibility = models.MomentVisibilityPublic
	}

	claims := jwts.GetClaims(c)
	content := strings.TrimSpace(xss_filter.SanitizeText(body.Content))
	model := models.MomentModel{
		UserID:       claims.UserID,
		Type:         models.MomentTypeMoment,
		Visibility:   body.Visibility,
		Content:      content,
		RepostFromID: &source.ID,
		Status:       models.StatusPublished,
	}
	//转发记录与源动态转发数放在同一事务
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		if e := tx.Create(&model).Error; e != nil {
			return e
		}
		return tx.Model(&models.MomentModel{}).Where("id = ?", source.ID).
			UpdateColumn("repost_count", gorm.Expr("repost_count + 1")).Error
	}); err != nil {
		response.FailWithMsg("转发失败", c)
		return
	}

	//源动态作者与转发者都只回填公开字段,避免泄露邮箱等隐私
	source.UserModel = models.UserModel{
		Model:         source.UserModel.Model,
		NickName:      source.UserModel.NickName,
		Avatar:        source.UserModel.Avatar,
		LastLoginTime: source.UserModel.LastLoginTime,
		Age:           source.UserModel.Age,
		LikeTags:      source.UserModel.LikeTags,
	}
	var author models.UserModel
	if global.DB.Take(&author, claims.UserID).Error == nil {
		model.UserModel = models.UserModel{
			Model:         author.Model,
			NickName:      author.NickName,
			Avatar:        author.Avatar,
			LastLoginTime: author.LastLoginTime,
			Age:           author.Age,
			LikeTags:      author.LikeTags,
		}
	}
	model.RepostFrom = &source
	response.OkWithData(model, c)
}
