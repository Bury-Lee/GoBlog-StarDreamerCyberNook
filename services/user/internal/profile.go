package internal

import (
	"context"
	"encoding/json"
	"time"

	"StarDreamerCyberNook/common"
	userv1 "StarDreamerCyberNook/gen/user/v1"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/utils/ip"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetUserByLogin 按 user_name / email 查用户(供网关做密码校验)。
func (s *Server) GetUserByLogin(_ context.Context, req *userv1.LoginLookupRequest) (*userv1.LoginLookupReply, error) {
	col := ""
	switch req.GetKind() {
	case "user_name":
		col = "user_name"
	case "email":
		col = "email"
	default:
		return nil, status.Error(codes.InvalidArgument, "bad kind")
	}
	var u models.UserModel
	if err := s.db.Take(&u, col+" = ?", req.GetVal()).Error; err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	return &userv1.LoginLookupReply{
		Id: uint64(u.ID), UserName: u.UserName, Role: int32(u.Role), Password: u.Password,
	}, nil
}

// RecordLogin 记录登录日志并更新最近登录时间/IP。
func (s *Server) RecordLogin(_ context.Context, req *userv1.RecordLoginRequest) (*userv1.Empty, error) {
	userID := uint(req.GetUserId())
	if err := s.db.Create(&models.UserLoginModel{
		UserID: userID, IP: req.GetIp(), UserAgent: req.GetUserAgent(),
	}).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&models.UserModel{}).Where("id = ?", userID).
		Updates(map[string]any{"last_login_time": time.Now(), "ip": req.GetAddr()}).Error; err != nil {
		return nil, err
	}
	return &userv1.Empty{}, nil
}

// GetUserDetail 个人主页详情(用户表 + 配置表)。
func (s *Server) GetUserDetail(_ context.Context, req *userv1.IDRequest) (*userv1.UserDetailReply, error) {
	var u models.UserModel
	if err := s.db.Preload("UserConfModel").Take(&u, uint(req.GetId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}
	rep := &userv1.UserDetailReply{
		Id: uint64(u.ID), CreatedAt: u.CreatedAt.UnixMilli(), UpdatedAt: u.UpdatedAt.UnixMilli(),
		UserName: u.UserName, NickName: u.NickName, Avatar: u.Avatar, Abstract: u.Abstract,
		Age: int32(u.Age), LikeTags: u.LikeTags, ContactInfo: u.ContactInfo, Role: int32(u.Role),
		Email: u.Email,
	}
	if u.UserConfModel != nil {
		if u.UserConfModel.UpdateUsernameDate != nil {
			v := u.UserConfModel.UpdateUsernameDate.UnixMilli()
			rep.UpdateUsernameDate = &v
		}
		rep.OpenFollow = u.UserConfModel.OpenFollow
		rep.OpenFans = u.UserConfModel.OpenFans
		rep.HomeStyleId = uint64(u.UserConfModel.HomeStyleID)
	}
	return rep, nil
}

// GetUserBaseInfo 用户基本信息(含文章/粉丝/关注计数)。
func (s *Server) GetUserBaseInfo(_ context.Context, req *userv1.IDRequest) (*userv1.UserBaseInfoReply, error) {
	var u models.UserModel
	if err := s.db.Take(&u, uint(req.GetId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}
	var articleCount, fansCount, followCount int64
	s.db.Model(&models.ArticleModel{}).Where("user_id = ?", u.ID).Count(&articleCount)
	s.db.Model(&models.UserFollowModel{}).Where("focus_user_id = ?", u.ID).Count(&fansCount)
	s.db.Model(&models.UserFollowModel{}).Where("user_id = ?", u.ID).Count(&followCount)

	var lastLogin int64
	if !u.LastLoginTime.IsZero() {
		lastLogin = u.LastLoginTime.UnixMilli()
	}
	return &userv1.UserBaseInfoReply{
		UserId: uint64(u.ID), Age: int32(u.Age), NickName: u.NickName, Avatar: u.Avatar,
		LastLoginTime: lastLogin, Region: ip.GetIpAddr(u.IP), ExistDay: int32(u.ExistDays()),
		ArticleCount: articleCount, FansCount: fansCount, FollowCount: followCount,
	}, nil
}

// ListUsers 用户列表(仅公开字段)。
func (s *Server) ListUsers(_ context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersReply, error) {
	opts := common.Options{
		PageInfo:      common.PageInfo{Page: int(req.GetPage()), Limit: int(req.GetLimit()), Order: req.GetOrder(), EndId: uint(req.GetEndId())},
		Likes:         []string{"nick_name"},
		DefaultOrder:  "id desc",
		AllowedOrders: []string{"id", "created_at", "last_login_time", "age"},
		CountCap:      common.DefaultCountCap,
	}
	if req.GetUserId() != 0 {
		opts.Where = s.db.Where("id = ?", uint(req.GetUserId()))
	}
	list, count, capped, err := common.ListQuery[models.UserModel](s.db, models.UserModel{}, opts)
	if err != nil {
		return nil, err
	}
	rep := &userv1.ListUsersReply{Count: int64(count), Capped: capped}
	for _, m := range list {
		rep.List = append(rep.List, &userv1.UserListItem{
			Id: uint64(m.ID), NickName: m.NickName, Avatar: m.Avatar, Abstract: m.Abstract,
		})
	}
	return rep, nil
}

// UpdateUser 增量更新用户信息(用户表 / 配置表补丁)。
func (s *Server) UpdateUser(_ context.Context, req *userv1.UpdateUserRequest) (*userv1.Empty, error) {
	userID := uint(req.GetUserId())

	if req.GetUserPatch() != "" {
		var patch map[string]any
		if err := json.Unmarshal([]byte(req.GetUserPatch()), &patch); err != nil {
			return nil, status.Error(codes.InvalidArgument, "bad user patch")
		}
		if len(patch) > 0 {
			var u models.UserModel
			if err := s.db.Preload("UserConfModel").Take(&u, userID).Error; err != nil {
				return nil, status.Error(codes.NotFound, "用户不存在")
			}
			if err := s.db.Model(&u).Updates(patch).Error; err != nil {
				return nil, err
			}
		}
	}

	if req.GetConfPatch() != "" {
		var patch map[string]any
		if err := json.Unmarshal([]byte(req.GetConfPatch()), &patch); err != nil {
			return nil, status.Error(codes.InvalidArgument, "bad conf patch")
		}
		if len(patch) > 0 {
			var conf models.UserConfModel
			if err := s.db.Take(&conf, "user_id = ?", userID).Error; err != nil {
				return nil, status.Error(codes.NotFound, "用户配置信息不存在")
			}
			if err := s.db.Model(&conf).Updates(patch).Error; err != nil {
				return nil, err
			}
		}
	}
	return &userv1.Empty{}, nil
}
