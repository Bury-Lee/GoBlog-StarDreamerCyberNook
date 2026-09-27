package internal

import (
	"context"
	"fmt"

	"StarDreamerCyberNook/common"
	communityv1 "StarDreamerCyberNook/gen/community/v1"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/pkg/dbx"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// Config 是 community 服务的配置。
type Config struct {
	Driver     string
	DSN        string
	Standalone bool // true: 自建独立库连接池;false: 复用宿主 global.DB
}

// Server 实现 CommunityService。
type Server struct {
	communityv1.UnimplementedCommunityServiceServer
	db *gorm.DB
}

// NewServer 获取数据库句柄:未启用独立库时复用宿主 global.DB,启用时自建连接池。
func NewServer(cfg Config) (*Server, error) {
	db := global.DB
	if cfg.Standalone {
		d, err := dbx.Open(cfg.Driver, cfg.DSN)
		if err != nil {
			return nil, fmt.Errorf("community: open db: %w", err)
		}
		db = d
	}
	if db == nil {
		return nil, fmt.Errorf("community: 无可用数据库(共享模式需宿主先初始化 global.DB;独立模式请开启 dbStandalone)")
	}
	return &Server{db: db}, nil
}

// ---- feedback ----

func toProtoFeedback(f models.FeedbackModel) *communityv1.Feedback {
	return &communityv1.Feedback{
		Id: uint64(f.ID), UserId: uint64(f.UserID), IsAnonymous: f.IsAnonymous, Content: f.Content,
		Contact: f.Contact, Type: int32(f.Type), Status: int32(f.Status), Reply: f.Reply,
		HandlerId: uint64(f.HandlerID), CreatedAt: f.CreatedAt.UnixMilli(), UpdatedAt: f.UpdatedAt.UnixMilli(),
	}
}

func (s *Server) CreateFeedback(_ context.Context, req *communityv1.CreateFeedbackRequest) (*communityv1.Empty, error) {
	err := s.db.Create(&models.FeedbackModel{
		UserID: uint(req.GetUserId()), IsAnonymous: req.GetIsAnonymous(), Content: req.GetContent(),
		Contact: req.GetContact(), Type: models.FeedbackType(req.GetType()),
	}).Error
	return &communityv1.Empty{}, err
}

func (s *Server) ListFeedbacks(_ context.Context, req *communityv1.ListFeedbacksRequest) (*communityv1.FeedbackListReply, error) {
	var where *gorm.DB
	if req.Status != nil {
		where = s.db.Where("status = ?", *req.Status)
	}
	if req.Type != nil {
		if where != nil {
			where = where.Where("type = ?", *req.Type)
		} else {
			where = s.db.Where("type = ?", *req.Type)
		}
	}
	opts := common.Options{
		PageInfo:      common.PageInfo{Page: int(req.GetPage()), Limit: int(req.GetLimit()), Order: req.GetOrder(), EndId: uint(req.GetEndId())},
		Where:         where,
		DefaultOrder:  "created_at desc",
		AllowedOrders: []string{"id", "created_at", "status"},
		CountCap:      common.DefaultCountCap,
	}
	list, count, capped, err := common.ListQuery(s.db, models.FeedbackModel{}, opts)
	if err != nil {
		return nil, err
	}
	rep := &communityv1.FeedbackListReply{Count: int64(count), Capped: capped}
	for _, f := range list {
		f.Contact = ""
		f.HandlerID = 0
		if f.IsAnonymous {
			f.UserID = 0
		}
		rep.List = append(rep.List, toProtoFeedback(f))
	}
	return rep, nil
}

func (s *Server) HandleFeedback(_ context.Context, req *communityv1.HandleFeedbackRequest) (*communityv1.FeedbackReply, error) {
	var f models.FeedbackModel
	if err := s.db.Take(&f, uint(req.GetId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "反馈不存在")
	}
	if req.GetStatus() < int32(models.FeedbackStatusPending) || req.GetStatus() > int32(models.FeedbackStatusResolved) {
		return nil, status.Error(codes.InvalidArgument, "非法的状态")
	}
	if err := s.db.Model(&f).Updates(map[string]any{
		"status": req.GetStatus(), "reply": req.GetReply(), "handler_id": uint(req.GetHandlerId()),
	}).Error; err != nil {
		return nil, err
	}
	f.Status = models.FeedbackStatus(req.GetStatus())
	f.Reply = req.GetReply()
	f.Contact = ""
	f.HandlerID = 0
	if f.IsAnonymous {
		f.UserID = 0
	}
	return &communityv1.FeedbackReply{Feedback: toProtoFeedback(f)}, nil
}

// ---- banner ----

func toProtoBanner(b models.BannerModel) *communityv1.Banner {
	return &communityv1.Banner{
		Id: uint64(b.ID), IsShow: b.IsShow, Cover: b.Cover, Href: b.Href,
		CreatedAt: b.CreatedAt.UnixMilli(), UpdatedAt: b.UpdatedAt.UnixMilli(),
	}
}

func (s *Server) CreateBanner(_ context.Context, req *communityv1.BannerWrite) (*communityv1.Empty, error) {
	err := s.db.Create(&models.BannerModel{Cover: req.GetCover(), Href: req.GetHref(), IsShow: req.GetIsShow()}).Error
	return &communityv1.Empty{}, err
}

func (s *Server) ListBanners(_ context.Context, req *communityv1.ListRequest) (*communityv1.BannerListReply, error) {
	model := models.BannerModel{IsShow: true}
	if req.GetAll() {
		model = models.BannerModel{}
	}
	opts := common.Options{
		PageInfo:      common.PageInfo{Page: int(req.GetPage()), Limit: int(req.GetLimit()), Order: req.GetOrder(), EndId: uint(req.GetEndId())},
		AllowedOrders: []string{"id", "created_at"},
		CountCap:      common.DefaultCountCap,
	}
	list, count, capped, err := common.ListQuery(s.db, model, opts)
	if err != nil {
		return nil, err
	}
	rep := &communityv1.BannerListReply{Count: int64(count), Capped: capped}
	for _, b := range list {
		rep.List = append(rep.List, toProtoBanner(b))
	}
	return rep, nil
}

func (s *Server) RemoveBanners(_ context.Context, req *communityv1.IDs) (*communityv1.Empty, error) {
	ids := make([]uint, 0, len(req.GetIds()))
	for _, id := range req.GetIds() {
		ids = append(ids, uint(id))
	}
	if len(ids) > 0 {
		if err := s.db.Where("id in ?", ids).Delete(&models.BannerModel{}).Error; err != nil {
			return nil, err
		}
	}
	return &communityv1.Empty{}, nil
}

func (s *Server) UpdateBanner(_ context.Context, req *communityv1.UpdateBannerRequest) (*communityv1.Empty, error) {
	var m models.BannerModel
	if err := s.db.Take(&m, uint(req.GetId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "未找到记录")
	}
	err := s.db.Model(&m).Updates(map[string]any{"cover": req.GetCover(), "href": req.GetHref(), "is_show": req.GetIsShow()}).Error
	return &communityv1.Empty{}, err
}

// ---- friend link ----

func toProtoFriendLink(f models.FriendLink) *communityv1.FriendLink {
	return &communityv1.FriendLink{
		Id: uint64(f.ID), Name: f.Name, Url: f.URL, Logo: f.Logo, IsShow: f.IsShow,
		SortOrder: int32(f.SortOrder), Remark: f.Remark,
		CreatedAt: f.CreatedAt.UnixMilli(), UpdatedAt: f.UpdatedAt.UnixMilli(),
	}
}

func (s *Server) CreateFriendLink(_ context.Context, req *communityv1.FriendLinkWrite) (*communityv1.Empty, error) {
	err := s.db.Create(&models.FriendLink{
		Name: req.GetName(), URL: req.GetUrl(), Logo: req.GetLogo(),
		IsShow: req.GetIsShow(), SortOrder: int(req.GetSortOrder()), Remark: req.GetRemark(),
	}).Error
	return &communityv1.Empty{}, err
}

func (s *Server) ListFriendLinks(_ context.Context, req *communityv1.ListRequest) (*communityv1.FriendLinkListReply, error) {
	model := models.FriendLink{IsShow: true}
	if req.GetAll() {
		model = models.FriendLink{}
	}
	list, count, capped, err := common.ListQuery(s.db, model, common.Options{
		PageInfo:      common.PageInfo{Page: int(req.GetPage()), Limit: int(req.GetLimit()), Order: req.GetOrder(), EndId: uint(req.GetEndId())},
		AllowedOrders: []string{"id", "created_at", "sort_order"},
		CountCap:      common.DefaultCountCap,
	})
	if err != nil {
		return nil, err
	}
	rep := &communityv1.FriendLinkListReply{Count: int64(count), Capped: capped}
	for _, f := range list {
		rep.List = append(rep.List, toProtoFriendLink(f))
	}
	return rep, nil
}

func (s *Server) RemoveFriendLinks(_ context.Context, req *communityv1.IDs) (*communityv1.Empty, error) {
	ids := make([]uint, 0, len(req.GetIds()))
	for _, id := range req.GetIds() {
		ids = append(ids, uint(id))
	}
	if len(ids) > 0 {
		if err := s.db.Where("id IN ?", ids).Delete(&models.FriendLink{}).Error; err != nil {
			return nil, err
		}
	}
	return &communityv1.Empty{}, nil
}

func (s *Server) UpdateFriendLink(_ context.Context, req *communityv1.UpdateFriendLinkRequest) (*communityv1.Empty, error) {
	var m models.FriendLink
	if err := s.db.Take(&m, uint(req.GetId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "未找到记录")
	}
	err := s.db.Model(&m).Updates(map[string]any{
		"name": req.GetName(), "url": req.GetUrl(), "logo": req.GetLogo(),
		"is_show": req.GetIsShow(), "sort_order": req.GetSortOrder(), "remark": req.GetRemark(),
	}).Error
	return &communityv1.Empty{}, err
}

// ---- friend promotion ----

func toProtoFriendPromotion(f models.FriendPromotion) *communityv1.FriendPromotion {
	return &communityv1.FriendPromotion{
		Id: uint64(f.ID), Title: f.Title, FriendName: f.FriendName, Avatar: f.Avatar,
		Category: f.Category, Description: f.Description, PreviewImages: f.PreviewImages,
		ContactInfo: f.ContactInfo, IsShow: f.IsShow, SortOrder: int32(f.SortOrder),
		Position: f.Position, Remark: f.Remark,
		CreatedAt: f.CreatedAt.UnixMilli(), UpdatedAt: f.UpdatedAt.UnixMilli(),
	}
}

func (s *Server) CreateFriendPromotion(_ context.Context, req *communityv1.FriendPromotionWrite) (*communityv1.Empty, error) {
	err := s.db.Create(&models.FriendPromotion{
		Title: req.GetTitle(), FriendName: req.GetFriendName(), Avatar: req.GetAvatar(),
		Category: req.GetCategory(), Description: req.GetDescription(), PreviewImages: req.GetPreviewImages(),
		ContactInfo: req.GetContactInfo(), IsShow: req.GetIsShow(), SortOrder: int(req.GetSortOrder()),
		Position: req.GetPosition(), Remark: req.GetRemark(),
	}).Error
	return &communityv1.Empty{}, err
}

func (s *Server) ListFriendPromotions(_ context.Context, req *communityv1.ListRequest) (*communityv1.FriendPromotionListReply, error) {
	model := models.FriendPromotion{IsShow: true}
	if req.GetAll() {
		model = models.FriendPromotion{}
	}
	list, count, capped, err := common.ListQuery(s.db, model, common.Options{
		PageInfo:      common.PageInfo{Page: int(req.GetPage()), Limit: int(req.GetLimit()), Order: req.GetOrder(), EndId: uint(req.GetEndId())},
		AllowedOrders: []string{"id", "created_at", "sort_order"},
		CountCap:      common.DefaultCountCap,
	})
	if err != nil {
		return nil, err
	}
	rep := &communityv1.FriendPromotionListReply{Count: int64(count), Capped: capped}
	for _, f := range list {
		rep.List = append(rep.List, toProtoFriendPromotion(f))
	}
	return rep, nil
}

func (s *Server) RemoveFriendPromotions(_ context.Context, req *communityv1.IDs) (*communityv1.Empty, error) {
	ids := make([]uint, 0, len(req.GetIds()))
	for _, id := range req.GetIds() {
		ids = append(ids, uint(id))
	}
	if len(ids) > 0 {
		if err := s.db.Where("id IN ?", ids).Delete(&models.FriendPromotion{}).Error; err != nil {
			return nil, err
		}
	}
	return &communityv1.Empty{}, nil
}

func (s *Server) UpdateFriendPromotion(_ context.Context, req *communityv1.UpdateFriendPromotionRequest) (*communityv1.Empty, error) {
	var m models.FriendPromotion
	if err := s.db.Take(&m, uint(req.GetId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "未找到记录")
	}
	err := s.db.Model(&m).Updates(map[string]any{
		"title": req.GetTitle(), "friend_name": req.GetFriendName(), "avatar": req.GetAvatar(),
		"category": req.GetCategory(), "description": req.GetDescription(), "preview_images": req.GetPreviewImages(),
		"contact_info": req.GetContactInfo(), "is_show": req.GetIsShow(), "sort_order": req.GetSortOrder(),
		"position": req.GetPosition(), "remark": req.GetRemark(),
	}).Error
	return &communityv1.Empty{}, err
}
