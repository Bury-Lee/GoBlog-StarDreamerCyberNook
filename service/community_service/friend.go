package community_service

import (
	"context"
	"time"

	communityv1 "StarDreamerCyberNook/gen/community/v1"
	"StarDreamerCyberNook/models"
)

// ---- 友链 ----

// CreateFriendLink 新增友链。
func CreateFriendLink(name, url, logo string, isShow bool, sortOrder int, remark string) error {
	return call(func(ctx context.Context, c communityv1.CommunityServiceClient) error {
		_, err := c.CreateFriendLink(ctx, &communityv1.FriendLinkWrite{
			Name: name, Url: url, Logo: logo, IsShow: isShow, SortOrder: int32(sortOrder), Remark: remark,
		})
		return err
	})
}

// ListFriendLinks 友链列表;all=true 时包含隐藏项(仅管理员)。
func ListFriendLinks(all bool, page, limit int, order string, endID uint) ([]models.FriendLink, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), communityTimeout)
	defer cancel()
	rep, err := c.ListFriendLinks(ctx, &communityv1.ListRequest{
		All: all, Page: int32(page), Limit: int32(limit), Order: order, EndId: uint64(endID),
	})
	if err != nil {
		return nil, 0, false, err
	}
	list := make([]models.FriendLink, 0, len(rep.GetList()))
	for _, f := range rep.GetList() {
		m := models.FriendLink{
			Name: f.GetName(), URL: f.GetUrl(), Logo: f.GetLogo(),
			IsShow: f.GetIsShow(), SortOrder: int(f.GetSortOrder()), Remark: f.GetRemark(),
		}
		m.ID = uint(f.GetId())
		m.CreatedAt = time.UnixMilli(f.GetCreatedAt())
		m.UpdatedAt = time.UnixMilli(f.GetUpdatedAt())
		list = append(list, m)
	}
	return list, int(rep.GetCount()), rep.GetCapped(), nil
}

// RemoveFriendLinks 批量删除友链。
func RemoveFriendLinks(ids []uint) error {
	return call(func(ctx context.Context, c communityv1.CommunityServiceClient) error {
		req := &communityv1.IDs{}
		for _, id := range ids {
			req.Ids = append(req.Ids, uint64(id))
		}
		_, err := c.RemoveFriendLinks(ctx, req)
		return err
	})
}

// UpdateFriendLink 更新友链。
func UpdateFriendLink(id uint, name, url, logo string, isShow bool, sortOrder int, remark string) error {
	return call(func(ctx context.Context, c communityv1.CommunityServiceClient) error {
		_, err := c.UpdateFriendLink(ctx, &communityv1.UpdateFriendLinkRequest{
			Id: uint64(id), Name: name, Url: url, Logo: logo, IsShow: isShow,
			SortOrder: int32(sortOrder), Remark: remark,
		})
		return err
	})
}

// ---- 友情推广 ----

// CreateFriendPromotion 新增友情推广。
func CreateFriendPromotion(title, friendName, avatar, category, description, previewImages string,
	contactInfo []string, isShow bool, sortOrder int, position, remark string) error {
	return call(func(ctx context.Context, c communityv1.CommunityServiceClient) error {
		_, err := c.CreateFriendPromotion(ctx, &communityv1.FriendPromotionWrite{
			Title: title, FriendName: friendName, Avatar: avatar, Category: category,
			Description: description, PreviewImages: previewImages, ContactInfo: contactInfo,
			IsShow: isShow, SortOrder: int32(sortOrder), Position: position, Remark: remark,
		})
		return err
	})
}

// ListFriendPromotions 友情推广列表;all=true 时包含隐藏项(仅管理员)。
func ListFriendPromotions(all bool, page, limit int, order string, endID uint) ([]models.FriendPromotion, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), communityTimeout)
	defer cancel()
	rep, err := c.ListFriendPromotions(ctx, &communityv1.ListRequest{
		All: all, Page: int32(page), Limit: int32(limit), Order: order, EndId: uint64(endID),
	})
	if err != nil {
		return nil, 0, false, err
	}
	list := make([]models.FriendPromotion, 0, len(rep.GetList()))
	for _, f := range rep.GetList() {
		m := models.FriendPromotion{
			Title: f.GetTitle(), FriendName: f.GetFriendName(), Avatar: f.GetAvatar(),
			Category: f.GetCategory(), Description: f.GetDescription(), PreviewImages: f.GetPreviewImages(),
			ContactInfo: f.GetContactInfo(), IsShow: f.GetIsShow(), SortOrder: int(f.GetSortOrder()),
			Position: f.GetPosition(), Remark: f.GetRemark(),
		}
		m.ID = uint(f.GetId())
		m.CreatedAt = time.UnixMilli(f.GetCreatedAt())
		m.UpdatedAt = time.UnixMilli(f.GetUpdatedAt())
		list = append(list, m)
	}
	return list, int(rep.GetCount()), rep.GetCapped(), nil
}

// RemoveFriendPromotions 批量删除友情推广。
func RemoveFriendPromotions(ids []uint) error {
	return call(func(ctx context.Context, c communityv1.CommunityServiceClient) error {
		req := &communityv1.IDs{}
		for _, id := range ids {
			req.Ids = append(req.Ids, uint64(id))
		}
		_, err := c.RemoveFriendPromotions(ctx, req)
		return err
	})
}

// UpdateFriendPromotion 更新友情推广。
func UpdateFriendPromotion(id uint, title, friendName, avatar, category, description, previewImages string,
	contactInfo []string, isShow bool, sortOrder int, position, remark string) error {
	return call(func(ctx context.Context, c communityv1.CommunityServiceClient) error {
		_, err := c.UpdateFriendPromotion(ctx, &communityv1.UpdateFriendPromotionRequest{
			Id: uint64(id), Title: title, FriendName: friendName, Avatar: avatar, Category: category,
			Description: description, PreviewImages: previewImages, ContactInfo: contactInfo,
			IsShow: isShow, SortOrder: int32(sortOrder), Position: position, Remark: remark,
		})
		return err
	})
}
