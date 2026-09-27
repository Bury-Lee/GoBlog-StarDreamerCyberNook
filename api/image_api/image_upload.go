// api/image_api/image_upload.go
package image_api

import (
	"context"
	"fmt"
	"io"
	"path"

	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/service/media_service"
	"StarDreamerCyberNook/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (ImageApi) ImageUploadView(c *gin.Context) {
	//前端看这里,图片上传完成后会返回一个图片的ID,这个ID可以用来访问图片,访问图片的接口是 /api/image/:id
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	// 文件大小判断
	s := global.Config.Upload.Size
	if fileHeader.Size > s*1024*1024 {
		response.FailWithMsg(fmt.Sprintf("文件大小大于%dMB", s), c)
		return
	}
	// 后缀判断
	filename := fileHeader.Filename
	suffix, ok := utils.ImageSuffixJudge(filename)
	if !ok {
		response.FailWithMsg("文件名非法:"+filename, c)
		return
	}
	// 读取文件内容
	file, err := fileHeader.Open()
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	defer file.Close()
	byteData, err := io.ReadAll(file)
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	// 魔数校验:文件内容必须与后缀匹配,防止伪造图片
	if !utils.ImageContentJudge(suffix, byteData) {
		response.FailWithMsg("文件内容与图片格式不符", c)
		return
	}
	hash := utils.Md5(byteData)
	// 判断这个hash有没有(元数据在 media 服务)
	if found, id, err := media_service.ImageExistsByHash(hash); err == nil && found {
		logrus.Infof("上传图片重复 %s hash=%s id=%d", filename, hash, id)
		response.Ok(id, "上传成功", c)
		return
	}
	// 存储键(本地/对象存储由 media 服务决定)
	filePath := path.Join(global.Config.Upload.UploadDir, fmt.Sprintf("%s.%s", hash, suffix))
	if err := media_service.Put(context.Background(), filePath, byteData, utils.GetContentType(suffix)); err != nil {
		response.FailWithError(fmt.Errorf("存储图片失败: %v", err), c)
		return
	}
	// 入库(media 服务);失败则清理已存对象
	id, err := media_service.SaveImage(filename, filePath, fileHeader.Size, hash)
	if err != nil {
		_ = media_service.Remove(context.Background(), filePath)
		response.FailWithError(err, c)
		return
	}
	response.Ok(id, "图片上传成功", c)
}
