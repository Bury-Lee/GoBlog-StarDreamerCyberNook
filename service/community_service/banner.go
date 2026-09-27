package community_service

import (
	"context"
	"time"

	communityv1 "StarDreamerCyberNook/gen/community/v1"
	"StarDreamerCyberNook/models"
)

// CreateBanner 新增轮播图。
func CreateBanner(cover, href string, isShow bool) error {
	return call(func(ctx context.Context, c communityv1.CommunityServiceClient) error {
		_, err := c.CreateBanner(ctx, &communityv1.BannerWrite{Cover: cover, Href: href, IsShow: isShow})
		return err
	})
}

// ListBanners 轮播图列表;all=true 时包含隐藏项(仅管理员)。
func ListBanners(all bool, page, limit int, order string, endID uint) ([]models.BannerModel, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), communityTimeout)
	defer cancel()
	rep, err := c.ListBanners(ctx, &communityv1.ListRequest{
		All: all, Page: int32(page), Limit: int32(limit), Order: order, EndId: uint64(endID),
	})
	if err != nil {
		return nil, 0, false, err
	}
	list := make([]models.BannerModel, 0, len(rep.GetList()))
	for _, b := range rep.GetList() {
		m := models.BannerModel{IsShow: b.GetIsShow(), Cover: b.GetCover(), Href: b.GetHref()}
		m.ID = uint(b.GetId())
		m.CreatedAt = time.UnixMilli(b.GetCreatedAt())
		m.UpdatedAt = time.UnixMilli(b.GetUpdatedAt())
		list = append(list, m)
	}
	return list, int(rep.GetCount()), rep.GetCapped(), nil
}

// RemoveBanners 批量删除轮播图。
func RemoveBanners(ids []uint) error {
	return call(func(ctx context.Context, c communityv1.CommunityServiceClient) error {
		req := &communityv1.IDs{}
		for _, id := range ids {
			req.Ids = append(req.Ids, uint64(id))
		}
		_, err := c.RemoveBanners(ctx, req)
		return err
	})
}

// UpdateBanner 更新轮播图。
func UpdateBanner(id uint, cover, href string, isShow bool) error {
	return call(func(ctx context.Context, c communityv1.CommunityServiceClient) error {
		_, err := c.UpdateBanner(ctx, &communityv1.UpdateBannerRequest{
			Id: uint64(id), Cover: cover, Href: href, IsShow: isShow,
		})
		return err
	})
}
