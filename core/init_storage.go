package core

import (
	"StarDreamerCyberNook/global"
	"context"
	"log"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

func InitClient() *minio.Client {
	ctx := context.Background()

	host := strings.TrimSpace(global.Config.ObjectStorage.Host) // 移除首尾空格
	u, err := url.Parse(host)                                   // 解析URL
	if err != nil {
		log.Fatalf("Host 格式错误: %v", err)
	}

	endpoint := u.Host
	secure := strings.EqualFold(u.Scheme, "https") // 检查是否为 HTTPS 协议,如果是则会在下方启用TLS加密
	if endpoint == "" {
		endpoint = host
	}

	// 初始化 MinIO 客户端
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(global.Config.ObjectStorage.AccessKey, global.Config.ObjectStorage.SecretKey, ""),
		Secure: secure,
		Region: global.Config.ObjectStorage.Region,
	})
	if err != nil {
		log.Fatalf("初始化 MinIO 客户端失败: %v", err)
	}

	// 检查存储桶是否存在
	exists, err := client.BucketExists(ctx, global.Config.ObjectStorage.Bucket)
	if err != nil {
		log.Fatalf("调用失败: %v", err)
	}
	if !exists {
		logrus.Warnf("存储桶 %s 不存在, 已自动创建", global.Config.ObjectStorage.Bucket)
		// 创建存储桶
		err = client.MakeBucket(ctx, global.Config.ObjectStorage.Bucket, minio.MakeBucketOptions{
			Region: global.Config.ObjectStorage.Region,
		})
		if err != nil {
			log.Fatalf("创建存储桶失败: %v", err)
		}
	}
	return client
}
