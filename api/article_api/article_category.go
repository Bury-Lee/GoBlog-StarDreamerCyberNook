package article_api

import (
	"fmt"

	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/service/content_service"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

type CategoryCreateRequest struct {
	ID    uint   `json:"id"` //id为0时为创建，否则为修改
	Title string `json:"title" binding:"required,max=32"`
}

// CategoryCreateView 创建/更新分类(经 content 服务)。
func (ArticleApi) CategoryCreateView(c *gin.Context) {
	var req CategoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("请求参数错误", c)
		return
	}
	claims := jwts.GetClaims(c)

	if err := content_service.CreateCategory(req.ID, req.Title, claims.UserID); err != nil {
		if req.ID == 0 {
			response.FailWithMsg("分类名称重复或创建失败", c)
		} else {
			response.FailWithMsg("更新分类错误", c)
		}
		return
	}
	if req.ID == 0 {
		response.OkWithMsg("创建分类成功", c)
		return
	}
	response.OkWithMsg("更新分类成功", c)
}

type CategoryListRequest struct {
	common.PageInfo
	UserID uint   `form:"userID"`
	Type   string `form:"type" binding:"required"` // self 查自己 other 查别人 admin 后台
}

type CategoryListResponse struct {
	models.CategoryModel
	ArticleCount int    `json:"articleCount"`
	Nickname     string `json:"nickname,omitempty"`
	Avatar       string `json:"avatar,omitempty"`
}

// CategoryListView 分类列表(经 content 服务)。
func (ArticleApi) CategoryListView(c *gin.Context) {
	var req CategoryListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg(err.Error(), c)
		return
	}

	withUser := false
	switch req.Type {
	case "self":
		claims, err := jwts.ParseTokenByGin(c)
		if err != nil {
			response.FailWithMsg("未登录", c)
			return
		}
		req.UserID = claims.UserID
	case "other":
		if req.UserID == 0 {
			response.FailWithMsg("用户ID不能为空", c)
			return
		}
	case "admin":
		claims, err := jwts.ParseTokenByGin(c)
		if err != nil {
			response.FailWithMsg("未登录", c)
			return
		}
		if claims.Role != enum.AdminRole {
			response.FailWithMsg("权限错误", c)
			return
		}
		withUser = true
	default:
		response.FailWithMsg("类型错误", c)
		return
	}

	items, count, capped, err := content_service.ListCategories(req.UserID, withUser, req.Page, req.Limit, req.Key, req.Order)
	if err != nil {
		response.FailWithMsg("查询失败", c)
		return
	}

	list := make([]CategoryListResponse, 0, len(items))
	for _, it := range items {
		list = append(list, CategoryListResponse{
			CategoryModel: models.CategoryModel{
				Model:  models.Model{ID: it.Category.ID, CreatedAt: it.Category.CreatedAt, UpdatedAt: it.Category.UpdatedAt},
				Title:  it.Category.Title,
				UserID: it.Category.UserID,
			},
			ArticleCount: it.ArticleCount,
			Nickname:     it.Nickname,
			Avatar:       it.Avatar,
		})
	}
	response.OkWithListCapped(list, count, capped, c)
}

// CategoryRemoveView 删除分类(经 content 服务)。
func (ArticleApi) CategoryRemoveView(c *gin.Context) {
	var req = models.RemoveRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("请求参数错误", c)
		return
	}
	claims := jwts.GetClaims(c)
	all := claims.Role == enum.AdminRole

	deleted, err := content_service.RemoveCategories(req.IDList, claims.UserID, all)
	if err != nil {
		response.FailWithMsg("删除分类失败", c)
		return
	}
	response.OkWithMsg(fmt.Sprintf("删除分类成功 共删除%d条", deleted), c)
}

// CategoryOptionsView 分类选项列表(经 content 服务)。
func (ArticleApi) CategoryOptionsView(c *gin.Context) {
	claims := jwts.GetClaims(c)

	opts, err := content_service.CategoryOptions(claims.UserID)
	if err != nil {
		response.FailWithMsg("查询失败", c)
		return
	}
	list := make([]models.OptionsResponse[uint], 0, len(opts))
	for _, o := range opts {
		list = append(list, models.OptionsResponse[uint]{Key: o.Label, Value: o.Value})
	}
	response.OkWithData(list, c)
}
