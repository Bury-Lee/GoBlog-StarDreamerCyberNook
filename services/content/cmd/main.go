// content 服务:拥有文章内容域(逐步铺开)。同库(后续可独立库)。
//
// 运行:
//
//	go run ./services/content/cmd -config services/content/content.yaml
package main

import (
	"flag"
	"log"
	"net"
	"os"

	contentv1 "StarDreamerCyberNook/gen/content/v1"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/services/content/internal"

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
	cfgPath := flag.String("config", "services/content/content.yaml", "配置文件路径")
	flag.Parse()

	data, err := os.ReadFile(*cfgPath)
	if err != nil {
		log.Fatalf("content: read config %s: %v", *cfgPath, err)
	}
	var fc fileConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		log.Fatalf("content: parse config %s: %v", *cfgPath, err)
	}

	listen := overrideStr("CONTENT_LISTEN", fc.Listen)
	if listen == "" {
		listen = "127.0.0.1:9260"
	}
	driver := overrideStr("CONTENT_DB_DRIVER", fc.DB.Driver)
	if driver == "" {
		driver = "mysql"
	}
	dsn := overrideStr("CONTENT_DB_DSN", fc.DB.DSN)
	if dsn == "" {
		log.Fatal("content: db.dsn (or CONTENT_DB_DSN) must not be empty")
	}

	srvImpl, err := internal.NewServer(internal.Config{Driver: driver, DSN: dsn, Standalone: true})
	if err != nil {
		log.Fatalf("content: init: %v", err)
	}

	lis, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatalf("content: listen %s: %v", listen, err)
	}

	s := grpcx.NewServer()
	contentv1.RegisterContentServiceServer(s, srvImpl)
	log.Printf("content service listening on %s (db=%s)", listen, driver)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("content: serve: %v", err)
	}
}

func overrideStr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
