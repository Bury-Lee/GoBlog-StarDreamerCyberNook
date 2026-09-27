// community 服务:社区内容域(反馈 / 轮播图 / 友链 / 友情推广)。同库。
//
// 运行:
//
//	go run ./services/community/cmd -config services/community/community.yaml
package main

import (
	"flag"
	"log"
	"net"
	"os"

	communityv1 "StarDreamerCyberNook/gen/community/v1"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/services/community/internal"

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
	cfgPath := flag.String("config", "services/community/community.yaml", "配置文件路径")
	flag.Parse()

	data, err := os.ReadFile(*cfgPath)
	if err != nil {
		log.Fatalf("community: read config %s: %v", *cfgPath, err)
	}
	var fc fileConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		log.Fatalf("community: parse config %s: %v", *cfgPath, err)
	}

	listen := overrideStr("COMMUNITY_LISTEN", fc.Listen)
	if listen == "" {
		listen = "127.0.0.1:9290"
	}
	driver := overrideStr("COMMUNITY_DB_DRIVER", fc.DB.Driver)
	if driver == "" {
		driver = "mysql"
	}
	dsn := overrideStr("COMMUNITY_DB_DSN", fc.DB.DSN)
	if dsn == "" {
		log.Fatal("community: db.dsn (or COMMUNITY_DB_DSN) must not be empty")
	}

	srvImpl, err := internal.NewServer(internal.Config{Driver: driver, DSN: dsn, Standalone: true})
	if err != nil {
		log.Fatalf("community: init: %v", err)
	}

	lis, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatalf("community: listen %s: %v", listen, err)
	}

	s := grpcx.NewServer()
	communityv1.RegisterCommunityServiceServer(s, srvImpl)
	log.Printf("community service listening on %s (db=%s)", listen, driver)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("community: serve: %v", err)
	}
}

func overrideStr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
