package internal

import (
	"context"
	"strings"
	"time"

	"StarDreamerCyberNook/common"
	userv1 "StarDreamerCyberNook/gen/user/v1"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// EmailExists 邮箱是否已存在。
func (s *Server) EmailExists(_ context.Context, req *userv1.EmailExistsRequest) (*userv1.EmailExistsReply, error) {
	var count int64
	if err := s.db.Model(&models.UserModel{}).Where("email = ?", strings.ToLower(req.GetEmail())).Count(&count).Error; err != nil {
		return nil, err
	}
	return &userv1.EmailExistsReply{Exists: count > 0}, nil
}

// CreateUserByEmail 邮箱注册建号(已存在返回 AlreadyExists)。
func (s *Server) CreateUserByEmail(_ context.Context, req *userv1.CreateUserByEmailRequest) (*userv1.CreateUserByEmailReply, error) {
	email := strings.ToLower(req.GetEmail())
	var exist models.UserModel
	if err := s.db.Take(&exist, "email = ?", email).Error; err == nil {
		return nil, status.Error(codes.AlreadyExists, "email used")
	}
	u := models.UserModel{
		NickName:       req.GetNickName(),
		UserName:       req.GetUserName(),
		Email:          email,
		Password:       req.GetPasswordHash(),
		Role:           enum.RoleType(req.GetRole()),
		RegisterSource: enum.RegisterEmail,
		LastLoginTime:  time.Now(),
	}
	if err := s.db.Create(&u).Error; err != nil {
		return nil, err
	}
	return &userv1.CreateUserByEmailReply{Id: uint64(u.ID), UserName: u.UserName, Role: int32(u.Role)}, nil
}

// UpdateUserEmail 重置绑定邮箱。
func (s *Server) UpdateUserEmail(_ context.Context, req *userv1.UpdateUserEmailRequest) (*userv1.Empty, error) {
	if err := s.db.Model(&models.UserModel{}).Where("id = ?", uint(req.GetUserId())).
		Update("email", strings.ToLower(req.GetEmail())).Error; err != nil {
		return nil, err
	}
	return &userv1.Empty{}, nil
}

// UpdatePassword 更新密码哈希。
func (s *Server) UpdatePassword(_ context.Context, req *userv1.UpdatePasswordRequest) (*userv1.Empty, error) {
	if err := s.db.Model(&models.UserModel{}).Where("id = ?", uint(req.GetUserId())).
		Update("password", req.GetPasswordHash()).Error; err != nil {
		return nil, err
	}
	return &userv1.Empty{}, nil
}

// ListLoginLogs 登录日志列表。
func (s *Server) ListLoginLogs(_ context.Context, req *userv1.ListLoginLogsRequest) (*userv1.ListLoginLogsReply, error) {
	where := s.db.Where("")
	if req.GetStartTime() != "" {
		where = where.Where("created_at >= ?", req.GetStartTime())
	}
	if req.GetEndTime() != "" {
		where = where.Where("created_at <= ?", req.GetEndTime())
	}
	var preloads []string
	if req.GetWithUser() {
		preloads = []string{"UserModel"}
	}
	opts := common.Options{
		PageInfo:      common.PageInfo{Page: int(req.GetPage()), Limit: int(req.GetLimit()), Order: req.GetOrder(), EndId: uint(req.GetEndId())},
		Where:         where,
		Preloads:      preloads,
		AllowedOrders: []string{"id", "created_at"},
		CountCap:      common.DefaultCountCap,
	}
	list, count, capped, err := common.ListQuery[models.UserLoginModel](s.db, models.UserLoginModel{
		UserID: uint(req.GetUserId()), IP: req.GetIp(), Addr: req.GetAddr(),
	}, opts)
	if err != nil {
		return nil, err
	}
	rep := &userv1.ListLoginLogsReply{Count: int64(count), Capped: capped}
	for _, m := range list {
		rep.List = append(rep.List, &userv1.LoginLog{
			Id: uint64(m.ID), UserId: uint64(m.UserID), Ip: m.IP, Addr: m.Addr, UserAgent: m.UserAgent,
			CreatedAt: m.CreatedAt.UnixMilli(), Nickname: m.UserModel.NickName, Avatar: m.UserModel.Avatar,
		})
	}
	return rep, nil
}
