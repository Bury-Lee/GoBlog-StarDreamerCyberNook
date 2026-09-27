package log_service

import (
	"context"
	"sync"
	"time"

	logv1 "StarDreamerCyberNook/gen/log/v1"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/pkg/svc"
)

const logTimeout = 10 * time.Second

var (
	logOnce sync.Once
	logCli  logv1.LogServiceClient
	logErr  error
)

func logClient() (logv1.LogServiceClient, error) {
	logOnce.Do(func() {
		var cfgs map[string][]string
		if global.Config != nil {
			cfgs = global.Config.Services
		}
		svc.Init(cfgs)
		conn, err := grpcx.Dial("log")
		if err != nil {
			logErr = err
			return
		}
		logCli = logv1.NewLogServiceClient(conn)
	})
	return logCli, logErr
}

func fromProtoLog(p *logv1.Log) models.LogModel {
	m := models.LogModel{
		LogType: enum.LogType(p.GetLogType()), Title: p.GetTitle(), Content: p.GetContent(),
		Level: enum.LogLevel(p.GetLevel()), UserID: uint(p.GetUserId()), IP: p.GetIp(), Addr: p.GetAddr(),
		IsRead: p.GetIsRead(), LoginStatus: p.GetLoginStatus(), LoginType: enum.LoginType(p.GetLoginType()),
		ServiceName: p.GetServiceName(),
	}
	m.ID = uint(p.GetId())
	m.CreatedAt = time.UnixMilli(p.GetCreatedAt())
	m.UpdatedAt = time.UnixMilli(p.GetUpdatedAt())
	m.UserModel = models.UserModel{NickName: p.GetNickname(), Avatar: p.GetAvatar()}
	return m
}

// ListLogs 日志列表。
func ListLogs(logType enum.LogType, level enum.LogLevel, ip, serviceName string, loginStatus *bool,
	userID uint, key string, page, limit int, order string, endID uint, withUser bool) ([]models.LogModel, int, bool, error) {
	c, err := logClient()
	if err != nil {
		return nil, 0, false, err
	}
	req := &logv1.ListLogsRequest{
		LogType: int32(logType), Level: int32(level), Ip: ip, ServiceName: serviceName,
		UserId: uint64(userID), Key: key, Page: int32(page), Limit: int32(limit), Order: order,
		EndId: uint64(endID), WithUser: withUser,
	}
	if loginStatus != nil {
		v := *loginStatus
		req.LoginStatus = &v
	}
	ctx, cancel := context.WithTimeout(context.Background(), logTimeout)
	defer cancel()
	rep, err := c.ListLogs(ctx, req)
	if err != nil {
		return nil, 0, false, err
	}
	out := make([]models.LogModel, 0, len(rep.GetList()))
	for _, p := range rep.GetList() {
		out = append(out, fromProtoLog(p))
	}
	return out, int(rep.GetCount()), rep.GetCapped(), nil
}

// ReadLog 标记日志已读。
func ReadLog(id uint) error {
	c, err := logClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), logTimeout)
	defer cancel()
	_, err = c.ReadLog(ctx, &logv1.ReadLogRequest{Id: uint64(id)})
	return err
}

// RemoveLogs 删除日志,返回删除条数。
func RemoveLogs(ids []uint) (int, error) {
	c, err := logClient()
	if err != nil {
		return 0, err
	}
	req := &logv1.RemoveLogsRequest{}
	for _, id := range ids {
		req.Ids = append(req.Ids, uint64(id))
	}
	ctx, cancel := context.WithTimeout(context.Background(), logTimeout)
	defer cancel()
	rep, err := c.RemoveLogs(ctx, req)
	if err != nil {
		return 0, err
	}
	return int(rep.GetDeleted()), nil
}
