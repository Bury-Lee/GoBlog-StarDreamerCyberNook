package internal

import (
	"context"
	"errors"

	"StarDreamerCyberNook/common"
	mediav1 "StarDreamerCyberNook/gen/media/v1"
	"StarDreamerCyberNook/models"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func toProtoImage(m models.ImageModel) *mediav1.ImageMeta {
	return &mediav1.ImageMeta{
		Id: uint64(m.ID), Filename: m.Filename, Path: m.Path, Size: m.Size, Hash: m.Hash,
		CreatedAt: m.CreatedAt.UnixMilli(), UpdatedAt: m.UpdatedAt.UnixMilli(),
	}
}

// ListImages 图片元数据分页列表。
func (s *Server) ListImages(_ context.Context, req *mediav1.ListImagesRequest) (*mediav1.ImageListReply, error) {
	list, count, capped, err := common.ListQuery[models.ImageModel](s.db, models.ImageModel{}, common.Options{
		PageInfo:      common.PageInfo{Page: int(req.GetPage()), Limit: int(req.GetLimit()), Order: req.GetOrder(), EndId: uint(req.GetEndId())},
		Likes:         []string{"filename"},
		AllowedOrders: []string{"id", "created_at", "size"},
		CountCap:      common.DefaultCountCap,
	})
	if err != nil {
		return nil, err
	}
	rep := &mediav1.ImageListReply{Count: int64(count), Capped: capped}
	for _, m := range list {
		rep.List = append(rep.List, toProtoImage(m))
	}
	return rep, nil
}

// GetImageMeta 单张图片元数据。
func (s *Server) GetImageMeta(_ context.Context, req *mediav1.GetImageMetaRequest) (*mediav1.ImageMeta, error) {
	var m models.ImageModel
	if err := s.db.Take(&m, uint(req.GetId())).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "图片已被删除")
		}
		return nil, err
	}
	return toProtoImage(m), nil
}

// FindImageByHash 按哈希查重。
func (s *Server) FindImageByHash(_ context.Context, req *mediav1.FindImageByHashRequest) (*mediav1.FindImageByHashReply, error) {
	var m models.ImageModel
	if err := s.db.Take(&m, "hash = ?", req.GetHash()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &mediav1.FindImageByHashReply{Found: false}, nil
		}
		return nil, err
	}
	return &mediav1.FindImageByHashReply{Found: true, Id: uint64(m.ID)}, nil
}

// SaveImage 落库图片元数据。
func (s *Server) SaveImage(_ context.Context, req *mediav1.SaveImageRequest) (*mediav1.SaveImageReply, error) {
	m := models.ImageModel{Filename: req.GetFilename(), Path: req.GetPath(), Size: req.GetSize(), Hash: req.GetHash()}
	if err := s.db.Create(&m).Error; err != nil {
		return nil, err
	}
	return &mediav1.SaveImageReply{Id: uint64(m.ID)}, nil
}

// RemoveImages 删除图片:先删存储对象,全部成功后再删库(任一失败则保留库,交调用方重试)。
func (s *Server) RemoveImages(ctx context.Context, req *mediav1.RemoveImagesRequest) (*mediav1.RemoveImagesReply, error) {
	ids := make([]uint, 0, len(req.GetIds()))
	for _, id := range req.GetIds() {
		ids = append(ids, uint(id))
	}
	rep := &mediav1.RemoveImagesReply{}
	if len(ids) == 0 {
		return rep, nil
	}
	var list []models.ImageModel
	if err := s.db.Find(&list, "id IN ?", ids).Error; err != nil {
		return nil, err
	}
	failed := 0
	for _, m := range list {
		if _, err := s.Delete(ctx, &mediav1.DeleteRequest{Key: m.Path}); err != nil {
			failed++
		}
	}
	if failed > 0 {
		rep.Failed = int32(failed)
		return rep, nil
	}
	if err := s.db.Delete(&list).Error; err != nil {
		return nil, err
	}
	rep.Deleted = int32(len(list))
	return rep, nil
}
