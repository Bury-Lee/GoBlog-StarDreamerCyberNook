package user_service

import (
	"context"
	"errors"
	"time"

	userv1 "StarDreamerCyberNook/gen/user/v1"
	"StarDreamerCyberNook/models/enum"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrEmailUsed 领域错误:邮箱已被使用。
var ErrEmailUsed = errors.New("user: email used")

func mapAccountErr(err error) error {
	if err == nil {
		return nil
	}
	if status.Code(err) == codes.AlreadyExists {
		return ErrEmailUsed
	}
	return err
}

// EmailExists 邮箱是否已存在。
func EmailExists(email string) (bool, error) {
	c, err := client()
	if err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), userTimeout)
	defer cancel()
	rep, err := c.EmailExists(ctx, &userv1.EmailExistsRequest{Email: email})
	if err != nil {
		return false, err
	}
	return rep.GetExists(), nil
}

// CreateUserByEmail 邮箱注册建号,返回用户ID/用户名/角色。
func CreateUserByEmail(email, nickName, userName, passwordHash string, role enum.RoleType) (uint, string, enum.RoleType, error) {
	c, err := client()
	if err != nil {
		return 0, "", 0, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), userTimeout)
	defer cancel()
	rep, err := c.CreateUserByEmail(ctx, &userv1.CreateUserByEmailRequest{
		Email: email, NickName: nickName, UserName: userName, PasswordHash: passwordHash, Role: int32(role),
	})
	if err != nil {
		return 0, "", 0, mapAccountErr(err)
	}
	return uint(rep.GetId()), rep.GetUserName(), enum.RoleType(rep.GetRole()), nil
}

// UpdateUserEmail 重置绑定邮箱。
func UpdateUserEmail(userID uint, email string) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), userTimeout)
	defer cancel()
	_, err = c.UpdateUserEmail(ctx, &userv1.UpdateUserEmailRequest{UserId: uint64(userID), Email: email})
	return err
}

// UpdatePassword 更新密码哈希。
func UpdatePassword(userID uint, passwordHash string) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), userTimeout)
	defer cancel()
	_, err = c.UpdatePassword(ctx, &userv1.UpdatePasswordRequest{UserId: uint64(userID), PasswordHash: passwordHash})
	return err
}

// LoginLogItem 登录日志条目。
type LoginLogItem struct {
	ID        uint
	UserID    uint
	IP        string
	Addr      string
	UserAgent string
	CreatedAt time.Time
	Nickname  string
	Avatar    string
}

// ListLoginLogs 登录日志列表。
func ListLoginLogs(userID uint, ip, addr, startTime, endTime string, page, limit int, order string, endID uint, withUser bool) ([]LoginLogItem, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), userTimeout)
	defer cancel()
	rep, err := c.ListLoginLogs(ctx, &userv1.ListLoginLogsRequest{
		UserId: uint64(userID), Ip: ip, Addr: addr, StartTime: startTime, EndTime: endTime,
		Page: int32(page), Limit: int32(limit), Order: order, EndId: uint64(endID), WithUser: withUser,
	})
	if err != nil {
		return nil, 0, false, err
	}
	out := make([]LoginLogItem, 0, len(rep.GetList()))
	for _, l := range rep.GetList() {
		out = append(out, LoginLogItem{
			ID: uint(l.GetId()), UserID: uint(l.GetUserId()), IP: l.GetIp(), Addr: l.GetAddr(),
			UserAgent: l.GetUserAgent(), CreatedAt: time.UnixMilli(l.GetCreatedAt()),
			Nickname: l.GetNickname(), Avatar: l.GetAvatar(),
		})
	}
	return out, int(rep.GetCount()), rep.GetCapped(), nil
}
