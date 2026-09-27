package image_api

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"path"

	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/log_service"
	"StarDreamerCyberNook/service/media_service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

//注:以后图片的操作都应该使用ID来查改,谁这么神人用哈希值来辨认

// ImageApi 图片管理API结构体
type ImageApi struct{}

// ImageListResponse 图片列表响应结构体
type ImageListResponse struct {
	models.ImageModel
	WebPath string `json:"webPath"` // Web访问路径
}

// ImageList 获取图片列表
func (ImageApi) ImageList(c *gin.Context) {
	var req common.PageInfo
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	_list, count, capped, err := media_service.ListImages(req.Page, req.Limit, req.Key, req.Order, req.EndId)
	if err != nil {
		response.FailWithMsg("查询失败", c)
		return
	}

	var list = make([]ImageListResponse, 0)
	for _, model := range _list {
		list = append(list, ImageListResponse{
			ImageModel: model,
			WebPath:    model.WebPath(),
		})
	}
	response.OkWithListCapped(list, count, capped, c)
}

// RemoveRequest 图片删除请求结构体
type RemoveRequest struct {
	IDlist []uint `json:"IDlist" binding:"required"`
}

// ImageRemoveView 批量删除图片
func (ImageApi) ImageRemoveView(c *gin.Context) {
	var req RemoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(err, c)
		return
	}

	log := log_service.GetLog(c)
	log.ShowRequest()
	log.ShowResponse()

	// 记录与存储对象均由 media 服务删除
	deleted, failed, err := media_service.RemoveImages(req.IDlist)
	if err != nil {
		response.FailWithMsg("删除失败", c)
		return
	}
	if failed > 0 {
		logrus.Errorf("有%d张图片对象删除失败", failed)
		response.FailWithMsg(fmt.Sprintf("有%d张图片删除失败,请重试", failed), c)
		return
	}
	response.OkWithMsg(fmt.Sprintf("图片删除成功,共删除%d张", deleted), c)
}

// GetImage 获取图片文件
func (ImageApi) GetImage(c *gin.Context) {
	id := c.Query("id")

	var imgID uint
	if _, err := fmt.Sscanf(id, "%d", &imgID); err != nil {
		response.FailWithMsg("图片已被删除", c)
		return
	}
	meta, err := media_service.GetImageMeta(imgID)
	if err != nil {
		response.FailWithMsg("图片已被删除", c)
		return
	}

	data, err := media_service.GetBytes(context.Background(), meta.Path)
	if err != nil {
		response.FailWithMsg("图片已被删除", c)
		return
	}

	contentType := mime.TypeByExtension(path.Ext(meta.Path))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Data(http.StatusOK, contentType, data)
}
