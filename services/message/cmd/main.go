// message 服务:站内信域(读接口 + 各业务关注通知)。同库。
//
// 运行:
//
//	go run ./services/message/cmd -config services/message/message.yaml
package main

import (
	"flag"
	"log"
	"net"
	"os"

	messagev1 "StarDreamerCyberNook/gen/message/v1"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/services/message/internal"

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
	cfgPath := flag.String("config", "services/message/message.yaml", "配置文件路径")
	flag.Parse()

	data, err := os.ReadFile(*cfgPath)
	if err != nil {
		log.Fatalf("message: read config %s: %v", *cfgPath, err)
	}
	var fc fileConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		log.Fatalf("message: parse config %s: %v", *cfgPath, err)
	}

	listen := overrideStr("MESSAGE_LISTEN", fc.Listen)
	if listen == "" {
		listen = "127.0.0.1:9280"
	}
	driver := overrideStr("MESSAGE_DB_DRIVER", fc.DB.Driver)
	if driver == "" {
		driver = "mysql"
	}
	dsn := overrideStr("MESSAGE_DB_DSN", fc.DB.DSN)
	if dsn == "" {
		log.Fatal("message: db.dsn (or MESSAGE_DB_DSN) must not be empty")
	}

	srvImpl, err := internal.NewServer(internal.Config{Driver: driver, DSN: dsn, Standalone: true})
	if err != nil {
		log.Fatalf("message: init: %v", err)
	}

	lis, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatalf("message: listen %s: %v", listen, err)
	}

	s := grpcx.NewServer()
	messagev1.RegisterMessageServiceServer(s, srvImpl)
	log.Printf("message service listening on %s (db=%s)", listen, driver)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("message: serve: %v", err)
	}
}

func overrideStr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
