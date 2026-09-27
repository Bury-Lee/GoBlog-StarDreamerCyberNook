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

// ErrNotFound 领域错误:网关据此映射为友好提示。
var ErrNotFound = errors.New("user: not found")

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if status.Code(err) == codes.NotFound {
		return ErrNotFound
	}
	return err
}

// LoginUser 登录查询结果。
type LoginUser struct {
	ID       uint
	UserName string
	Role     enum.RoleType
	Password string
}

// GetUserByLogin 按 user_name / email 查用户。
func GetUserByLogin(kind, val string) (LoginUser, error) {
	c, err := client()
	if err != nil {
		return LoginUser{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), userTimeout)
	defer cancel()
	rep, err := c.GetUserByLogin(ctx, &userv1.LoginLookupRequest{Kind: kind, Val: val})
	if err != nil {
		return LoginUser{}, mapErr(err)
	}
	return LoginUser{
		ID: uint(rep.GetId()), UserName: rep.GetUserName(),
		Role: enum.RoleType(rep.GetRole()), Password: rep.GetPassword(),
	}, nil
}

// RecordLogin 记录登录日志并更新最近登录。
func RecordLogin(userID uint, ip, addr, userAgent string) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), userTimeout)
	defer cancel()
	_, err = c.RecordLogin(ctx, &userv1.RecordLoginRequest{
		UserId: uint64(userID), Ip: ip, Addr: addr, UserAgent: userAgent,
	})
	return err
}

// UserDetail 个人主页详情。
type UserDetail struct {
	ID                 uint
	CreatedAt          time.Time
	UpdatedAt          time.Time
	UserName           string
	NickName           string
	Avatar             string
	Abstract           string
	Age                int
	LikeTags           []string
	ContactInfo        map[string]string
	Role               enum.RoleType
	UpdateUsernameDate *time.Time
	OpenFollow         bool
	OpenFans           bool
	HomeStyleID        uint
	Email              string
}

// GetUserDetail 个人主页详情。
func GetUserDetail(id uint) (UserDetail, error) {
	c, err := client()
	if err != nil {
		return UserDetail{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), userTimeout)
	defer cancel()
	rep, err := c.GetUserDetail(ctx, &userv1.IDRequest{Id: uint64(id)})
	if err != nil {
		return UserDetail{}, mapErr(err)
	}
	d := UserDetail{
		ID: uint(rep.GetId()), CreatedAt: time.UnixMilli(rep.GetCreatedAt()), UpdatedAt: time.UnixMilli(rep.GetUpdatedAt()),
		UserName: rep.GetUserName(), NickName: rep.GetNickName(), Avatar: rep.GetAvatar(), Abstract: rep.GetAbstract(),
		Age: int(rep.GetAge()), LikeTags: rep.GetLikeTags(), ContactInfo: rep.GetContactInfo(),
		Role: enum.RoleType(rep.GetRole()), OpenFollow: rep.GetOpenFollow(),
		OpenFans: rep.GetOpenFans(), HomeStyleID: uint(rep.GetHomeStyleId()), Email: rep.GetEmail(),
	}
	if rep.UpdateUsernameDate != nil {
		t := time.UnixMilli(rep.GetUpdateUsernameDate())
		d.UpdateUsernameDate = &t
	}
	return d, nil
}

// UserBaseInfo 用户基本信息。
type UserBaseInfo struct {
	UserID        uint
	Age           int
	NickName      string
	Avatar        string
	LastLoginTime time.Time
	Region        string
	ExistDay      int
	ArticleCount  int64
	FansCount     int64
	FollowCount   int64
}

// GetUserBaseInfo 用户基本信息(含计数)。
func GetUserBaseInfo(id uint) (UserBaseInfo, error) {
	c, err := client()
	if err != nil {
		return UserBaseInfo{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), userTimeout)
	defer cancel()
	rep, err := c.GetUserBaseInfo(ctx, &userv1.IDRequest{Id: uint64(id)})
	if err != nil {
		return UserBaseInfo{}, mapErr(err)
	}
	b := UserBaseInfo{
		UserID: uint(rep.GetUserId()), Age: int(rep.GetAge()), NickName: rep.GetNickName(), Avatar: rep.GetAvatar(),
		Region: rep.GetRegion(), ExistDay: int(rep.GetExistDay()),
		ArticleCount: rep.GetArticleCount(), FansCount: rep.GetFansCount(), FollowCount: rep.GetFollowCount(),
	}
	if rep.GetLastLoginTime() != 0 {
		b.LastLoginTime = time.UnixMilli(rep.GetLastLoginTime())
	}
	return b, nil
}

// UserItem 用户列表条目(仅公开字段)。
type UserItem struct {
	ID       uint
	NickName string
	Avatar   string
	Abstract string
}

// ListUsers 用户列表。
func ListUsers(userID uint, page, limit int, order, key string, endID uint) ([]UserItem, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), userTimeout)
	defer cancel()
	rep, err := c.ListUsers(ctx, &userv1.ListUsersRequest{
		UserId: uint64(userID), Page: int32(page), Limit: int32(limit), Order: order, Key: key, EndId: uint64(endID),
	})
	if err != nil {
		return nil, 0, false, err
	}
	out := make([]UserItem, 0, len(rep.GetList()))
	for _, it := range rep.GetList() {
		out = append(out, UserItem{ID: uint(it.GetId()), NickName: it.GetNickName(), Avatar: it.GetAvatar(), Abstract: it.GetAbstract()})
	}
	return out, int(rep.GetCount()), rep.GetCapped(), nil
}

// UpdateUser 增量更新用户信息(补丁为 JSON 字符串,列名 -> 值)。
func UpdateUser(userID uint, userPatch, confPatch string) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), userTimeout)
	defer cancel()
	_, err = c.UpdateUser(ctx, &userv1.UpdateUserRequest{
		UserId: uint64(userID), UserPatch: userPatch, ConfPatch: confPatch,
	})
	return err
}
