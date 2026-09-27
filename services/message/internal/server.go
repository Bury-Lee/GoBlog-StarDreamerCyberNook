package internal

import (
	"context"
	"fmt"

	"StarDreamerCyberNook/common"
	messagev1 "StarDreamerCyberNook/gen/message/v1"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/pkg/dbx"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// Config 是 message 服务的配置。
type Config struct {
	Driver     string
	DSN        string
	Standalone bool // true: 自建独立库连接池;false: 复用宿主 global.DB
}

// Server 实现 MessageService。
type Server struct {
	messagev1.UnimplementedMessageServiceServer
	db *gorm.DB
}

// NewServer 获取数据库句柄:未启用独立库时复用宿主 global.DB,启用时自建连接池。
func NewServer(cfg Config) (*Server, error) {
	db := global.DB
	if cfg.Standalone {
		d, err := dbx.Open(cfg.Driver, cfg.DSN)
		if err != nil {
			return nil, fmt.Errorf("message: open db: %w", err)
		}
		db = d
	}
	if db == nil {
		return nil, fmt.Errorf("message: 无可用数据库(共享模式需宿主先初始化 global.DB;独立模式请开启 dbStandalone)")
	}
	return &Server{db: db}, nil
}

func toProtoMessage(m models.MessageModel) *messagev1.Message {
	return &messagev1.Message{
		Id:                 uint64(m.ID),
		Type:               int32(m.Type),
		RevUserId:          uint64(m.RevUserID),
		ActionUserId:       uint64(m.ActionUserID),
		ActionUserNickname: m.ActionUserNickname,
		ActionUserAvatar:   m.ActionUserAvatar,
		Title:              m.Title,
		Content:            m.Content,
		ArticleId:          uint64(m.ArticleID),
		ArticleTitle:       m.ArticleTitle,
		CommentId:          uint64(m.CommentID),
		LinkTitle:          m.LinkTitle,
		LinkHref:           m.LinkHref,
		IsRead:             m.IsRead,
		CreatedAt:          m.CreatedAt.UnixMilli(),
		UpdatedAt:          m.UpdatedAt.UnixMilli(),
	}
}

// List 消息列表并标记已读。
func (s *Server) List(_ context.Context, req *messagev1.ListRequest) (*messagev1.ListReply, error) {
	uid := uint(req.GetUserId())
	query := s.db.Where("rev_user_id = ?", uid)
	switch models.MessageType(req.GetType()) {
	case models.MessageTypeAt, models.MessageTypeComment, models.MessageTypeReply:
		query = query.Where("type in ?", []models.MessageType{models.MessageTypeAt, models.MessageTypeComment, models.MessageTypeReply}).
			Where("action_user_id <> ?", uid)
	case models.MessageTypeCollect, models.MessageTypeDigg:
		query = query.Where("type in ?", []models.MessageType{models.MessageTypeCollect, models.MessageTypeDigg})
	case models.MessageTypePrivate:
		query = query.Where("type = ?", models.MessageTypePrivate)
	case models.MessageTypeSystem:
		query = query.Where("type = ?", models.MessageTypeSystem)
	default:
		return nil, status.Error(codes.InvalidArgument, "消息类型错误")
	}

	opts := common.Options{
		PageInfo:      common.PageInfo{Page: int(req.GetPage()), Limit: int(req.GetLimit()), Order: req.GetOrder(), EndId: uint(req.GetEndId()), Key: req.GetKey()},
		Where:         query,
		AllowedOrders: []string{"id", "created_at"},
		CountCap:      common.DefaultCountCap,
	}
	list, count, capped, err := common.ListQuery(s.db, models.MessageModel{RevUserID: uid}, opts)
	if err != nil {
		return nil, err
	}
	rep := &messagev1.ListReply{Count: int64(count), Capped: capped}
	var unread []uint
	for _, m := range list {
		rep.List = append(rep.List, toProtoMessage(m))
		if !m.IsRead {
			unread = append(unread, m.ID)
		}
	}
	if len(unread) > 0 {
		_ = s.db.Model(&models.MessageModel{}).Where("id IN ?", unread).Update("is_read", true).Error
	}
	return rep, nil
}

// UnreadCounts 各类型未读数。
func (s *Server) UnreadCounts(_ context.Context, req *messagev1.UnreadRequest) (*messagev1.UnreadReply, error) {
	uid := uint(req.GetUserId())
	type row struct {
		Type  models.MessageType
		Count int64
	}
	var rows []row
	if err := s.db.Model(&models.MessageModel{}).
		Select("type, count(*) as count").
		Where("rev_user_id = ? AND is_read = ? AND action_user_id <> ?", uid, false, uid).
		Group("type").Find(&rows).Error; err != nil {
		return nil, err
	}
	rep := &messagev1.UnreadReply{Counts: map[int32]int64{}}
	for _, r := range rows {
		rep.Counts[int32(r.Type)] = r.Count
	}
	return rep, nil
}

// Remove 删除消息。
func (s *Server) Remove(_ context.Context, req *messagev1.BatchRequest) (*messagev1.Empty, error) {
	ids := make([]uint, 0, len(req.GetIds()))
	for _, id := range req.GetIds() {
		ids = append(ids, uint(id))
	}
	if len(ids) == 0 {
		return &messagev1.Empty{}, nil
	}
	if err := s.db.Where("rev_user_id = ?", uint(req.GetUserId())).
		Where("id in ?", ids).Delete(&models.MessageModel{}).Error; err != nil {
		return nil, err
	}
	return &messagev1.Empty{}, nil
}

// OneKeyRead 一键已读。
func (s *Server) OneKeyRead(_ context.Context, req *messagev1.OneKeyReadRequest) (*messagev1.Empty, error) {
	var targetTypes []models.MessageType
	if req.GetCommentMessage() {
		targetTypes = append(targetTypes, models.MessageTypeAt, models.MessageTypeComment, models.MessageTypeReply)
	}
	if req.GetDiggAndCollectMessage() {
		targetTypes = append(targetTypes, models.MessageTypeDigg, models.MessageTypeCollect)
	}
	if req.GetPrivateMessage() {
		targetTypes = append(targetTypes, models.MessageTypePrivate)
	}
	if req.GetSystemMessage() {
		targetTypes = append(targetTypes, models.MessageTypeSystem)
	}
	if len(targetTypes) == 0 {
		return &messagev1.Empty{}, nil
	}
	if err := s.db.Model(&models.MessageModel{}).
		Where("rev_user_id = ?", uint(req.GetUserId())).
		Where("type IN ?", targetTypes).
		Update("is_read", true).Error; err != nil {
		return nil, err
	}
	return &messagev1.Empty{}, nil
}

// GetConf 用户消息配置。
func (s *Server) GetConf(_ context.Context, req *messagev1.ConfUserRequest) (*messagev1.ConfReply, error) {
	var conf models.UserMessageConfModel
	if err := s.db.Where("user_id = ?", uint(req.GetUserId())).Take(&conf).Error; err != nil {
		return nil, status.Error(codes.NotFound, "查询用户消息配置失败")
	}
	return &messagev1.ConfReply{
		UserId:             uint64(conf.UserID),
		OpenCommentMessage: conf.OpenCommentMessage,
		OpenReplyMessage:   conf.OpenReplyMessage,
		OpenDiggMessage:    conf.OpenDiggMessage,
		OpenCollectMessage: conf.OpenCollectMessage,
		OpenPrivateMessage: conf.OpenPrivateMessage,
	}, nil
}

// UpdateConf 更新用户消息配置。
func (s *Server) UpdateConf(_ context.Context, req *messagev1.UpdateConfRequest) (*messagev1.Empty, error) {
	mps := map[string]any{}
	if req.OpenCommentMessage != nil {
		mps["open_comment_message"] = req.GetOpenCommentMessage()
	}
	if req.OpenReplyMessage != nil {
		mps["open_reply_message"] = req.GetOpenReplyMessage()
	}
	if req.OpenDiggMessage != nil {
		mps["open_digg_message"] = req.GetOpenDiggMessage()
	}
	if req.OpenCollectMessage != nil {
		mps["open_collect_message"] = req.GetOpenCollectMessage()
	}
	if req.OpenPrivateMessage != nil {
		mps["open_private_message"] = req.GetOpenPrivateMessage()
	}
	if len(mps) == 0 {
		return &messagev1.Empty{}, nil
	}
	if err := s.db.Model(&models.UserMessageConfModel{}).
		Where("user_id = ?", uint(req.GetUserId())).Updates(mps).Error; err != nil {
		return nil, err
	}
	return &messagev1.Empty{}, nil
}

func (s *Server) userBrief(id uint) (models.UserModel, error) {
	var u models.UserModel
	err := s.db.Select("id", "nick_name", "avatar").Take(&u, id).Error
	return u, err
}

func (s *Server) insertComment(req *messagev1.InsertCommentRequest, t models.MessageType) error {
	action, err := s.userBrief(uint(req.GetActionUserId()))
	if err != nil {
		return err
	}
	var article models.ArticleModel
	if err := s.db.Select("id", "title").Take(&article, uint(req.GetArticleId())).Error; err != nil {
		return err
	}
	return s.db.Create(&models.MessageModel{
		Type:               t,
		RevUserID:          uint(req.GetRevUserId()),
		ActionUserID:       uint(req.GetActionUserId()),
		ActionUserNickname: action.NickName,
		ActionUserAvatar:   action.Avatar,
		Title:              t.String(),
		ArticleID:          uint(req.GetArticleId()),
		ArticleTitle:       article.Title,
		CommentID:          uint(req.GetCommentId()),
		Content:            req.GetContent(),
		IsRead:             false,
	}).Error
}

// InsertComment 评论通知。
func (s *Server) InsertComment(_ context.Context, req *messagev1.InsertCommentRequest) (*messagev1.Empty, error) {
	return &messagev1.Empty{}, s.insertComment(req, models.MessageTypeComment)
}

// InsertReply 回复通知。
func (s *Server) InsertReply(_ context.Context, req *messagev1.InsertCommentRequest) (*messagev1.Empty, error) {
	return &messagev1.Empty{}, s.insertComment(req, models.MessageTypeReply)
}

// InsertSystem 系统通知。
func (s *Server) InsertSystem(_ context.Context, req *messagev1.SystemMessage) (*messagev1.Empty, error) {
	err := s.db.Create(&models.MessageModel{
		Type:               models.MessageTypeSystem,
		RevUserID:          uint(req.GetRevUserId()),
		ActionUserID:       uint(req.GetActionUserId()),
		ActionUserNickname: req.GetActionUserNickname(),
		ActionUserAvatar:   req.GetActionUserAvatar(),
		Title:              req.GetTitle(),
		ArticleID:          uint(req.GetArticleId()),
		ArticleTitle:       req.GetArticleTitle(),
		CommentID:          uint(req.GetCommentId()),
		Content:            req.GetContent(),
		LinkTitle:          req.GetLinkTitle(),
		LinkHref:           req.GetLinkHref(),
		IsRead:             false,
	}).Error
	return &messagev1.Empty{}, err
}

// InsertAt @通知。
func (s *Server) InsertAt(_ context.Context, req *messagev1.InsertAtRequest) (*messagev1.Empty, error) {
	err := s.db.Create(&models.MessageModel{
		Type:               models.MessageTypeAt,
		RevUserID:          uint(req.GetRevUserId()),
		ActionUserID:       uint(req.GetActionUserId()),
		ActionUserNickname: req.GetActionUserNickname(),
		ActionUserAvatar:   req.GetActionUserAvatar(),
		Title:              models.MessageTypeAt.String(),
		IsRead:             false,
	}).Error
	return &messagev1.Empty{}, err
}

// InsertArticleDigg 文章点赞通知。
func (s *Server) InsertArticleDigg(_ context.Context, req *messagev1.InsertDiggRequest) (*messagev1.Empty, error) {
	var article models.ArticleModel
	if err := s.db.Select("id", "title", "user_id").Take(&article, uint(req.GetArticleId())).Error; err != nil {
		return nil, err
	}
	action, err := s.userBrief(uint(req.GetActionUserId()))
	if err != nil {
		return nil, err
	}
	err = s.db.Create(&models.MessageModel{
		Type:               models.MessageTypeDigg,
		RevUserID:          article.UserID,
		ActionUserID:       action.ID,
		ActionUserNickname: action.NickName,
		ActionUserAvatar:   action.Avatar,
		Title:              models.MessageTypeDigg.String(),
		ArticleID:          article.ID,
		ArticleTitle:       article.Title,
		IsRead:             false,
	}).Error
	return &messagev1.Empty{}, err
}

// InsertCollect 文章收藏通知。
func (s *Server) InsertCollect(_ context.Context, req *messagev1.InsertCollectRequest) (*messagev1.Empty, error) {
	var article models.ArticleModel
	if err := s.db.Select("id", "title", "user_id").Take(&article, uint(req.GetArticleId())).Error; err != nil {
		return nil, err
	}
	action, err := s.userBrief(uint(req.GetActionUserId()))
	if err != nil {
		return nil, err
	}
	err = s.db.Create(&models.MessageModel{
		Type:               models.MessageTypeCollect,
		RevUserID:          article.UserID,
		ActionUserID:       action.ID,
		ActionUserNickname: action.NickName,
		ActionUserAvatar:   action.Avatar,
		Title:              models.MessageTypeCollect.String(),
		ArticleID:          article.ID,
		ArticleTitle:       article.Title,
		IsRead:             false,
	}).Error
	return &messagev1.Empty{}, err
}
