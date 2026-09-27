package OSS_img_api

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"

	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/log_service"
	"StarDreamerCyberNook/service/media_service"
	"StarDreamerCyberNook/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ImageListResponse 图片列表响应结构体
type ImageListResponse struct {
	models.ImageModel
	WebPath string `json:"webPath"` // Web访问路径
}

// RemoveRequest 图片删除请求结构体
type RemoveRequest struct {
	IDlist []uint `json:"IDlist" binding:"required"`
}

// ImageUploadView 上传图片(存储与元数据均经 media 服务)。
func (OSSImgApi) ImageUploadView(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.FailWithError(err, c)
		return
	}

	s := global.Config.Upload.Size
	if fileHeader.Size > s*1024*1024 {
		response.FailWithMsg(fmt.Sprintf("文件大小大于%dMB", s), c)
		return
	}

	filename := fileHeader.Filename
	suffix, ok := utils.ImageSuffixJudge(filename)
	if !ok {
		response.FailWithMsg("文件名非法:"+filename, c)
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	byteData, err := io.ReadAll(file)
	file.Close()
	if err != nil {
		response.FailWithError(err, c)
		return
	}

	if !utils.ImageContentJudge(suffix, byteData) {
		response.FailWithMsg("文件内容与图片格式不符", c)
		return
	}

	hash := utils.Md5(byteData)
	if found, id, err := media_service.ImageExistsByHash(hash); err == nil && found {
		logrus.Infof("上传图片重复 %s hash=%s id=%d", filename, hash, id)
		response.Ok(id, "上传成功", c)
		return
	}

	objectName := path.Join(global.Config.Upload.UploadDir, fmt.Sprintf("%s.%s", hash, suffix))
	if err := media_service.Put(context.Background(), objectName, byteData, utils.GetContentType(suffix)); err != nil {
		response.FailWithError(fmt.Errorf("上传失败: %v", err), c)
		return
	}

	id, err := media_service.SaveImage(filename, objectName, fileHeader.Size, hash)
	if err != nil {
		_ = media_service.Remove(context.Background(), objectName)
		response.FailWithError(err, c)
		return
	}
	response.Ok(id, "图片上传成功", c)
}

// GetImage 查询图片(从 media 服务读取)。
func (OSSImgApi) GetImage(c *gin.Context) {
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
	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, meta.Filename))
	c.Data(http.StatusOK, contentType, data)
}

// ImageList 管理员查询图片列表
func (OSSImgApi) ImageList(c *gin.Context) {
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

// ImageRemoveView 管理员批量删除图片
func (OSSImgApi) ImageRemoveView(c *gin.Context) {
	var req RemoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	log := log_service.GetLog(c)
	log.ShowRequest()
	log.ShowResponse()

	if len(req.IDlist) == 0 {
		response.OkWithMsg("图片删除成功,共删除0张", c)
		return
	}
	if len(req.IDlist) > 100 {
		response.FailWithMsg("单次最多删除100张图片", c)
		return
	}

	deleted, failed, err := media_service.RemoveImages(req.IDlist)
	if err != nil {
		logrus.Errorf("删除图片失败:%s", err)
		response.FailWithMsg("删除失败", c)
		return
	}
	if failed > 0 {
		response.FailWithMsg(fmt.Sprintf("有%d张图片删除失败,请重试", failed), c)
		return
	}

	response.OkWithMsg(fmt.Sprintf("图片删除成功,共删除%d张", deleted), c)
}
