// notify 服务:把邮件发送从博客剥离为独立 gRPC 微服务。
//
// 运行:
//
//	go run ./services/notify/cmd -config services/notify/notify.yaml
//
// 配置以文件为主(见 notify.yaml);环境变量 EMAIL_* / SITE_TITLE 可选覆盖。
package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strconv"

	notifyv1 "StarDreamerCyberNook/gen/notify/v1"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/services/notify/internal"

	"gopkg.in/yaml.v2"
)

type fileConfig struct {
	Listen    string `yaml:"listen"`
	SiteTitle string `yaml:"siteTitle"`
	Email     struct {
		Domain       string `yaml:"domain"`
		Port         int    `yaml:"port"`
		SendEmail    string `yaml:"sendEmail"`
		AuthCode     string `yaml:"authCode"`
		SendNickname string `yaml:"sendNickname"`
		SSL          bool   `yaml:"SSL"`
		TLS          bool   `yaml:"TLS"`
	} `yaml:"email"`
}

func main() {
	cfgPath := flag.String("config", "services/notify/notify.yaml", "配置文件路径")
	flag.Parse()

	data, err := os.ReadFile(*cfgPath)
	if err != nil {
		log.Fatalf("notify: read config %s: %v", *cfgPath, err)
	}
	var fc fileConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		log.Fatalf("notify: parse config %s: %v", *cfgPath, err)
	}

	listen := overrideStr("NOTIFY_LISTEN", fc.Listen)
	if listen == "" {
		listen = "127.0.0.1:9230"
	}
	cfg := internal.Config{
		Domain:       overrideStr("EMAIL_DOMAIN", fc.Email.Domain),
		Port:         overrideInt("EMAIL_PORT", fc.Email.Port),
		SendEmail:    overrideStr("EMAIL_SEND", fc.Email.SendEmail),
		AuthCode:     overrideStr("EMAIL_AUTHCODE", fc.Email.AuthCode),
		SendNickname: overrideStr("EMAIL_NICKNAME", fc.Email.SendNickname),
		SSL:          fc.Email.SSL,
		TLS:          fc.Email.TLS,
		SiteTitle:    overrideStr("SITE_TITLE", fc.SiteTitle),
	}

	lis, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatalf("notify: listen %s: %v", listen, err)
	}

	srv := grpcx.NewServer()
	notifyv1.RegisterNotifyServiceServer(srv, internal.NewServer(cfg))
	log.Printf("notify service listening on %s", listen)

	if err := srv.Serve(lis); err != nil {
		log.Fatalf("notify: serve: %v", err)
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
