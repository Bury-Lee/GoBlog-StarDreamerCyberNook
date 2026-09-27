package internal

import (
	"context"
	"fmt"

	"StarDreamerCyberNook/common"
	logv1 "StarDreamerCyberNook/gen/log/v1"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/pkg/dbx"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// Config 是 log 服务的数据库配置。
type Config struct {
	Driver     string
	DSN        string
	Standalone bool // true: 自建独立库连接池;false: 复用宿主 global.DB
}

// Server 实现 LogService。
type Server struct {
	logv1.UnimplementedLogServiceServer
	db *gorm.DB
}

// NewServer 获取数据库句柄:未启用独立库时复用宿主 global.DB,启用时自建连接池。
func NewServer(cfg Config) (*Server, error) {
	db := global.DB
	if cfg.Standalone {
		d, err := dbx.Open(cfg.Driver, cfg.DSN)
		if err != nil {
			return nil, fmt.Errorf("log: open db: %w", err)
		}
		db = d
	}
	if db == nil {
		return nil, fmt.Errorf("log: 无可用数据库(共享模式需宿主先初始化 global.DB;独立模式请开启 dbStandalone)")
	}
	return &Server{db: db}, nil
}

func toProtoLog(m models.LogModel) *logv1.Log {
	return &logv1.Log{
		Id: uint64(m.ID), LogType: int32(m.LogType), Title: m.Title, Content: m.Content, Level: int32(m.Level),
		UserId: uint64(m.UserID), Ip: m.IP, Addr: m.Addr, IsRead: m.IsRead, LoginStatus: m.LoginStatus,
		LoginType: int32(m.LoginType), ServiceName: m.ServiceName,
		CreatedAt: m.CreatedAt.UnixMilli(), UpdatedAt: m.UpdatedAt.UnixMilli(),
		Nickname: m.UserModel.NickName, Avatar: m.UserModel.Avatar,
	}
}

// ListLogs 日志列表。
func (s *Server) ListLogs(_ context.Context, req *logv1.ListLogsRequest) (*logv1.LogListReply, error) {
	var where *gorm.DB
	if req.LoginStatus != nil {
		where = s.db.Where("login_status = ?", req.GetLoginStatus())
	}
	var preloads []string
	if req.GetWithUser() {
		preloads = []string{"UserModel"}
	}
	list, count, capped, err := common.ListQuery[models.LogModel](s.db, models.LogModel{
		UserID: uint(req.GetUserId()), LogType: enum.LogType(req.GetLogType()), Level: enum.LogLevel(req.GetLevel()),
		IP: req.GetIp(), ServiceName: req.GetServiceName(),
	}, common.Options{
		PageInfo:      common.PageInfo{Page: int(req.GetPage()), Limit: int(req.GetLimit()), Order: req.GetOrder(), EndId: uint(req.GetEndId())},
		Likes:         []string{"title"},
		Where:         where,
		Preloads:      preloads,
		AllowedOrders: []string{"id", "created_at"},
		CountCap:      common.DefaultCountCap,
	})
	if err != nil {
		return nil, err
	}
	rep := &logv1.LogListReply{Count: int64(count), Capped: capped}
	for _, m := range list {
		rep.List = append(rep.List, toProtoLog(m))
	}
	return rep, nil
}

// ReadLog 标记已读。
func (s *Server) ReadLog(_ context.Context, req *logv1.ReadLogRequest) (*logv1.Empty, error) {
	var m models.LogModel
	if err := s.db.Take(&m, uint(req.GetId())).Error; err != nil {
		return nil, status.Error(codes.NotFound, "日志不存在")
	}
	if !m.IsRead {
		s.db.Model(&m).Update("is_read", true)
	}
	return &logv1.Empty{}, nil
}

// RemoveLogs 删除日志。
func (s *Server) RemoveLogs(_ context.Context, req *logv1.RemoveLogsRequest) (*logv1.RemoveLogsReply, error) {
	ids := make([]uint, 0, len(req.GetIds()))
	for _, id := range req.GetIds() {
		ids = append(ids, uint(id))
	}
	rep := &logv1.RemoveLogsReply{}
	if len(ids) == 0 {
		return rep, nil
	}
	var list []models.LogModel
	if err := s.db.Find(&list, "id IN ?", ids).Error; err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return rep, nil
	}
	if err := s.db.Delete(&list).Error; err != nil {
		return nil, err
	}
	rep.Deleted = int32(len(list))
	return rep, nil
}
