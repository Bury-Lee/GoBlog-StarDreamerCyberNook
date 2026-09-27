package content_service

import (
	"context"
	"errors"

	contentv1 "StarDreamerCyberNook/gen/content/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 领域错误:供网关映射为友好提示。
var (
	ErrNotFound         = errors.New("content: not found")
	ErrAlreadyTop       = errors.New("content: already top")
	ErrTopLimit         = errors.New("content: top limit")
	ErrPermissionDenied = errors.New("content: permission denied")
	ErrFolderNotFound   = errors.New("content: collect folder not found")
)

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	switch status.Code(err) {
	case codes.NotFound:
		return ErrNotFound
	case codes.AlreadyExists:
		return ErrAlreadyTop
	case codes.ResourceExhausted:
		return ErrTopLimit
	case codes.PermissionDenied:
		return ErrPermissionDenied
	case codes.FailedPrecondition:
		return ErrFolderNotFound
	default:
		return err
	}
}

// ToggleArticleDigg 切换文章点赞;返回是否已点赞与状态是否真正改变。
func ToggleArticleDigg(userID, articleID uint) (digged, changed bool, err error) {
	c, err := client()
	if err != nil {
		return false, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.ToggleArticleDigg(ctx, &contentv1.ToggleArticleDiggRequest{
		UserId: uint64(userID), ArticleId: uint64(articleID),
	})
	if err != nil {
		return false, false, mapErr(err)
	}
	return rep.GetDigged(), rep.GetChanged(), nil
}

// GetArticleInteraction 查询用户对文章的点赞/收藏状态。
func GetArticleInteraction(userID, articleID uint) (digged, collected bool, err error) {
	c, err := client()
	if err != nil {
		return false, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	rep, err := c.GetArticleInteraction(ctx, &contentv1.GetArticleInteractionRequest{
		UserId: uint64(userID), ArticleId: uint64(articleID),
	})
	if err != nil {
		return false, false, err
	}
	return rep.GetDigged(), rep.GetCollected(), nil
}

// TopArticle 置顶文章;非管理员受 maxTop 限制。
func TopArticle(userID, articleID uint, isAdmin bool, maxTop int) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	_, err = c.TopArticle(ctx, &contentv1.TopArticleRequest{
		UserId: uint64(userID), ArticleId: uint64(articleID), IsAdmin: isAdmin, MaxTop: int32(maxTop),
	})
	return mapErr(err)
}

// CancelArticleTop 取消置顶。
func CancelArticleTop(userID, articleID uint) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	_, err = c.CancelArticleTop(ctx, &contentv1.CancelArticleTopRequest{
		UserId: uint64(userID), ArticleId: uint64(articleID),
	})
	return mapErr(err)
}

// AdminCancelArticleTop 管理员强制取消指定用户的置顶。
func AdminCancelArticleTop(userID, articleID uint) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), contentTimeout)
	defer cancel()
	_, err = c.AdminCancelArticleTop(ctx, &contentv1.CancelArticleTopRequest{
		UserId: uint64(userID), ArticleId: uint64(articleID),
	})
	return mapErr(err)
}
