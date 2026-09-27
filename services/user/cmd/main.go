// user 服务:用户关系域(关注/粉丝/好友)。同库。
//
// 运行:
//
//	go run ./services/user/cmd -config services/user/user.yaml
package main

import (
	"flag"
	"log"
	"net"
	"os"

	userv1 "StarDreamerCyberNook/gen/user/v1"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/services/user/internal"

	"gopkg.in/yaml.v2"
)

type fileConfig struct {
	Listen string `yaml:"listen"`
	DB     struct {
		Driver string `yaml:"driver"`
		DSN    string `yaml:"dsn"`
	} `yaml:"db"`
}

func main() {
	cfgPath := flag.String("config", "services/user/user.yaml", "配置文件路径")
	flag.Parse()

	data, err := os.ReadFile(*cfgPath)
	if err != nil {
		log.Fatalf("user: read config %s: %v", *cfgPath, err)
	}
	var fc fileConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		log.Fatalf("user: parse config %s: %v", *cfgPath, err)
	}

	listen := overrideStr("USER_LISTEN", fc.Listen)
	if listen == "" {
		listen = "127.0.0.1:9270"
	}
	driver := overrideStr("USER_DB_DRIVER", fc.DB.Driver)
	if driver == "" {
		driver = "mysql"
	}
	dsn := overrideStr("USER_DB_DSN", fc.DB.DSN)
	if dsn == "" {
		log.Fatal("user: db.dsn (or USER_DB_DSN) must not be empty")
	}

	srvImpl, err := internal.NewServer(internal.Config{Driver: driver, DSN: dsn, Standalone: true})
	if err != nil {
		log.Fatalf("user: init: %v", err)
	}

	lis, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatalf("user: listen %s: %v", listen, err)
	}

	s := grpcx.NewServer()
	userv1.RegisterUserServiceServer(s, srvImpl)
	log.Printf("user service listening on %s (db=%s)", listen, driver)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("user: serve: %v", err)
	}
}

func overrideStr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
