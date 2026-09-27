package community_service

import (
	"context"
	"time"

	communityv1 "StarDreamerCyberNook/gen/community/v1"
	"StarDreamerCyberNook/models"
)

// CreateFeedback 提交反馈。
func CreateFeedback(userID uint, isAnonymous bool, content, contact string, typ models.FeedbackType) error {
	return call(func(ctx context.Context, c communityv1.CommunityServiceClient) error {
		_, err := c.CreateFeedback(ctx, &communityv1.CreateFeedbackRequest{
			UserId: uint64(userID), IsAnonymous: isAnonymous, Content: content,
			Contact: contact, Type: int32(typ),
		})
		return err
	})
}

// ListFeedbacks 反馈墙列表(状态/类型可选筛选)。
func ListFeedbacks(status *models.FeedbackStatus, typ *models.FeedbackType, page, limit int, order string, endID uint) ([]models.FeedbackModel, int, bool, error) {
	c, err := client()
	if err != nil {
		return nil, 0, false, err
	}
	req := &communityv1.ListFeedbacksRequest{
		Page: int32(page), Limit: int32(limit), Order: order, EndId: uint64(endID),
	}
	if status != nil {
		v := int32(*status)
		req.Status = &v
	}
	if typ != nil {
		v := int32(*typ)
		req.Type = &v
	}
	ctx, cancel := context.WithTimeout(context.Background(), communityTimeout)
	defer cancel()
	rep, err := c.ListFeedbacks(ctx, req)
	if err != nil {
		return nil, 0, false, err
	}
	list := make([]models.FeedbackModel, 0, len(rep.GetList()))
	for _, f := range rep.GetList() {
		list = append(list, fromProtoFeedback(f))
	}
	return list, int(rep.GetCount()), rep.GetCapped(), nil
}

// HandleFeedback 管理员处理反馈,返回更新后的条目供前端增量替换。
func HandleFeedback(id uint, status models.FeedbackStatus, reply string, handlerID uint) (models.FeedbackModel, error) {
	c, err := client()
	if err != nil {
		return models.FeedbackModel{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), communityTimeout)
	defer cancel()
	rep, err := c.HandleFeedback(ctx, &communityv1.HandleFeedbackRequest{
		Id: uint64(id), Status: int32(status), Reply: reply, HandlerId: uint64(handlerID),
	})
	if err != nil {
		return models.FeedbackModel{}, err
	}
	return fromProtoFeedback(rep.GetFeedback()), nil
}

func fromProtoFeedback(f *communityv1.Feedback) models.FeedbackModel {
	m := models.FeedbackModel{
		UserID: uint(f.GetUserId()), IsAnonymous: f.GetIsAnonymous(), Content: f.GetContent(),
		Contact: f.GetContact(), Type: models.FeedbackType(f.GetType()),
		Status: models.FeedbackStatus(f.GetStatus()), Reply: f.GetReply(), HandlerID: uint(f.GetHandlerId()),
	}
	m.ID = uint(f.GetId())
	m.CreatedAt = time.UnixMilli(f.GetCreatedAt())
	m.UpdatedAt = time.UnixMilli(f.GetUpdatedAt())
	return m
}
