package internal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	authv1 "StarDreamerCyberNook/gen/auth/v1"

	"github.com/golang-jwt/jwt/v5"
)

// Config 是 auth 服务的 JWT 配置(密钥集中管理)。
type Config struct {
	AccessSecret        string
	RefreshSecret       string
	AccessExpireMinutes int
	RefreshExpireHours  int
	Issuer              string
}

// Server 实现 AuthService。
type Server struct {
	authv1.UnimplementedAuthServiceServer
	cfg Config
}

// NewServer 创建服务。
func NewServer(cfg Config) *Server { return &Server{cfg: cfg} }

var signingMethod = jwt.SigningMethodHS256

func (s *Server) newJTI() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

type accessClaims struct {
	UserID   uint64 `json:"ID"`
	Username string `json:"name"`
	Role     int32  `json:"role"`
	jwt.RegisteredClaims
}

type refreshClaims struct {
	ID uint64 `json:"id"`
	jwt.RegisteredClaims
}

func (s *Server) access(userID uint64, username string, role int32) (string, error) {
	claims := jwt.NewWithClaims(signingMethod, accessClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.cfg.AccessExpireMinutes) * time.Minute)),
			ID:        s.newJTI(),
			Issuer:    s.cfg.Issuer,
		},
	})
	return claims.SignedString([]byte(s.cfg.AccessSecret))
}

func (s *Server) refresh(userID uint64) (string, error) {
	claims := jwt.NewWithClaims(signingMethod, refreshClaims{
		ID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.cfg.RefreshExpireHours) * time.Hour)),
			ID:        s.newJTI(),
			Issuer:    s.cfg.Issuer,
		},
	})
	return claims.SignedString([]byte(s.cfg.RefreshSecret))
}

// Issue 签发访问+刷新令牌。
func (s *Server) Issue(_ context.Context, req *authv1.IssueRequest) (*authv1.TokenPair, error) {
	access, err := s.access(req.GetUserId(), req.GetUsername(), req.GetRole())
	if err != nil {
		return nil, err
	}
	refresh, err := s.refresh(req.GetUserId())
	if err != nil {
		return nil, err
	}
	return &authv1.TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

// IssueAccess 仅签发访问令牌。
func (s *Server) IssueAccess(_ context.Context, req *authv1.IssueRequest) (*authv1.AccessReply, error) {
	access, err := s.access(req.GetUserId(), req.GetUsername(), req.GetRole())
	if err != nil {
		return nil, err
	}
	return &authv1.AccessReply{AccessToken: access}, nil
}

// IssueRefresh 仅签发刷新令牌。
func (s *Server) IssueRefresh(_ context.Context, req *authv1.RefreshRequest) (*authv1.RefreshReply, error) {
	refresh, err := s.refresh(req.GetUserId())
	if err != nil {
		return nil, err
	}
	return &authv1.RefreshReply{RefreshToken: refresh}, nil
}
