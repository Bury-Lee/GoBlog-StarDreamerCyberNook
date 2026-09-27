package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"StarDreamerCyberNook/core"
	"StarDreamerCyberNook/flags"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/pkg/blog"
	"StarDreamerCyberNook/pkg/hostapp"

	"github.com/sirupsen/logrus"
)

// 博客默认开启;命令行参数作为"附加功能"由 flags.Run 执行,不决定平台是否运行。
// 是否以统一宿主模式运行由 setting.yaml 的 host.enable 决定,
// 启用哪些组件由 setting.yaml 的 components 段决定(无需独立蓝图文件)。
func main() {
	opt := flags.Parse(os.Args[1:])

	global.Config = core.ReadConf(opt.ConfigPath())

	// 需要数据库的命令行附加功能先初始化共享库(幂等),再由 flags.Run 统一执行
	if opt.DB || opt.Search || opt.Account != nil {
		if global.DB == nil {
			global.DB = core.InitDB()
		}
	}
	if err := flags.Run(opt); err != nil {
		logrus.Fatalf("执行命令行功能失败: %v", err)
	}

	if global.Config.Host.Enabled() {
		logrus.Debugf("当前配置: %+v", opt)
		//启动主程序调度微服务
		if err := hostapp.Run(context.Background(), opt.ConfigPath()); err != nil {
			logrus.Fatalf("运行失败: %v", err)
		}
		logrus.Info("宿主调度已停止")
		return
	}

	// 纯博客模式:仅启动博客 HTTP 服务(微服务需另行启动)
	app, err := blog.Run(context.Background(), opt.ConfigPath())
	if err != nil {
		logrus.Fatalf("初始化失败: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("正在关闭服务...")
	if err := app.Stop(context.Background()); err != nil {
		logrus.Errorf("服务关闭失败: %v", err)
	}
	logrus.Info("服务已关闭")
}
