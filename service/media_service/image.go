package media_service

import (
	"context"
	"time"

	mediav1 "StarDreamerCyberNook/gen/media/v1"
	"StarDreamerCyberNook/models"
)

func fromProtoImage(m *mediav1.ImageMeta) models.ImageModel {
	img := models.ImageModel{Filename: m.GetFilename(), Path: m.GetPath(), Size: m.GetSize(), Hash: m.GetHash()}
	img.ID = uint(m.GetId())
	img.CreatedAt = time.UnixMilli(m.GetCreatedAt())
	img.UpdatedAt = time.UnixMilli(m.GetUpdatedAt())
	return img
}

// ListImages 图片元数据分页列表。
func ListImages(page, limit int, key, order string, endID uint) ([]models.ImageModel, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), mediaTimeout)
	defer cancel()
	rep, err := c.ListImages(ctx, &mediav1.ListImagesRequest{
		Page: int32(page), Limit: int32(limit), Key: key, Order: order, EndId: uint64(endID),
	})
	if err != nil {
		return nil, 0, false, err
	}
	out := make([]models.ImageModel, 0, len(rep.GetList()))
	for _, m := range rep.GetList() {
		out = append(out, fromProtoImage(m))
	}
	return out, int(rep.GetCount()), rep.GetCapped(), nil
}

// GetImageMeta 单张图片元数据。
func GetImageMeta(id uint) (models.ImageModel, error) {
	c, err := client()
	if err != nil {
		return models.ImageModel{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), mediaTimeout)
	defer cancel()
	rep, err := c.GetImageMeta(ctx, &mediav1.GetImageMetaRequest{Id: uint64(id)})
	if err != nil {
		return models.ImageModel{}, err
	}
	return fromProtoImage(rep), nil
}

// ImageExistsByHash 按哈希查重,返回是否存在及已有图片ID。
func ImageExistsByHash(hash string) (bool, uint, error) {
	c, err := client()
	if err != nil {
		return false, 0, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), mediaTimeout)
	defer cancel()
	rep, err := c.FindImageByHash(ctx, &mediav1.FindImageByHashRequest{Hash: hash})
	if err != nil {
		return false, 0, err
	}
	return rep.GetFound(), uint(rep.GetId()), nil
}

// SaveImage 落库图片元数据,返回图片ID。
func SaveImage(filename, path string, size int64, hash string) (uint, error) {
	c, err := client()
	if err != nil {
		return 0, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), mediaTimeout)
	defer cancel()
	rep, err := c.SaveImage(ctx, &mediav1.SaveImageRequest{Filename: filename, Path: path, Size: size, Hash: hash})
	if err != nil {
		return 0, err
	}
	return uint(rep.GetId()), nil
}

// RemoveImages 删除图片(记录+对象),返回删除数与失败数。
func RemoveImages(ids []uint) (deleted, failed int, err error) {
	c, err := client()
	if err != nil {
		return 0, 0, err
	}
	req := &mediav1.RemoveImagesRequest{}
	for _, id := range ids {
		req.Ids = append(req.Ids, uint64(id))
	}
	ctx, cancel := context.WithTimeout(context.Background(), mediaTimeout)
	defer cancel()
	rep, err := c.RemoveImages(ctx, req)
	if err != nil {
		return 0, 0, err
	}
	return int(rep.GetDeleted()), int(rep.GetFailed()), nil
}
