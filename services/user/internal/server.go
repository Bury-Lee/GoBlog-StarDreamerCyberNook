package internal

import (
	"context"
	"fmt"

	"StarDreamerCyberNook/common"
	userv1 "StarDreamerCyberNook/gen/user/v1"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/pkg/dbx"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// Config 是 user 服务的配置。
type Config struct {
	Driver     string
	DSN        string
	Standalone bool // true: 自建独立库连接池;false: 复用宿主 global.DB
}

// Server 实现 UserService。
type Server struct {
	userv1.UnimplementedUserServiceServer
	db *gorm.DB
}

// NewServer 获取数据库句柄:未启用独立库时复用宿主 global.DB,启用时自建连接池。
func NewServer(cfg Config) (*Server, error) {
	db := global.DB
	if cfg.Standalone {
		d, err := dbx.Open(cfg.Driver, cfg.DSN)
		if err != nil {
			return nil, fmt.Errorf("user: open db: %w", err)
		}
		db = d
	}
	if db == nil {
		return nil, fmt.Errorf("user: 无可用数据库(共享模式需宿主先初始化 global.DB;独立模式请开启 dbStandalone)")
	}
	return &Server{db: db}, nil
}

// Follow 关注用户。
func (s *Server) Follow(_ context.Context, req *userv1.FollowRequest) (*userv1.FollowReply, error) {
	if req.GetFocusUserId() == req.GetUserId() {
		return nil, status.Error(codes.InvalidArgument, "其实你时刻都在关注自己~")
	}
	var user models.UserModel
	if err := s.db.Take(&user, uint(req.GetFocusUserId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "关注用户不存在")
	}
	var focus models.UserFollowModel
	if err := s.db.Take(&focus, "user_id = ? and focus_user_id = ?", uint(req.GetUserId()), user.ID).Error; err == nil {
		return &userv1.FollowReply{Already: true}, nil
	}
	if err := s.db.Create(&models.UserFollowModel{
		UserID:      uint(req.GetUserId()),
		FocusUserID: uint(req.GetFocusUserId()),
	}).Error; err != nil {
		return nil, err
	}
	return &userv1.FollowReply{}, nil
}

// Unfollow 取关用户。
func (s *Server) Unfollow(_ context.Context, req *userv1.FollowRequest) (*userv1.FollowReply, error) {
	if req.GetFocusUserId() == req.GetUserId() {
		return nil, status.Error(codes.InvalidArgument, "你无法取关自己")
	}
	var user models.UserModel
	if err := s.db.Take(&user, uint(req.GetFocusUserId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "取关用户不存在")
	}
	var focus models.UserFollowModel
	if err := s.db.Take(&focus, "user_id = ? and focus_user_id = ?", uint(req.GetUserId()), user.ID).Error; err != nil {
		return nil, status.Error(codes.NotFound, "未关注此用户")
	}
	if err := s.db.Delete(&focus).Error; err != nil {
		return nil, err
	}
	return &userv1.FollowReply{}, nil
}

// FollowCheck 是否已关注。
func (s *Server) FollowCheck(_ context.Context, req *userv1.CheckRequest) (*userv1.CheckReply, error) {
	var count int64
	if err := s.db.Model(&models.UserFollowModel{}).
		Where("user_id = ? and focus_user_id = ?", uint(req.GetUserId()), uint(req.GetFocusUserId())).
		Count(&count).Error; err != nil {
		return nil, err
	}
	return &userv1.CheckReply{Followed: count > 0}, nil
}

func listOptions(page, limit int, order string, endID uint, defaultOrder string, where *gorm.DB) common.Options {
	return common.Options{
		PageInfo:      common.PageInfo{Page: page, Limit: limit, Order: order, EndId: endID},
		Preloads:      []string{"FocusUserModel"},
		Where:         where,
		DefaultOrder:  defaultOrder,
		AllowedOrders: []string{"id", "created_at"},
		CountCap:      common.DefaultCountCap,
	}
}

// FollowingList 关注列表(遵循 OpenFollow 隐私)。
func (s *Server) FollowingList(_ context.Context, req *userv1.ListRequest) (*userv1.FollowItemListReply, error) {
	var conf models.UserConfModel
	if err := s.db.Take(&conf, "user_id = ?", uint(req.GetUserId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "用户配置信息不存在")
	}
	if !conf.OpenFollow && !(req.GetViewerLogged() && req.GetViewerId() == req.GetUserId()) {
		return nil, status.Error(codes.PermissionDenied, "此用户未公开我的关注")
	}
	opts := listOptions(int(req.GetPage()), int(req.GetLimit()), req.GetOrder(), uint(req.GetEndId()), "", s.db.Where("user_id = ?", uint(req.GetUserId())))
	list, count, capped, err := common.ListQuery(s.db, models.UserFollowModel{}, opts)
	if err != nil {
		return nil, err
	}
	rep := &userv1.FollowItemListReply{Count: int64(count), Capped: capped}
	for _, m := range list {
		rep.List = append(rep.List, &userv1.FollowItem{
			FocusUserId: uint64(m.FocusUserID),
			Nickname:    m.FocusUserModel.NickName,
			Avatar:      m.FocusUserModel.Avatar,
			Abstract:    m.FocusUserModel.Abstract,
			CreatedAt:   m.CreatedAt.UnixMilli(),
		})
	}
	return rep, nil
}

// FollowerList 粉丝列表(遵循 OpenFans 隐私)。
func (s *Server) FollowerList(_ context.Context, req *userv1.ListRequest) (*userv1.FollowRecordListReply, error) {
	var conf models.UserConfModel
	_ = s.db.Where("user_id = ?", uint(req.GetUserId())).First(&conf).Error
	if !conf.OpenFans && !(req.GetViewerLogged() && req.GetViewerId() == req.GetUserId()) {
		return nil, status.Error(codes.PermissionDenied, "此用户未公开我的粉丝")
	}
	opts := common.Options{
		PageInfo:      common.PageInfo{Page: int(req.GetPage()), Limit: int(req.GetLimit()), Order: req.GetOrder(), EndId: uint(req.GetEndId())},
		Preloads:      []string{"UserModel"},
		DefaultOrder:  "created_at desc",
		AllowedOrders: []string{"id", "created_at"},
		CountCap:      common.DefaultCountCap,
	}
	list, count, capped, err := common.ListQuery(s.db, models.UserFollowModel{FocusUserID: uint(req.GetUserId())}, opts)
	if err != nil {
		return nil, err
	}
	rep := &userv1.FollowRecordListReply{Count: int64(count), Capped: capped}
	for _, m := range list {
		rep.List = append(rep.List, &userv1.FollowRecord{
			Id:           uint64(m.ID),
			UserId:       uint64(m.UserID),
			FocusUserId:  uint64(m.FocusUserID),
			Friend:       m.Friend,
			CreatedAt:    m.CreatedAt.UnixMilli(),
		})
	}
	return rep, nil
}

// FriendList 好友列表(仅本人)。
func (s *Server) FriendList(_ context.Context, req *userv1.ListRequest) (*userv1.FollowItemListReply, error) {
	opts := listOptions(int(req.GetPage()), int(req.GetLimit()), req.GetOrder(), uint(req.GetEndId()), "", s.db.Where("user_id = ? and friend = ?", uint(req.GetUserId()), true))
	list, count, capped, err := common.ListQuery(s.db, models.UserFollowModel{}, opts)
	if err != nil {
		return nil, err
	}
	rep := &userv1.FollowItemListReply{Count: int64(count), Capped: capped}
	for _, m := range list {
		rep.List = append(rep.List, &userv1.FollowItem{
			FocusUserId: uint64(m.FocusUserID),
			Nickname:    m.FocusUserModel.NickName,
			Avatar:      m.FocusUserModel.Avatar,
			Abstract:    m.FocusUserModel.Abstract,
			CreatedAt:   m.CreatedAt.UnixMilli(),
		})
	}
	return rep, nil
}
