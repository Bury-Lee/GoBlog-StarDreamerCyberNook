// auth 服务:集中签发 JWT(密钥集中管理)。本地校验仍由各服务完成。
//
// 运行:
//
//	go run ./services/auth/cmd -config services/auth/auth.yaml
package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strconv"

	authv1 "StarDreamerCyberNook/gen/auth/v1"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/services/auth/internal"

	"gopkg.in/yaml.v2"
)

type fileConfig struct {
	Listen string `yaml:"listen"`
	Jwt    struct {
		AccessTokenSecret   string `yaml:"accessTokenSecret"`
		RefreshTokenSecret  string `yaml:"refreshTokenSecret"`
		AccessExpireMinutes int    `yaml:"accessExpireMinutes"`
		RefreshExpireHours  int    `yaml:"refreshExpireHours"`
		Issuer              string `yaml:"issuer"`
	} `yaml:"jwt"`
}

func main() {
	cfgPath := flag.String("config", "services/auth/auth.yaml", "配置文件路径")
	flag.Parse()

	data, err := os.ReadFile(*cfgPath)
	if err != nil {
		log.Fatalf("auth: read config %s: %v", *cfgPath, err)
	}
	var fc fileConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		log.Fatalf("auth: parse config %s: %v", *cfgPath, err)
	}

	listen := overrideStr("AUTH_LISTEN", fc.Listen)
	if listen == "" {
		listen = "127.0.0.1:9250"
	}
	cfg := internal.Config{
		AccessSecret:        overrideStr("JWT_ACCESS_SECRET", fc.Jwt.AccessTokenSecret),
		RefreshSecret:       overrideStr("JWT_REFRESH_SECRET", fc.Jwt.RefreshTokenSecret),
		AccessExpireMinutes: overrideInt("JWT_ACCESS_EXPIRE_MIN", fc.Jwt.AccessExpireMinutes),
		RefreshExpireHours:  overrideInt("JWT_REFRESH_EXPIRE_HOURS", fc.Jwt.RefreshExpireHours),
		Issuer:              overrideStr("JWT_ISSUER", fc.Jwt.Issuer),
	}
	if cfg.AccessSecret == "" || cfg.RefreshSecret == "" {
		log.Fatal("auth: jwt secrets must not be empty")
	}

	lis, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatalf("auth: listen %s: %v", listen, err)
	}

	srv := grpcx.NewServer()
	authv1.RegisterAuthServiceServer(srv, internal.NewServer(cfg))
	log.Printf("auth service listening on %s", listen)

	if err := srv.Serve(lis); err != nil {
		log.Fatalf("auth: serve: %v", err)
	}
}

func overrideStr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func overrideInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
