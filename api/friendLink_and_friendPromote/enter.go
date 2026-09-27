package friendlink_and_friendpromote

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/service/community_service"
	jwts "StarDreamerCyberNook/utils/jwts"
	"fmt"

	"github.com/gin-gonic/gin"
)

type FriendApi struct{}

type FriendLinkCreateRequest struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	Logo      string `json:"logo"`
	IsShow    bool   `json:"is_show"`
	SortOrder int    `json:"sort_order"`
	Remark    string `json:"remark"`
}

type FriendPromotionCreateRequest struct {
	Title         string   `json:"title"`
	FriendName    string   `json:"friend_name"`
	Avatar        string   `json:"avatar"`
	Category      string   `json:"category"`
	Description   string   `json:"description"`
	PreviewImages string   `json:"preview_images"`
	ContactInfo   []string `json:"contact_info"`
	IsShow        bool     `json:"is_show"`
	SortOrder     int      `json:"sort_order"`
	Position      string   `json:"position"`
	Remark        string   `json:"remark"`
}

func (FriendApi) FriendLinkCreateView(c *gin.Context) {
	var req FriendLinkCreateRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数绑定失败", c)
		return
	}
	if err := community_service.CreateFriendLink(req.Name, req.URL, req.Logo, req.IsShow, req.SortOrder, req.Remark); err != nil {
		response.FailWithMsg("添加失败", c)
		return
	}
	response.OkWithMsg("添加成功", c)
}

func (FriendApi) FriendLinkListView(c *gin.Context) {
	var req common.PageInfo
	c.ShouldBind(&req)

	//管理员传 all=1 时返回全部数据(含已隐藏),否则隐藏后条目在后台无法再找到
	all := false
	if c.Query("all") == "1" {
		if claims, err := jwts.ParseTokenByGin(c); err == nil && claims.Role == enum.AdminRole {
			all = true
		}
	}

	list, count, capped, _ := community_service.ListFriendLinks(all, req.Page, req.Limit, req.Order, req.EndId)
	response.OkWithListCapped(list, count, capped, c)
}

func (FriendApi) FriendLinkRemoveView(c *gin.Context) {
	var req models.RemoveRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	if err := community_service.RemoveFriendLinks(req.IDList); err != nil {
		response.FailWithMsg("删除失败", c)
		return
	}
	response.OkWithMsg(fmt.Sprintf("成功删除%d个", len(req.IDList)), c)
}

func (FriendApi) FriendLinkUpdateView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("绑定参数失败", c)
		return
	}
	var data FriendLinkCreateRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		response.FailWithMsg("绑定参数失败", c)
		return
	}
	if err := community_service.UpdateFriendLink(req.ID, data.Name, data.URL, data.Logo, data.IsShow, data.SortOrder, data.Remark); err != nil {
		response.FailWithMsg("更新失败", c)
	} else {
		response.OkWithMsg("更新成功", c)
	}
}

func (FriendApi) FriendPromotionListView(c *gin.Context) {
	var req common.PageInfo
	c.ShouldBind(&req)

	//管理员传 all=1 时返回全部数据(含已隐藏)
	all := false
	if c.Query("all") == "1" {
		if claims, err := jwts.ParseTokenByGin(c); err == nil && claims.Role == enum.AdminRole {
			all = true
		}
	}

	list, count, capped, _ := community_service.ListFriendPromotions(all, req.Page, req.Limit, req.Order, req.EndId)
	response.OkWithListCapped(list, count, capped, c)
}

func (FriendApi) FriendPromotionRemoveView(c *gin.Context) {
	var req models.RemoveRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	if err := community_service.RemoveFriendPromotions(req.IDList); err != nil {
		response.FailWithMsg("删除失败", c)
		return
	}
	response.OkWithMsg(fmt.Sprintf("成功删除%d个", len(req.IDList)), c)
}

func (FriendApi) FriendPromotionUpdateView(c *gin.Context) {
	var idReq models.IDRequest
	if err := c.ShouldBindUri(&idReq); err != nil {
		response.FailWithMsg("绑定参数失败", c)
		return
	}
	var updateReq FriendPromotionCreateRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response.FailWithMsg("绑定参数失败", c)
		return
	}
	if err := community_service.UpdateFriendPromotion(idReq.ID, updateReq.Title, updateReq.FriendName, updateReq.Avatar,
		updateReq.Category, updateReq.Description, updateReq.PreviewImages, updateReq.ContactInfo,
		updateReq.IsShow, updateReq.SortOrder, updateReq.Position, updateReq.Remark); err != nil {
		response.FailWithMsg("更新失败", c)
	} else {
		response.OkWithMsg("更新成功", c)
	}
}

func (FriendApi) FriendPromotionCreateView(c *gin.Context) {
	var req FriendPromotionCreateRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数绑定失败", c)
		return
	}
	if err := community_service.CreateFriendPromotion(req.Title, req.FriendName, req.Avatar, req.Category,
		req.Description, req.PreviewImages, req.ContactInfo, req.IsShow, req.SortOrder, req.Position, req.Remark); err != nil {
		response.FailWithMsg("添加失败", c)
		return
	}
	response.OkWithMsg("添加成功", c)
}
