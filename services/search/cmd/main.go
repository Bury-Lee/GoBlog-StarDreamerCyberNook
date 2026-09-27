// search 服务:把 Elasticsearch 从博客剥离为独立 gRPC 微服务。
//
// 运行:
//
//	go run ./services/search/cmd -config services/search/search.yaml
//
// 配置以文件为主(见 search.yaml);环境变量 ES_* 可选覆盖。
package main

import (
	"flag"
	"log"
	"net"
	"os"

	searchv1 "StarDreamerCyberNook/gen/search/v1"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/services/search/internal"

	"gopkg.in/yaml.v2"
)

type fileConfig struct {
	Listen string `yaml:"listen"`
	ES     struct {
		URL      string `yaml:"url"`
		UserName string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"es"`
}

func main() {
	cfgPath := flag.String("config", "services/search/search.yaml", "配置文件路径")
	flag.Parse()

	data, err := os.ReadFile(*cfgPath)
	if err != nil {
		log.Fatalf("search: read config %s: %v", *cfgPath, err)
	}
	var fc fileConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		log.Fatalf("search: parse config %s: %v", *cfgPath, err)
	}

	listen := overrideStr("SEARCH_LISTEN", fc.Listen)
	if listen == "" {
		listen = "127.0.0.1:9220"
	}
	esURL := overrideStr("ES_URL", fc.ES.URL)
	if esURL == "" {
		esURL = "http://127.0.0.1:9200"
	}

	srvImpl, err := internal.NewServer(internal.Config{
		URL:      esURL,
		UserName: overrideStr("ES_USERNAME", fc.ES.UserName),
		Password: overrideStr("ES_PASSWORD", fc.ES.Password),
	})
	if err != nil {
		log.Fatalf("search: connect es: %v", err)
	}

	lis, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatalf("search: listen %s: %v", listen, err)
	}

	s := grpcx.NewServer()
	searchv1.RegisterSearchServiceServer(s, srvImpl)
	log.Printf("search service listening on %s", listen)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("search: serve: %v", err)
	}
}

func overrideStr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
