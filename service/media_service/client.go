// Package media_service 是经 gRPC 调用独立 media 服务的客户端(二进制存取)。
package media_service

import (
	"bytes"
	"context"
	"io"
	"sync"
	"time"

	mediav1 "StarDreamerCyberNook/gen/media/v1"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/pkg/svc"
)

const mediaTimeout = 30 * time.Second

var (
	cliOnce sync.Once
	cli     mediav1.MediaServiceClient
	cliErr  error
)

// client 懒加载到 media 服务的 gRPC 连接(服务键 "media")。
func client() (mediav1.MediaServiceClient, error) {
	cliOnce.Do(func() {
		var cfgs map[string][]string
		if global.Config != nil {
			cfgs = global.Config.Services
		}
		svc.Init(cfgs)
		conn, err := grpcx.Dial("media")
		if err != nil {
			cliErr = err
			return
		}
		cli = mediav1.NewMediaServiceClient(conn)
	})
	return cli, cliErr
}

// Put 存储对象。
func Put(ctx context.Context, key string, data []byte, contentType string) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, mediaTimeout)
	defer cancel()
	_, err = c.Put(ctx, &mediav1.PutRequest{Key: key, Data: data, ContentType: contentType})
	return err
}

// GetBytes 读取对象全部字节。
func GetBytes(ctx context.Context, key string) ([]byte, error) {
	c, err := client()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, mediaTimeout)
	defer cancel()
	stream, err := c.Get(ctx, &mediav1.GetRequest{Key: key})
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		buf.Write(chunk.GetData())
	}
	return buf.Bytes(), nil
}

// Remove 删除对象。
func Remove(ctx context.Context, key string) error {
	c, err := client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, mediaTimeout)
	defer cancel()
	_, err = c.Delete(ctx, &mediav1.DeleteRequest{Key: key})
	return err
}
