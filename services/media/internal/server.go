package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	mediav1 "StarDreamerCyberNook/gen/media/v1"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/pkg/dbx"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/gorm"
)

// Config 是 media 服务的存储与数据库配置。
type Config struct {
	Backend   string // local | minio
	LocalDir  string // 本地后端根目录
	Host      string // MinIO/S3 地址
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string

	Driver     string // 图片元数据库:mysql | postgres | sqlite
	DSN        string
	Standalone bool // true: 自建独立库连接池;false: 复用宿主 global.DB
}

// Server 实现 MediaService。
type Server struct {
	mediav1.UnimplementedMediaServiceServer
	cfg   Config
	minio *minio.Client
	db    *gorm.DB
}

// NewServer 按配置初始化后端与数据库句柄。
func NewServer(cfg Config) (*Server, error) {
	s := &Server{cfg: cfg}
	if cfg.Backend == "minio" {
		endpoint := strings.TrimPrefix(strings.TrimPrefix(cfg.Host, "https://"), "http://")
		secure := strings.HasPrefix(cfg.Host, "https://")
		cli, err := minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
			Secure: secure,
			Region: cfg.Region,
		})
		if err != nil {
			return nil, err
		}
		s.minio = cli
	}
	if cfg.Standalone {
		db, err := dbx.Open(cfg.Driver, cfg.DSN)
		if err != nil {
			return nil, fmt.Errorf("media: open db: %w", err)
		}
		s.db = db
	} else {
		s.db = global.DB
	}
	return s, nil
}

func (s *Server) localPath(key string) string {
	return filepath.Join(s.cfg.LocalDir, filepath.FromSlash(key))
}

// Put 存储对象。
func (s *Server) Put(ctx context.Context, req *mediav1.PutRequest) (*mediav1.PutReply, error) {
	key := req.GetKey()
	if key == "" {
		return nil, fmt.Errorf("media: empty key")
	}
	if s.cfg.Backend == "minio" {
		_, err := s.minio.PutObject(ctx, s.cfg.Bucket, key,
			bytes.NewReader(req.GetData()), int64(len(req.GetData())),
			minio.PutObjectOptions{ContentType: req.GetContentType()})
		if err != nil {
			return nil, err
		}
		return &mediav1.PutReply{Key: key}, nil
	}
	// local
	p := s.localPath(key)
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		return nil, err
	}
	if err := os.WriteFile(p, req.GetData(), 0o640); err != nil {
		return nil, err
	}
	return &mediav1.PutReply{Key: key}, nil
}

// Get 读取对象(流式分块)。
func (s *Server) Get(req *mediav1.GetRequest, stream mediav1.MediaService_GetServer) error {
	key := req.GetKey()
	var r io.ReadCloser
	if s.cfg.Backend == "minio" {
		obj, err := s.minio.GetObject(stream.Context(), s.cfg.Bucket, key, minio.GetObjectOptions{})
		if err != nil {
			return err
		}
		r = obj
	} else {
		f, err := os.Open(s.localPath(key))
		if err != nil {
			return err
		}
		r = f
	}
	defer r.Close()

	buf := make([]byte, 64*1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			if serr := stream.Send(&mediav1.GetChunk{Data: append([]byte(nil), buf[:n]...)}); serr != nil {
				return serr
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// Delete 删除对象。
func (s *Server) Delete(ctx context.Context, req *mediav1.DeleteRequest) (*mediav1.DeleteReply, error) {
	key := req.GetKey()
	if s.cfg.Backend == "minio" {
		if err := s.minio.RemoveObject(ctx, s.cfg.Bucket, key, minio.RemoveObjectOptions{}); err != nil {
			return nil, err
		}
		return &mediav1.DeleteReply{}, nil
	}
	if err := os.Remove(s.localPath(key)); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return &mediav1.DeleteReply{}, nil
}
