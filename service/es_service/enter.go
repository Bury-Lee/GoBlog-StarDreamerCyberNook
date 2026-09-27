// Package es_service 已从"直接操作 ES"改为"经 gRPC 调用独立 search 服务"的客户端。
// 对外函数签名保持不变,调用点无需改动;ES 被剥离到 services/search。
package es_service

import (
	"context"
	"sync"
	"time"

	searchv1 "StarDreamerCyberNook/gen/search/v1"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/pkg/svc"

	"github.com/sirupsen/logrus"
)

var (
	cliOnce sync.Once
	cli     searchv1.SearchServiceClient
	cliErr  error
)

// client 懒加载到 search 服务的 gRPC 连接(服务键 "search")。
func client() (searchv1.SearchServiceClient, error) {
	cliOnce.Do(func() {
		var cfgs map[string][]string
		if global.Config != nil {
			cfgs = global.Config.Services
		}
		svc.Init(cfgs)
		conn, err := grpcx.Dial("search")
		if err != nil {
			cliErr = err
			return
		}
		cli = searchv1.NewSearchServiceClient(conn)
	})
	return cli, cliErr
}

// Update 索引存在则删除再创建。
func Update(index, mapping string) {
	if _, err := EnsureIndex(index, mapping, true); err != nil {
		logrus.Errorf("ES更新索引失败:%s - %s", index, err)
		return
	}
	logrus.Infof("ES索引已更新:%s", index)
}

// CreateIndex 创建索引。
func CreateIndex(index, mapping string) {
	if _, err := EnsureIndex(index, mapping, false); err != nil {
		logrus.Errorf("ES创建索引失败:%s - %s", index, err)
	}
}

// EnsureIndex 确保索引存在;recreate=true 先删后建。
func EnsureIndex(index, mapping string, recreate bool) (bool, error) {
	c, err := client()
	if err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := c.EnsureIndex(ctx, &searchv1.EnsureIndexRequest{Index: index, Mapping: mapping, Recreate: recreate}); err != nil {
		return false, err
	}
	return true, nil
}

// ExistsIndex 判断索引是否存在。
func ExistsIndex(index string) bool {
	c, err := client()
	if err != nil {
		logrus.Errorf("检查索引失败:%s - %s", index, err)
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rep, err := c.ExistsIndex(ctx, &searchv1.ExistsIndexRequest{Index: index})
	if err != nil {
		logrus.Errorf("检查索引 %s 存在性时出错: %s", index, err)
		return false
	}
	return rep.GetExists()
}

// DeleteIndex 删除索引。
func DeleteIndex(index string) {
	c, err := client()
	if err != nil {
		logrus.Errorf("删除索引失败:%s - %s", index, err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := c.DeleteIndex(ctx, &searchv1.DeleteIndexRequest{Index: index}); err != nil {
		logrus.Errorf("删除索引失败:%s - %s", index, err)
		return
	}
	logrus.Infof("索引 %s 删除成功", index)
}
