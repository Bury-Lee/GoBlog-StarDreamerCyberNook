// media 服务:把二进制存储(本地目录 / S3 对象存储)从博客剥离为独立 gRPC 微服务。
//
// 运行:
//
//	go run ./services/media/cmd -config services/media/media.yaml
package main

import (
	"flag"
	"log"
	"net"
	"os"

	mediav1 "StarDreamerCyberNook/gen/media/v1"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/services/media/internal"

	"gopkg.in/yaml.v2"
)

type fileConfig struct {
	Listen   string `yaml:"listen"`
	Backend  string `yaml:"backend"`
	LocalDir string `yaml:"localDir"`
	Minio    struct {
		Host      string `yaml:"host"`
		AccessKey string `yaml:"accessKey"`
		SecretKey string `yaml:"secretKey"`
		Bucket    string `yaml:"bucket"`
		Region    string `yaml:"region"`
	} `yaml:"minio"`
}

func main() {
	cfgPath := flag.String("config", "services/media/media.yaml", "配置文件路径")
	flag.Parse()

	data, err := os.ReadFile(*cfgPath)
	if err != nil {
		log.Fatalf("media: read config %s: %v", *cfgPath, err)
	}
	var fc fileConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		log.Fatalf("media: parse config %s: %v", *cfgPath, err)
	}

	listen := overrideStr("MEDIA_LISTEN", fc.Listen)
	if listen == "" {
		listen = "127.0.0.1:9240"
	}
	backend := overrideStr("MEDIA_BACKEND", fc.Backend)
	if backend == "" {
		backend = "local"
	}

	srvImpl, err := internal.NewServer(internal.Config{
		Backend:   backend,
		LocalDir:  overrideStr("MEDIA_LOCAL_DIR", fc.LocalDir),
		Host:      overrideStr("MEDIA_MINIO_HOST", fc.Minio.Host),
		AccessKey: overrideStr("MEDIA_MINIO_ACCESS_KEY", fc.Minio.AccessKey),
		SecretKey: overrideStr("MEDIA_MINIO_SECRET_KEY", fc.Minio.SecretKey),
		Bucket:    overrideStr("MEDIA_MINIO_BUCKET", fc.Minio.Bucket),
		Region:    overrideStr("MEDIA_MINIO_REGION", fc.Minio.Region),
	})
	if err != nil {
		log.Fatalf("media: init backend: %v", err)
	}

	lis, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatalf("media: listen %s: %v", listen, err)
	}

	s := grpcx.NewServer()
	mediav1.RegisterMediaServiceServer(s, srvImpl)
	log.Printf("media service listening on %s (backend=%s)", listen, backend)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("media: serve: %v", err)
	}
}

func overrideStr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
