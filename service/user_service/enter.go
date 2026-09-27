// Package user_service 是经 gRPC 调用独立 user 服务的客户端(用户关系域:关注/粉丝/好友)。
package user_service

import (
	"context"
	"sync"
	"time"

	userv1 "StarDreamerCyberNook/gen/user/v1"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/pkg/svc"
)

const userTimeout = 10 * time.Second

var (
	cliOnce sync.Once
	cli     userv1.UserServiceClient
	cliErr  error
)

func client() (userv1.UserServiceClient, error) {
	cliOnce.Do(func() {
		var cfgs map[string][]string
		if global.Config != nil {
			cfgs = global.Config.Services
		}
		svc.Init(cfgs)
		conn, err := grpcx.Dial("user")
		if err != nil {
			cliErr = err
			return
		}
		cli = userv1.NewUserServiceClient(conn)
	})
	return cli, cliErr
}

// FollowItem 关注条目展示信息。
type FollowItem struct {
	FocusUserID uint
	Nickname    string
	Avatar      string
	Abstract    string
	CreatedAt   time.Time
}

// FollowRecord 原始关注记录(粉丝列表)。
type FollowRecord struct {
	ID          uint
	UserID      uint
	FocusUserID uint
	Friend      bool
	CreatedAt   time.Time
}

// Follow 关注用户;already 表示此前已关注。
func Follow(userID, focusUserID uint) (bool, error) {
	c, err := client()
	if err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), userTimeout)
	defer cancel()
	rep, err := c.Follow(ctx, &userv1.FollowRequest{UserId: uint64(userID), FocusUserId: uint64(focusUserID)})
	if err != nil {
		return false, err
	}
	return rep.GetAlready(), nil
}

// Unfollow 取关用户。
func Unfollow(userID, focusUserID uint) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), userTimeout)
	defer cancel()
	_, err = c.Unfollow(ctx, &userv1.FollowRequest{UserId: uint64(userID), FocusUserId: uint64(focusUserID)})
	return err
}

// FollowCheck 是否已关注。
func FollowCheck(userID, focusUserID uint) (bool, error) {
	c, err := client()
	if err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), userTimeout)
	defer cancel()
	rep, err := c.FollowCheck(ctx, &userv1.CheckRequest{UserId: uint64(userID), FocusUserId: uint64(focusUserID)})
	if err != nil {
		return false, err
	}
	return rep.GetFollowed(), nil
}

func listReq(userID, viewerID uint, logged bool, page, limit int, order string, endID uint) *userv1.ListRequest {
	return &userv1.ListRequest{
		UserId: uint64(userID), ViewerId: uint64(viewerID), ViewerLogged: logged,
		Page: int32(page), Limit: int32(limit), Order: order, EndId: uint64(endID),
	}
}

func itemsOf(rep *userv1.FollowItemListReply) []FollowItem {
	out := make([]FollowItem, 0, len(rep.GetList()))
	for _, it := range rep.GetList() {
		out = append(out, FollowItem{
			FocusUserID: uint(it.GetFocusUserId()),
			Nickname:    it.GetNickname(),
			Avatar:      it.GetAvatar(),
			Abstract:    it.GetAbstract(),
			CreatedAt:   time.UnixMilli(it.GetCreatedAt()),
		})
	}
	return out
}

// FollowingList 关注列表。
func FollowingList(userID, viewerID uint, logged bool, page, limit int, order string, endID uint) ([]FollowItem, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), userTimeout)
	defer cancel()
	rep, err := c.FollowingList(ctx, listReq(userID, viewerID, logged, page, limit, order, endID))
	if err != nil {
		return nil, 0, false, err
	}
	return itemsOf(rep), int(rep.GetCount()), rep.GetCapped(), nil
}

// FollowerList 粉丝列表。
func FollowerList(userID, viewerID uint, logged bool, page, limit int, order string, endID uint) ([]FollowRecord, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), userTimeout)
	defer cancel()
	rep, err := c.FollowerList(ctx, listReq(userID, viewerID, logged, page, limit, order, endID))
	if err != nil {
		return nil, 0, false, err
	}
	out := make([]FollowRecord, 0, len(rep.GetList()))
	for _, r := range rep.GetList() {
		out = append(out, FollowRecord{
			ID: uint(r.GetId()), UserID: uint(r.GetUserId()), FocusUserID: uint(r.GetFocusUserId()),
			Friend: r.GetFriend(), CreatedAt: time.UnixMilli(r.GetCreatedAt()),
		})
	}
	return out, int(rep.GetCount()), rep.GetCapped(), nil
}

// FriendList 好友列表。
func FriendList(userID, viewerID uint, logged bool, page, limit int, order string, endID uint) ([]FollowItem, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), userTimeout)
	defer cancel()
	rep, err := c.FriendList(ctx, listReq(userID, viewerID, logged, page, limit, order, endID))
	if err != nil {
		return nil, 0, false, err
	}
	return itemsOf(rep), int(rep.GetCount()), rep.GetCapped(), nil
}
