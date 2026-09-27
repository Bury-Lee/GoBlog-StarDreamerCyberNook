// ai 服务:把模型调用从博客剥离为独立 gRPC 微服务。
//
// 运行:
//
//	go run ./services/ai/cmd -config services/ai/ai.yaml
//
// 配置以文件为主(见 ai.yaml);环境变量 AI_* 可选覆盖。
package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strconv"

	aiv1 "StarDreamerCyberNook/gen/ai/v1"
	"StarDreamerCyberNook/pkg/grpcx"
	"StarDreamerCyberNook/services/ai/internal"

	"gopkg.in/yaml.v2"
)

type fileConfig struct {
	Listen string `yaml:"listen"`
	AI     struct {
		Host        string  `yaml:"host"`
		APIKey      string  `yaml:"apiKey"`
		APIType     string  `yaml:"apiType"`
		Model       string  `yaml:"model"`
		Temperature float32 `yaml:"temperature"`
		MaxTokens   int     `yaml:"maxTokens"`
	} `yaml:"ai"`
}

func main() {
	cfgPath := flag.String("config", "services/ai/ai.yaml", "配置文件路径")
	flag.Parse()

	data, err := os.ReadFile(*cfgPath)
	if err != nil {
		log.Fatalf("ai: read config %s: %v", *cfgPath, err)
	}
	var fc fileConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		log.Fatalf("ai: parse config %s: %v", *cfgPath, err)
	}

	// 环境变量可选覆盖(密钥等敏感项建议走环境变量)
	listen := overrideStr("AI_LISTEN", fc.Listen)
	if listen == "" {
		listen = "127.0.0.1:9210"
	}
	cfg := internal.Config{
		Host:        overrideStr("AI_HOST", fc.AI.Host),
		APIKey:      overrideStr("AI_API_KEY", fc.AI.APIKey),
		APIType:     overrideStr("AI_APIType", fc.AI.APIType),
		Model:       overrideStr("AI_MODEL", fc.AI.Model),
		Temperature: float32(overrideFloat("AI_TEMPERATURE", float64(fc.AI.Temperature))),
		MaxTokens:   overrideInt("AI_MAX_TOKENS", fc.AI.MaxTokens),
	}

	lis, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatalf("ai: listen %s: %v", listen, err)
	}

	srv := grpcx.NewServer()
	aiv1.RegisterAIServiceServer(srv, internal.NewServer(cfg))
	log.Printf("ai service listening on %s (model=%s)", listen, cfg.Model)

	if err := srv.Serve(lis); err != nil {
		log.Fatalf("ai: serve: %v", err)
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

func overrideFloat(k string, def float64) float64 {
	if v := os.Getenv(k); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}
