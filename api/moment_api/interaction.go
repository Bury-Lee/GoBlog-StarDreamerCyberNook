// api/moment_api/interaction.go
// 动态的点赞与转发(经 content 服务)
package moment_api

import (
	"strings"

	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/content_service"
	xss_filter "StarDreamerCyberNook/utils/XSSfilter"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
)

// MomentDiggView 点赞/取消点赞动态,返回最新状态与点赞数
func (MomentApi) MomentDiggView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claims := jwts.GetClaims(c)

	digged, likeCount, err := content_service.ToggleMomentDigg(req.ID, claims.UserID)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			response.FailWithMsg(st.Message(), c)
		} else {
			response.FailWithMsg("操作失败", c)
		}
		return
	}
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
	_ = c.ShouldBindJSON(&body)

	viewerID, isAdmin := viewerOf(c)
	if body.Visibility < models.MomentVisibilityPublic || body.Visibility > models.MomentVisibilityPrivate {
		body.Visibility = models.MomentVisibilityPublic
	}
	claims := jwts.GetClaims(c)
	content := strings.TrimSpace(xss_filter.SanitizeText(body.Content))

	model, err := content_service.RepostMoment(req.ID, claims.UserID, viewerID, isAdmin, content, body.Visibility)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			response.FailWithMsg(st.Message(), c)
		} else {
			response.FailWithMsg("转发失败", c)
		}
		return
	}
	response.OkWithData(model, c)
}
